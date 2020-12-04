package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"nimona.io/internal/rand"
)

type (
	App struct {
		Windows       Windows
		StatusText    []Status
		Conversations Conversations
		Channels      Channels
		Chat          *chat
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
	Conversations []*Conversation // helper, used for sorting
	Conversation  struct {
		Hash         string
		Messages     Messages
		Participants Participants
		LastActivity time.Time
	}
	Messages []*Message // helper, used for sorting
	Message  struct {
		Hash             string
		ConversationHash string
		Body             string
		SenderHash       string
		Created          time.Time
	}
	Participants []*Participant // helper, used for sorting
	Participant  struct {
		Hash     string
		Nickname string
	}
)

func (a Messages) Len() int {
	return len(a)
}

func (a Messages) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a Messages) Less(i, j int) bool {
	return a[i].Created.Before(a[j].Created)
}

func (a Participants) Len() int {
	return len(a)
}

func (a Participants) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a Participants) Less(i, j int) bool {
	return a[i].Nickname < a[j].Nickname
}

func NewApp(conversationHash string) *App {
	app := &App{
		Windows: Windows{
			Participants: tview.NewList(),
			Input:        tview.NewInputField(),
			Chat:         tview.NewTextView(),
			App:          tview.NewApplication(),
		},
		StatusText: []Status{},
		Channels: Channels{
			InputLines:          make(chan string, 100),
			SelfNicknameUpdated: make(chan string, 100),
			MessageAdded:        make(chan *Message, 100),
			ParticipantUpdated:  make(chan *Participant, 100),
		},
		Conversations: Conversations{
			&Conversation{
				Hash:     conversationHash,
				Messages: Messages{},
			},
		},
	}

	go func() {
		conv := app.Conversations[0]

		formatParticipant := func(p *Participant) string {
			if p.Nickname == last(p.Hash, 8) {
				return fmt.Sprintf("<%s>", p.Nickname)
			}
			return fmt.Sprintf("%s <%s>", p.Nickname, last(p.Hash, 8))
		}

		participantsViewRefresh := func() {
			app.Windows.Participants.Clear()
			for _, p := range conv.Participants {
				nickname := formatParticipant(p)
				app.Windows.Participants.AddItem(nickname, "", 0, func() {})
			}
		}

		messagesViewRefresh := func() {
			app.Windows.Chat.Clear()
			min := 0
			if len(conv.Messages) > 100 {
				min = len(conv.Messages) - 100
			}
			for _, message := range conv.Messages[min:] {
				if message.SenderHash == "system" {
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
				for _, p := range conv.Participants {
					if message.SenderHash == p.Hash {
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
				for _, u := range conv.Participants {
					if u.Hash == participantUpdated.Hash {
						u.Nickname = participantUpdated.Nickname
						break
					}
				}
				participantsViewRefresh()
				messagesViewRefresh()

			case messageAdded := <-app.Channels.MessageAdded:
				duplicate := false
				for _, message := range conv.Messages {
					if message.Hash == messageAdded.Hash {
						duplicate = true
						break
					}
				}
				if duplicate {
					break
				}
				conv.Messages = append(
					conv.Messages,
					messageAdded,
				)
				sort.Sort(conv.Messages)
				messagesViewRefresh()

				// deal with users
				if messageAdded.SenderHash == "system" {
					continue
				}
				userExists := false
				for _, u := range conv.Participants {
					if u.Hash == messageAdded.SenderHash {
						userExists = true
						break
					}
				}
				if !userExists {
					conv.Participants = append(
						conv.Participants,
						&Participant{
							Hash:     messageAdded.SenderHash,
							Nickname: last(messageAdded.SenderHash, 8),
						},
					)
				}
				sort.Sort(conv.Participants)
				participantsViewRefresh()
			}
		}
	}()

	return app
}

func (app *App) AddSystemText(msg string) {
	app.Channels.MessageAdded <- &Message{
		Hash:       rand.String(16),
		SenderHash: "system",
		Body:       msg,
		Created:    time.Now(),
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
					app.AddSystemText("Info:")
					app.AddSystemText(
						fmt.Sprintf(
							"* public key: %s",
							app.Chat.local.GetPrimaryPeerKey().PublicKey(),
						),
					)
					app.AddSystemText(
						fmt.Sprintf(
							"* addresses: %s",
							app.Chat.local.GetAddresses(),
						),
					)
					app.AddSystemText(
						fmt.Sprintf(
							"* pinned hashes: %v",
							app.Chat.local.GetContentHashes(),
						),
					)
				default:
					app.AddSystemText(fmt.Sprintf("[red]No such command '%s'.", text[1:]))
				}
				app.Windows.Input.SetText("")
				return
			}

			app.Channels.InputLines <- text
			app.Windows.Input.SetText("")
		}
	})
	app.Windows.Input.SetInputCapture(app.Windows.Input.GetInputCapture())

	app.Windows.Chat.SetTitle(" " + app.Conversations[0].Hash + " ")

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
