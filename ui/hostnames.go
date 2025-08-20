package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type HostnamesView struct {
	UI   *UI
	List *tview.List
	Page *tview.Flex
}

func NewHostnamesView(ui *UI) *HostnamesView {
	list := tview.NewList().ShowSecondaryText(false)

	view := HostnamesView{
		UI:   ui,
		List: list,
		Page: tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewTextView().SetText("Select a domain to view requests. Press '[' to quit."), 2, 0, false).
			AddItem(list, 0, 1, true),
	}

	list.SetSelectedFunc(func(index int, text string, hostname string, _ rune) {
		ui.SelectedHostname = hostname
		ui.RequestsView.Header.SetText(fmt.Sprintf("%s (Press 'q' to go back to domain list)", hostname))

		ui.RequestsView.List.Clear()

		for _, event := range ui.Events {
			if event.EventType == "request" {
				req, _ := ui.Requests.Get(event.EventID)

				if req.Hostname == hostname {
					ui.RequestsView.List.AddItem(craftRequestLine(req), req.ID, 0, nil)
				}
			} else if event.EventType == "websocket" {
				ws, _ := ui.Websockets.Get(event.EventID)

				url := ws.URL
				if len(url) > 50 {
					url = url[50:]
				}

				if ws.Hostname == hostname {
					ui.RequestsView.List.AddItem(fmt.Sprintf("[WEBSOCKET CONN] %s", formatWebsocketURL(ws.TLS, url)), ws.ID, 0, nil)
				}
			}
		}

		ui.App.SetRoot(ui.RequestsView.Page, true).SetFocus(ui.RequestsView.List)
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == '[' {
			ui.App.Stop()
			return nil
		}

		return event
	})

	return &view
}
