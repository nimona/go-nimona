package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type (
	App struct {
		Windows    Windows
		StatusText []Status
		Store      *store
		Channels   Channels
		Chat       *chat
	}
	Channels struct {
		InputLines          chan string
		SelfNicknameUpdated chan string
		MessageAdded        chan *Message
		ParticipantUpdated  chan *Participant
	}
	Windows struct {
		Participants *tview.List
		Input        *tview.InputField
		Chat         *tview.TextView
		App          *tview.Application
	}
	Status struct {
		Text    string
		Created string
	}
)

func NewApp(conversationHash string) *App {
	app := &App{
		Windows: Windows{
			Participants: tview.NewList(),
			Input:        tview.NewInputField(),
			Chat:         tview.NewTextView(),
			App:          tview.NewApplication(),
		},
		Store:      NewMemoryStore(),
		StatusText: []Status{},
		Channels: Channels{
			InputLines:          make(chan string, 100),
			SelfNicknameUpdated: make(chan string, 100),
			MessageAdded:        make(chan *Message, 100),
			ParticipantUpdated:  make(chan *Participant, 100),
		},
	}

	app.Store.PutConversation(&Conversation{
		Hash: conversationHash,
	})

	go func() {
		convs, _ := app.Store.GetConversations()
		conv := convs[0]

		formatParticipant := func(p *Participant) string {
			key := strings.Replace(p.Key, "ed25519.", "", 1)
			if p.Nickname == "" {
				return fmt.Sprintf("<%s>", first(key, 8))
			}
			nickname := p.Nickname
			if len(nickname) > 12 {
				nickname = first(nickname, 11) + "…"
			}
			return fmt.Sprintf("%s <%s>", nickname, first(key, 8))
		}

		participantsViewRefresh := func() {
			convParticipants, _ := app.Store.GetParticipants(conv.Hash)
			app.Windows.Participants.Clear()
			for _, p := range convParticipants {
				nickname := formatParticipant(p)
				app.Windows.Participants.AddItem(nickname, "", 0, func() {})
			}
		}

		messagesViewRefresh := func() {
			convMessages, _ := app.Store.GetMessages(conv.Hash, 100, 0)
			convParticipants, _ := app.Store.GetParticipants(conv.Hash)
			app.Windows.Chat.Clear()
			for _, message := range convMessages {
				if message.SenderKey == "system" {
					app.Windows.Chat.Write([]byte(
						fmt.Sprintf(
							"\n[red][%s] %s",
							message.Created.Format("02/01 15:04:05"),
							message.Body,
						),
					))
					continue
				}
				nickname := ""
				for _, p := range convParticipants {
					if message.SenderKey == p.Key {
						nickname = formatParticipant(p)
						break
					}
				}
				app.Windows.Chat.Write([]byte(
					fmt.Sprintf(
						"\n[lightcyan][%s][gold] %s[white] %s",
						message.Created.Format("02/01 15:04:05"),
						nickname,
						message.Body,
					),
				))
			}
			app.Windows.Chat.ScrollToEnd()
			app.Windows.App.Draw()
		}

		for {
			select {
			case participantUpdated := <-app.Channels.ParticipantUpdated:
				app.Store.PutMessage(&Message{
					Hash: strconv.Itoa(
						int(time.Now().UnixNano()),
					),
					ConversationHash: participantUpdated.ConversationHash,
					SenderKey:        "system",
					Created:          participantUpdated.Updated,
					Body: fmt.Sprintf(
						"* <%s> is now known as %s",
						participantUpdated.Key,
						participantUpdated.Nickname,
					),
				})
				app.Store.PutParticipant(participantUpdated)
				participantsViewRefresh()
				messagesViewRefresh()

			case messageAdded := <-app.Channels.MessageAdded:
				app.Store.PutMessage(messageAdded)
				messagesViewRefresh()

				// deal with users
				if messageAdded.SenderKey == "system" {
					continue
				}
				par := &Participant{
					Key:              messageAdded.SenderKey,
					ConversationHash: messageAdded.ConversationHash,
				}
				app.Store.PutParticipant(par)
				participantsViewRefresh()
			}
		}
	}()

	return app
}

func (app *App) AddSystemText(conversationHash string, msg string) {
	app.Channels.MessageAdded <- &Message{
		Hash:             strconv.Itoa(int(time.Now().UnixNano())),
		ConversationHash: conversationHash,
		Created:          time.Now(),
		SenderKey:        "system",
		Body:             msg,
	}
}

func (app *App) Quit() {
	app.Windows.App.Stop()
	os.Exit(0)
}

func (app *App) Show() {
	log.SetOutput(ioutil.Discard)

	app.Windows.Participants.SetBorder(true).SetTitle(" Participants ")
	app.Windows.Participants.ShowSecondaryText(false)
	app.Windows.Participants.SetSelectedBackgroundColor(tcell.ColorBlack)
	app.Windows.Participants.SetSelectedTextColor(tcell.ColorWhite)
	app.Windows.Participants.SetBorderPadding(0, 0, 1, 1)

	app.Windows.Chat.SetBorder(true)
	app.Windows.Chat.SetTitleAlign(tview.AlignLeft)
	app.Windows.Chat.SetDynamicColors(true)
	app.Windows.Chat.SetScrollable(true)
	app.Windows.Chat.SetWordWrap(true)
	app.Windows.Chat.SetBorderPadding(0, 0, 0, 0)

	app.Windows.Input.SetFieldBackgroundColor(tcell.ColorBlack)
	app.Windows.Input.SetBorder(true)
	app.Windows.Input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {

			convs, _ := app.Store.GetConversations()
			conv := convs[0]

			if app.Windows.Input.GetText() == "" {
				return
			}

			text := app.Windows.Input.GetText()

			if text[0] == '/' {
				words := strings.Fields(text[1:])
				if len(words) == 0 {
					return
				}
				switch words[0] {
				case "nick":
					nickname := strings.TrimPrefix(text, "/nick ")
					app.Channels.SelfNicknameUpdated <- nickname
				case "quit", "q":
					app.Quit()
					os.Exit(0)
				case "info", "i":
					app.AddSystemText(
						conv.Hash,
						"Info:",
					)
					app.AddSystemText(
						conv.Hash,
						fmt.Sprintf(
							"* public key: %s",
							app.Chat.mesh.GetPeerKey().PublicKey(),
						),
					)
					app.AddSystemText(
						conv.Hash,
						fmt.Sprintf(
							"* addresses: %s",
							app.Chat.mesh.GetAddresses(),
						),
					)
					// TODO handle error
					// nolint: errcheck
					pinnedHashes, _ := app.Chat.objectstore.GetPinned()
					app.AddSystemText(
						conv.Hash,
						fmt.Sprintf(
							"* pinned hashes: %v",
							pinnedHashes,
						),
					)
				default:
					app.AddSystemText(
						conv.Hash,
						fmt.Sprintf("[red]No such command '%s'.", text[1:]),
					)
				}
				app.Windows.Input.SetText("")
				return
			}

			app.Channels.InputLines <- text
			app.Windows.Input.SetText("")
		}
	})
	app.Windows.Input.SetInputCapture(app.Windows.Input.GetInputCapture())

	// app.Windows.Chat.SetTitle(" " + app.Conversations[0].Hash + " ")

	flexLists := tview.NewFlex().SetDirection(tview.FlexRow)
	flexChat := tview.NewFlex().SetDirection(tview.FlexRow)

	flex := tview.NewFlex()
	flexLists.AddItem(app.Windows.Participants, 0, 4, false)
	flexChat.AddItem(app.Windows.Chat, 0, 10, false)
	flexChat.AddItem(app.Windows.Input, 3, 2, false)
	flex.AddItem(flexLists, 0, 1, false)
	flex.AddItem(flexChat, 0, 5, false)

	if err := app.Windows.App.SetRoot(flex, true).SetFocus(app.Windows.Input).Run(); err != nil {
		panic(err)
	}
}
