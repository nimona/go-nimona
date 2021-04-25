package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"nimona.io/pkg/blob"
	"nimona.io/pkg/config"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/daemon"
	"nimona.io/pkg/filesharing"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectstore"
)

const (
	OutgoingTransferRequestSent     = "TransferRequestSent"
	OutgoingTransferFileSent        = "OutgoingTransferFileSent"
	OutgoingTransferAccepted        = "OutgoingTransferAccepted"
	OutgoingTransferRejected        = "OutgoingTransferRejected"
	IncomingTransferRequestReceived = "IncomingTransferRequestReceived"
	IncomingTransferFileReceived    = "IncomingTransferFileReceived"
	IncomingTransferAccepted        = "IncomingTransferAccepted"
	IncomingTransferRejected        = "IncomingTransferRejected"
)

var (
	transferResponseType = new(filesharing.TransferResponse).Type()
	transferDoneType     = new(filesharing.TransferDone).Type()
)

type (
	Config struct {
		ReceivedFolder string `envconfig:"RECEIVED_FOLDER" default:"received_files"`
	}
	comboConf struct {
		hconf *Config
		nconf *config.Config
	}
	transferWrap struct {
		transfer *filesharing.Transfer
		status   string
		updated  time.Time
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
		transfers     map[string]*transferWrap
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

	d, err := daemon.New(
		ctx,
		daemon.WithConfigOptions(
			config.WithDefaultPath("~/.nimona-hermod"),
			config.WithExtraConfig("HERMOD", cfg),
		),
	)
	if err != nil {
		panic(err)
	}

	local := d.LocalPeer()
	man := d.ObjectManager()
	nnet := d.Network()
	res := d.Resolver()
	str := d.ObjectStore()
	nconf := d.Config()

	local.PutContentTypes(
		new(filesharing.File).Type(),
		new(blob.Blob).Type(),
		new(blob.Chunk).Type(),
	)

	cconf := &comboConf{
		hconf: cfg,
		nconf: &nconf,
	}

	// init textinput
	ti := textinput.NewModel()
	ti.Focus()

	fsh := filesharing.New(
		man,
		nnet,
		cfg.ReceivedFolder,
	)

	her.config = cconf
	her.local = local
	her.textInput = ti
	her.config = cconf
	her.resolver = res
	her.objectstore = str
	her.objectmanager = man
	her.fsh = fsh
	her.blobmanager = blob.NewManager(ctx, blob.WithObjectManager(man))
	her.transfers = make(map[string]*transferWrap)

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

	go func() {
		sub := nnet.Subscribe()
		for {
			env, err := sub.Next()
			if err != nil {
				fmt.Println("Failed to get object: ", err)
				return
			}
			switch env.Payload.Type {
			case transferDoneType:
				req := &filesharing.TransferDone{}

				if err := req.FromObject(env.Payload); err != nil {
					fmt.Println("Failed to get object: ", err)
					continue
				}

				her.transfers[req.Nonce].status = OutgoingTransferFileSent

			case transferResponseType:
				req := &filesharing.TransferResponse{}

				if err := req.FromObject(env.Payload); err != nil {
					fmt.Println("Failed to get object: ", err)
					continue
				}

				if !req.Accepted {
					her.transfers[req.Nonce].status = OutgoingTransferRejected
				}
				her.transfers[req.Nonce].status = OutgoingTransferAccepted

			}
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
	h.transfers[msg.nonce].status = IncomingTransferFileReceived
	return h, cmd
}

func (h *hermod) handleTransferMsg(
	msg transferMsg,
) (
	tea.Model,
	tea.Cmd,
) {
	var cmd tea.Cmd
	h.transfers[msg.trf.Request.Nonce] = &transferWrap{
		status:   IncomingTransferRequestReceived,
		transfer: msg.trf,
		updated:  time.Now(),
	}
	return h, cmd
}

func (h hermod) View() string {
	tpl := "%s\n%s\n"
	if len(h.transfers) > 0 {
		tpl += "Transfers:\n"

		transfers := []*transferWrap{}

		for _, trw := range h.transfers {
			transfers = append(transfers, trw)
		}

		sort.SliceStable(transfers, func(i, j int) bool {
			return transfers[i].updated.Unix() > transfers[j].updated.Unix()
		})

		for _, tr := range transfers {
			tpl += fmt.Sprintf(
				"-> Peer: %s File: %s ID: %s Status: %s\n", // TODO arrow based on direction?
				tr.transfer.Peer.String(),
				tr.transfer.Request.File.Name,
				tr.transfer.Request.Nonce,
				tr.status, // TODO convert status to string
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

	fc := strings.Split(fullCommand, " ")
	command := fc[0]
	params := []string{}

	for _, p := range fc[1:] {
		np := strings.Trim(p, " ")
		if np != "" {
			params = append(params, np)
		}
	}

	switch command {
	case "help":
		r := "* Send file: `send <file> <peer>`\n"
		r += "* Request offered file: `request <hash>`\n"
		r += "* Show local peer info: `local`\n"
		r += "* Quit: `quit`\n"
		h.result = r
	case "send":
		if len(params) < 2 {
			h.result = "usage: send <file> <peer>"
			break
		}
		h.result = fmt.Sprintf("%d", len(params))
		file := strings.Join(params[:len(params)-1], " ")
		recipient := crypto.PublicKey{}
		err := recipient.UnmarshalString(params[len(params)-1])
		if err != nil {
			h.result = "invalid recipient key"
		}
		h.result = fmt.Sprintf(
			"Sending file %s to %s ...",
			file,
			recipient.String(),
		)
		h.sendFile(file, recipient)
	case "local":
		h.result = fmt.Sprintf(
			"public_key: %s\naddresses: %s\n",
			h.local.ConnectionInfo().PublicKey,
			h.local.ConnectionInfo().Addresses,
		)
	case "request":
		if len(params) != 1 {
			h.result = "usage: request <hash>"
			break
		}
		h.result = fmt.Sprintf(
			"Requesting transfer: %s ...",
			params,
		)
		h.requestFile(params[0])
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

	fr := &filesharing.File{
		Name:   filename,
		Chunks: bl.Chunks,
	}
	nonce, err := h.fsh.RequestTransfer(
		ctx,
		fr,
		peerKey,
	)
	h.transfers[nonce] = &transferWrap{
		status: OutgoingTransferRequestSent,
		transfer: &filesharing.Transfer{
			Request: filesharing.TransferRequest{
				Nonce: nonce,
				File:  *fr,
			},
			Peer: peerKey,
		},
		updated: time.Now(),
	}
	if err != nil {
		h.result = err.Error()
		return
	}
}

func (h *hermod) requestFile(
	nonce string,
) {
	trf := h.transfers[nonce]
	_, err := h.fsh.RequestFile(context.Background(), trf.transfer)
	if err != nil {
		h.result = err.Error()
		return
	}
	h.Update(fileReceivedMsg{
		nonce: trf.transfer.Request.Nonce,
	})
}
