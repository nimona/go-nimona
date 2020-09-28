package main

import (
	"container/list"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type (
	App struct {
		Windows          Windows
		StatusText       []Status
		InputHistory     *list.List
		InputHistoryHead *list.Element
		InputPos         *list.Element
		InputLines       chan string
		Chat             *chat
	}
	Windows struct {
		Users *tview.List
		Input *tview.InputField
		Chat  *tview.TextView
		App   *tview.Application
	}
	Status struct {
		Text    string
		Created string
	}
)

func NewApp() *App {
	inputHistory := list.New()
	return &App{
		Windows: Windows{
			Users: tview.NewList(),
			Input: tview.NewInputField(),
			Chat:  tview.NewTextView(),
			App:   tview.NewApplication(),
		},
		StatusText:       []Status{},
		InputHistory:     inputHistory,
		InputHistoryHead: inputHistory.PushFront(""),
		InputPos:         inputHistory.Front(),
		InputLines:       make(chan string, 100),
	}
}

func (app *App) AddStatusText(msg string) {
	app.Windows.Chat.Write([]byte(fmt.Sprintf("[red]%s %s\n", time.Now().Format("02/01 15:04:05"), msg)))
	app.Windows.Chat.ScrollToEnd()
	app.StatusText = append(app.StatusText, Status{Text: msg, Created: time.Now().Format("02/01 15:04:05")})

}

func (app *App) AddUserText(msg, usr string, t time.Time) {
	app.Windows.Chat.Write([]byte(fmt.Sprintf("[cyan][%s][yellow] <%s>[white] %s\n", t.Format("02/01 15:04:05"), usr, msg)))
	app.Windows.Chat.ScrollToEnd()
	app.Windows.App.Draw()
}

func (app *App) ResetInputHistoryPosition() {
	app.InputPos = app.InputHistory.Front()
}

func (app *App) Quit() {
	app.Windows.App.Stop()
	os.Exit(0)
}

func (app *App) Show() {
	log.SetOutput(ioutil.Discard)

	app.Windows.Users.SetBorder(true).SetTitle("Users")
	app.Windows.Users.SetTitleColor(tcell.ColorGreen)
	app.Windows.Users.ShowSecondaryText(false)
	app.Windows.Users.SetSelectedBackgroundColor(tcell.ColorBlack)
	app.Windows.Users.SetSelectedTextColor(tcell.ColorNavy)

	app.Windows.Chat.SetBorder(true)
	app.Windows.Chat.SetTitleAlign(tview.AlignLeft)
	app.Windows.Chat.SetDynamicColors(true)
	app.Windows.Chat.SetScrollable(true)
	app.Windows.Chat.SetWordWrap(true)

	app.Windows.Input.SetFieldBackgroundColor(tcell.ColorBlack)
	app.Windows.Input.SetBorder(true)
	app.Windows.Input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			if app.Windows.Input.GetText() == "" {
				return
			}

			text := app.Windows.Input.GetText()
			// AddInputHistory(text)

			if text[0] == '/' {
				words := strings.Fields(text[1:])
				if len(words) == 0 {
					return
				}
				switch words[0] {
				case "quit", "q":
					app.Quit()
				case "whoami", "i":
					app.AddStatusText(
						fmt.Sprintf(
							"> public key: %s",
							app.Chat.local.GetPrimaryPeerKey().PublicKey(),
						),
					)
					app.AddStatusText(
						fmt.Sprintf(
							"> addresses: %s",
							app.Chat.local.GetAddresses(),
						),
					)
				default:
					app.AddStatusText(fmt.Sprintf("[red]No such command '%s'.", text[1:]))
				}
				app.Windows.Input.SetText("")
				app.ResetInputHistoryPosition()
				return
			}

			// if user.ActiveSpaceId != "status" {
			// go SendMessageToChannel(text)
			// AddOwnText(text, user.Info.DisplayName, "")
			// own = append(own, OwnMessages{SpaceId: user.ActiveSpaceId, Text: text})
			app.InputLines <- text
			app.Windows.Input.SetText("")
			app.ResetInputHistoryPosition()
			// }
		}
	})
	app.Windows.Input.SetInputCapture(app.Windows.Input.GetInputCapture())

	app.Windows.Chat.SetTitle("foo")

	flexLists := tview.NewFlex().SetDirection(tview.FlexRow)
	flexChat := tview.NewFlex().SetDirection(tview.FlexRow)

	flex := tview.NewFlex()
	flexLists.AddItem(app.Windows.Users, 0, 4, false)
	flexChat.AddItem(app.Windows.Chat, 0, 10, false)
	flexChat.AddItem(app.Windows.Input, 3, 2, false)
	flex.AddItem(flexLists, 0, 1, false)
	flex.AddItem(flexChat, 0, 5, false)

	// AddStatusText(`[#A60000] ███╗   ██╗██╗███╗   ███╗ ██████╗ ███╗   ██╗ █████╗  `)
	// AddStatusText(`[#C40000] ████╗  ██║██║████╗ ████║██╔═══██╗████╗  ██║██╔══██╗ `)
	// AddStatusText(`[#EC1E0D] ██╔██╗ ██║██║██╔████╔██║██║   ██║██╔██╗ ██║███████║ `)
	// AddStatusText(`[#F54E16] ██║╚██╗██║██║██║╚██╔╝██║██║   ██║██║╚██╗██║██╔══██║ `)
	// AddStatusText(`[#F57316] ██║ ╚████║██║██║ ╚═╝ ██║╚██████╔╝██║ ╚████║██║  ██║ `)
	// AddStatusText(`[#F5E216] ╚═╝  ╚═══╝╚═╝╚═╝     ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚═╝  ╚═╝ `)

	// AddStatusText(`[#A60000]         _                                    `)
	// AddStatusText(`[#C40000]        (_)                                   `)
	// AddStatusText(`[#EC1E0D]   _ __  _ _ __ ___   ___  _ __   __ _        `)
	// AddStatusText(`[#F54E16]  | '_ \| | '_ ' _ \ / _ \| '_ \ / _' |       `)
	// AddStatusText(`[#F57316]  | | | | | | | | | | (_) | | | | (_| |       `)
	// AddStatusText(`[#F5E216]  |_| |_|_|_| |_| |_|\___/|_| |_|\__,_|       `)
	// AddStatusText(`                                                       `)

	// AddStatusText("[#A60000]                     ")
	// AddStatusText("[#C40000]    .                ")
	// AddStatusText("[#EC1E0D] .-...-.-..-..-..-.  ")
	// AddStatusText("[#F54E16] ' ''' ' '`-'' '`-`- ")
	// AddStatusText("[#F57316]                     ")

	// AddStatusText(fmt.Sprintf("Theme used: %s", config.ThemeFile))
	// AddStatusText(fmt.Sprintf("Webhook url used: %s", user.GrokUrl))

	if err := app.Windows.App.SetRoot(flex, true).SetFocus(app.Windows.Input).Run(); err != nil {
		panic(err)
	}

	// End all workers
	// for i := 0; i < channels.workers; i++ {
	// 	channels.Quit <- 1
	// }
}
