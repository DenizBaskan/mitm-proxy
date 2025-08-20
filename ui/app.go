package ui

import (
	"fmt"
	"http-proxy/proxy"

	"github.com/rivo/tview"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type Event struct {
	EventID   string
	EventType string // "request" "websocket"
}

type UI struct {
	App               *tview.Application
	ProxyChan         proxy.ProxyChan
	Requests          *orderedmap.OrderedMap[string, proxy.HTTPRequest]
	Websockets        *orderedmap.OrderedMap[string, proxy.WebsocketConnection]
	WebsocketMessages *orderedmap.OrderedMap[string, []proxy.WebsocketMessage]
	Events            []Event

	SelectedHostname    string
	SelectedWebsocketID string
	SelectedRequestID   string

	RequestsView  *RequestsView
	HostnamesView *HostnamesView
	InspectView   *InspectView
	WebsocketView *WebsocketView
}

func NewUI(c proxy.ProxyChan) *UI {
	app := UI{
		App:               tview.NewApplication(),
		Requests:          orderedmap.New[string, proxy.HTTPRequest](),
		Websockets:        orderedmap.New[string, proxy.WebsocketConnection](),
		WebsocketMessages: orderedmap.New[string, []proxy.WebsocketMessage](),
		ProxyChan:         c,
	}

	app.RequestsView = NewRequestsView(&app)
	app.HostnamesView = NewHostnamesView(&app)
	app.InspectView = NewInspectView(&app)
	app.WebsocketView = NewWebsocketView(&app)

	return &app
}

func (ui *UI) Run() {
	go func() {
		for req := range ui.ProxyChan.ReqChan {
			ui.Events = append(ui.Events, Event{EventID: req.ID, EventType: "request"})
			ui.Requests.Set(req.ID, req)

			ui.App.QueueUpdateDraw(func() {
				currentIndex := ui.HostnamesView.List.GetCurrentItem()
				ui.HostnamesView.List.Clear()

				m := orderedmap.New[string, int]()

				for pair := ui.Requests.Oldest(); pair != nil; pair = pair.Next() {
					tmp, _ := m.Get(pair.Value.Hostname)
					m.Set(pair.Value.Hostname, tmp + 1)
				}

				for pair := m.Oldest(); pair != nil; pair = pair.Next() {
					ui.HostnamesView.List.AddItem(fmt.Sprintf("[%d] %s", pair.Value, pair.Key), pair.Key, 0, nil)
				}

				ui.HostnamesView.List.SetCurrentItem(currentIndex)
			})

			if req.Hostname == ui.SelectedHostname {
				ui.App.QueueUpdateDraw(func() {
					currentIndex := ui.RequestsView.List.GetCurrentItem()
					ui.RequestsView.List.AddItem(craftRequestLine(req), req.Hostname, 0, nil)
					ui.RequestsView.List.SetCurrentItem(currentIndex)
				})
			}
		}
	}()

	go func() {
		for ws := range ui.ProxyChan.WsChan {
			ui.Events = append(ui.Events, Event{EventID: ws.ID, EventType: "websocket"})
			ui.Websockets.Set(ws.ID, ws)

			if ws.Hostname == ui.SelectedHostname {
				ui.App.QueueUpdateDraw(func() {
					currentIndex := ui.RequestsView.List.GetCurrentItem()

					url := ws.URL
					if len(url) > 50 {
						url = url[50:] + "..."
					}

					ui.RequestsView.List.AddItem(fmt.Sprintf("[WEBSOCKET CONN] %s", formatWebsocketURL(ws.TLS, url)), ws.ID, 0, nil)
					ui.RequestsView.List.SetCurrentItem(currentIndex)
				})
			}
		}
	}()

	go func() {
		for msg := range ui.ProxyChan.WsMsgChan {
			msgs, _ := ui.WebsocketMessages.Get(msg.WebsocketID)
			ui.WebsocketMessages.Set(msg.WebsocketID, append(msgs, msg))

			if msg.WebsocketID == ui.SelectedWebsocketID {
				ui.App.QueueUpdateDraw(func() {
					currentIndex := ui.WebsocketView.List.GetCurrentItem()

					ui.WebsocketView.List.AddItem(craftWebsocketMessage(msg), string(msg.Message), 0, nil)
					
					ui.WebsocketView.List.SetCurrentItem(currentIndex)
				})
			}
		}
	}()

	if err := ui.App.SetRoot(ui.HostnamesView.Page, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}
}
