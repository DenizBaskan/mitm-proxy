package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type RequestsView struct {
	UI     *UI
	Header *tview.TextView
	List   *tview.List
	Page   *tview.Flex
}

func NewRequestsView(ui *UI) *RequestsView {
	list := tview.NewList().ShowSecondaryText(false)
	header := tview.NewTextView()
	page := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(header, 2, 0, false).AddItem(list, 0, 1, true)

	view := &RequestsView{
		UI:     ui,
		List:   list,
		Header: header,
		Page:   page,
	}

	list.SetSelectedFunc(func(index int, text string, identifier string, shortcut rune) {
		if strings.HasPrefix(text, "[WEBSOCKET") {
			websocketID := identifier
			ui.SelectedWebsocketID = websocketID

			websocket, _ := ui.Websockets.Get(websocketID)
			ui.WebsocketView.Header.SetText(formatWebsocketURL(websocket.TLS, websocket.URL))

			ui.WebsocketView.List.Clear()
			msgs, _ := ui.WebsocketMessages.Get(websocketID)

			for _, msg := range msgs {
				ui.WebsocketView.List.AddItem(craftWebsocketMessage(msg), string(msg.Message), 0, nil)
			}

			ui.App.SetRoot(ui.WebsocketView.Page, true).SetFocus(ui.WebsocketView.Header)

		} else {
			id := identifier
			ui.SelectedRequestID = id

			req, _ := ui.Requests.Get(id)
			reqString, resString := craftInspect(req)

			ui.InspectView.Header.SetText(req.Req.Method + " " + req.URL)
			ui.InspectView.Request.SetText(reqString)
			ui.InspectView.Response.SetText(resString)

			ui.App.SetRoot(ui.InspectView.Page, true).SetFocus(ui.InspectView.Header)
		}
	})

	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			ui.App.SetRoot(ui.HostnamesView.Page, true).SetFocus(ui.HostnamesView.List)
			return nil
		}

		return event
	})

	return view
}
