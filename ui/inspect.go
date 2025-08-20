package ui

import (
	"bytes"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/gdamore/tcell/v2"
	"github.com/moul/http2curl"
	"github.com/rivo/tview"
	"github.com/atotto/clipboard"
)

type InspectView struct {
	Header   *tview.TextView
	Request  *tview.TextView
	Response *tview.TextView
	Page     *tview.Flex
}

func NewInspectView(ui *UI) *InspectView {
	header := tview.NewTextView().SetTextAlign(tview.AlignCenter)

	request := tview.NewTextView().SetTextAlign(tview.AlignLeft).
		SetScrollable(true).
		SetWrap(true).
		SetDynamicColors(true)
	response := tview.NewTextView().SetTextAlign(tview.AlignRight).
		SetScrollable(true).
		SetWrap(true).
		SetDynamicColors(true)

	flex := tview.NewFlex().AddItem(request, 0, 1, false).AddItem(response, 0, 1, false)

	page := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText("Press TAB to switch between url, request and response Press 'q' to go back to request list. Press 'c' to copy request as curl").SetTextAlign(tview.AlignCenter), 3, 1, false).
		AddItem(header, 5, 1, false).
		AddItem(flex, 0, 1, false)

	view := &InspectView{
		Header:   header,
		Request:  request,
		Response: response,
		Page:     page,
	}

	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'c':
			req, _ := ui.Requests.Get(ui.SelectedRequestID)
			r, err := http.NewRequest(req.Req.Method, req.URL, bytes.NewReader(req.Req.Body))
			if err != nil {
				log.Errorf("Http new request fail %v", r)
				return nil
			}

			r.Header = req.Req.Headers

			cmd, err := http2curl.GetCurlCommand(r)
			if err != nil {
				log.Errorf("Get curl command fail %v", r)
				return nil
			}

			if err = clipboard.WriteAll(cmd.String()); err != nil {
				log.Errorf("Clipboard write all fail %v", r)
				return nil
			}

			return nil
		case 'q':
			ui.App.SetRoot(ui.RequestsView.Page, true).SetFocus(ui.RequestsView.List)
			return nil
		}

		return event
	})

	header.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			ui.App.SetFocus(request)
			return nil
		}

		return event
	})

	request.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			ui.App.SetFocus(response)
			return nil
		}

		return event
	})

	response.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			ui.App.SetFocus(header)
			return nil
		}

		return event
	})

	return view
}
