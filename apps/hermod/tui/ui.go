package tui

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"nimona.io/internal/net"
	"nimona.io/pkg/blob"
	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/filesharing"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/network"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
)

type (
	Config struct {
		ReceivedFolder string `envconfig:"RECEIVED_FOLDER" default:"received_files"`
	}
	comboConf struct {
		hconf *Config
		nconf *config.Config
	}
	hermod struct {
		textInput textinput.Model
		result    string

		config *comboConf

		local         localpeer.LocalPeer
		objectmanager objectmanager.ObjectManager
		blobmanager   blob.Manager
		objectstore   objectstore.Store
		resolver      resolver.Resolver
		fsh           filesharing.Filesharer
		listener      net.Listener

		incomingTransfers map[string]*filesharing.Transfer
		outgoingTransfers map[string]*filesharing.Transfer
		receivedTransfers map[string]*filesharing.Transfer
	}
	transferMsg struct {
		trf *filesharing.Transfer
	}
	fileReceivedMsg struct {
		nonce string
	}
)

func NewHermod() hermod {
	her := &hermod{}

	ctx := context.New(
		context.WithCorrelationID("nimona"),
	)

	// init config
	cfg := &Config{}
	ncfg, err := config.New(
		config.WithExtraConfig("HERMOD", cfg),
	)
	cconf := &comboConf{
		hconf: cfg,
		nconf: ncfg,
	}
	if err != nil {
		fmt.Println("Failed to parse config: ", err)
		os.Exit(-1)
	}

	// construct local peer
	local := localpeer.New()
	// attach peer private key from config
	local.PutPrimaryPeerKey(cconf.nconf.Peer.PrivateKey)
	local.PutContentTypes(
		new(filesharing.File).Type(),
		new(blob.Blob).Type(),
		new(blob.Chunk).Type(),
	)

	// construct new network
	nnet := network.New(
		ctx,
		network.WithLocalPeer(local),
	)

	// make sure we have some bootstrap peers to start with
	if len(cconf.nconf.Peer.Bootstraps) == 0 {
		cconf.nconf.Peer.Bootstraps = []peer.Shorthand{
			"bahwqcabae4kl233toxg4qtvual2pcwylp32ht5b4xkmbjwuqkgtweizczltq@tcps:asimov.bootstrap.nimona.io:22581",
			"bahwqcabarcrxtiaha3uq25gvntnqb6uokgdp442dysocya42ckiugohxmqkq@tcps:egan.bootstrap.nimona.io:22581",
			"bahwqcabafguo2axx2ydpk5mrjlrsjw2rjwo34uzzr6kvtfb6cevx72q5t4bq@tcps:sloan.bootstrap.nimona.io:22581",
		}
	}

	// convert shorthands into peers
	bootstrapPeers := []*peer.ConnectionInfo{}
	for _, s := range cconf.nconf.Peer.Bootstraps {
		bootstrapPeer, err := s.ConnectionInfo()
		if err != nil {
			fmt.Println("error parsing bootstrap peer:", err)
			os.Exit(-1)
		}
		bootstrapPeers = append(bootstrapPeers, bootstrapPeer)
	}

	// add bootstrap peers as relays
	local.PutRelays(bootstrapPeers...)

	// construct new resolver
	res := resolver.New(
		ctx,
		nnet,
		resolver.WithBoostrapPeers(bootstrapPeers...),
	)

	// construct object store
	db, err := sql.Open("sqlite3", "file_transfer.db")
	if err != nil {
		fmt.Println("error opening sql file", err)
		os.Exit(-1)
	}

	str, err := sqlobjectstore.New(db)
	if err != nil {
		fmt.Println("error starting sql store", err)
		os.Exit(-1)
	}

	// construct object manager
	man := objectmanager.New(
		ctx,
		nnet,
		res,
		str,
	)

	// init textinput
	ti := textinput.NewModel()
	ti.Focus()

	fsh := filesharing.New(
		man,
		nnet,
		cfg.ReceivedFolder,
	)

	// start listening
	lis, err := nnet.Listen(
		ctx,
		cconf.nconf.Peer.BindAddress,
		network.ListenOnLocalIPs,
		network.ListenOnPrivateIPs,
	)
	if err != nil {
		fmt.Println("error while listening", err)
		os.Exit(-1)
	}

	her.config = cconf
	her.local = local
	her.textInput = ti
	her.config = cconf
	her.resolver = res
	her.objectstore = str
	her.listener = lis
	her.objectmanager = man
	her.fsh = fsh
	her.blobmanager = blob.NewManager(ctx, blob.WithObjectManager(man))
	her.incomingTransfers = make(map[string]*filesharing.Transfer)
	her.receivedTransfers = make(map[string]*filesharing.Transfer)

	go func() {
		transfers, err := her.fsh.Listen(ctx)
		if err != nil {
			fmt.Println("failed to listen: ", err)
			os.Exit(-1)
		}

		for transfer := range transfers {
			her.Update(transferMsg{
				trf: transfer,
			})
		}
	}()
	return *her
}

func (h hermod) Init() tea.Cmd {
	return textinput.Blink
}

func (h hermod) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return h, tea.Quit
		case "enter":
			return h.execute()
		}
	case transferMsg:
		return h.handleTransferMsg(msg)
	case fileReceivedMsg:
		return h.handleFileReceivedMsg(msg)
	}

	h.textInput, cmd = h.textInput.Update(msg)

	return h, cmd
}

func (h *hermod) handleFileReceivedMsg(
	msg fileReceivedMsg,
) (
	tea.Model, tea.Cmd,
) {
	var cmd tea.Cmd
	trf := h.incomingTransfers[msg.nonce]
	h.receivedTransfers[msg.nonce] = trf
	delete(h.incomingTransfers, msg.nonce)
	return h, cmd
}

func (h *hermod) handleTransferMsg(
	msg transferMsg,
) (
	tea.Model,
	tea.Cmd,
) {
	var cmd tea.Cmd
	h.incomingTransfers[msg.trf.Request.Nonce] = msg.trf
	return h, cmd
}

func (h hermod) View() string {
	tpl := "%s\n%s\n"
	if len(h.incomingTransfers) > 0 {
		tpl += "TransferRequests\n"
		for _, trf := range h.incomingTransfers {
			tpl += fmt.Sprintf(
				"Peer: %s requested to send file: %s id: %s\n",
				trf.Peer.String(),
				trf.Request.File.Name,
				trf.Request.Nonce,
			)
		}
	}
	if len(h.receivedTransfers) > 0 {
		tpl += "Received Files\n"
		for _, trf := range h.receivedTransfers {
			tpl += fmt.Sprintf(
				"Peer: %s sent file: %s\n",
				trf.Peer.String(),
				trf.Request.File.Name,
			)
		}
	}
	v := fmt.Sprintf(
		tpl,
		h.textInput.View(),
		h.result,
	)
	return v
}

func (h *hermod) execute() (tea.Model, tea.Cmd) {
	h.textInput.Blur()

	fullCommand := h.textInput.Value()

	commands := strings.Split(fullCommand, " ")

	switch commands[0] {
	case "send":
		if len(commands) != 3 {
			h.result = "invalid command"
		}
		h.result = fmt.Sprintf("Sending file %s to %s ...", commands[1], commands[2])
		h.sendFile(commands[1], crypto.PublicKey(commands[2]))
	case "list":
		h.result = "Listing local files..."
	case "local":
		h.result = fmt.Sprintf(
			"public_key: %s\naddresses: %s\n",
			h.local.ConnectionInfo().PublicKey,
			h.local.ConnectionInfo().Addresses,
		)
	case "request":
		h.result = fmt.Sprintf(
			"Requesting transfer: %s ...",
			commands[1],
		)
		h.requestFile(commands[1])
	case "quit":
		return h, tea.Quit
	default:
		h.result = ""
	}

	h.textInput.Reset()
	h.textInput.Focus()

	return h, nil
}

func (h *hermod) sendFile(
	file string,
	peerKey crypto.PublicKey,
) {
	ctx := context.Background()
	filename := filepath.Base(file)
	bl, err := h.blobmanager.ImportFromFile(ctx, file)
	if err != nil {
		h.result = err.Error()
		return
	}
	h.fsh.RequestTransfer(
		ctx,
		&filesharing.File{
			Name:   filename,
			Chunks: bl.Chunks,
		},
		peerKey,
	)
}

func (h *hermod) requestFile(
	nonce string,
) {
	trf := h.incomingTransfers[nonce]
	h.result = fmt.Sprintf("%v", trf)
	_, err := h.fsh.RequestFile(context.Background(), trf)
	if err != nil {
		h.result = err.Error()
		return
	}
	h.Update(fileReceivedMsg{
		nonce: trf.Request.Nonce,
	})
}
