package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type WebsocketView struct {
	UI     *UI
	Header *tview.TextView
	List   *tview.List
	Page   *tview.Flex
}

func NewWebsocketView(ui *UI) *WebsocketView {
	text := tview.NewTextView().SetScrollable(true)
	textPage := tview.NewFlex().AddItem(text, 0, 1, false)

	header := tview.NewTextView().SetTextAlign(tview.AlignCenter)
	list := tview.NewList().ShowSecondaryText(false)

	page := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText("Press TAB to switch between url and websocket stream. Press 'q' to go back to request list").SetTextAlign(tview.AlignCenter), 3, 1, false).
		AddItem(header, 5, 1, false).
		AddItem(list, 0, 1, false)

	view := &WebsocketView{
		UI:     ui,
		Header: header,
		List:   list,
		Page:   page,
	}

	header.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			ui.App.SetFocus(list)
			return nil
		}

		return event
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			ui.App.SetFocus(header)
			return nil
		}

		return event
	})

	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			ui.App.SetRoot(ui.RequestsView.Page, true).SetFocus(ui.RequestsView.List)
			return nil
		}

		return event
	})

	textPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			ui.App.SetRoot(ui.WebsocketView.Page, true).SetFocus(ui.WebsocketView.List)
			return nil
		}

		return event
	})

	list.SetSelectedFunc(func(_ int, _ string, message string, _ rune) {
		text.SetText(message)
		ui.App.SetRoot(textPage, true).SetFocus(text)
	})

	return view
}
