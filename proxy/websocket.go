package proxy

import (
	"net"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	log "github.com/sirupsen/logrus"
)

type WebsocketConnection struct {
	ID       string
	Hostname string
	URL      string
	TLS      bool
}

type WebsocketMessage struct {
	WebsocketID string
	Message     []byte
	Opcode      ws.OpCode
	Sender      string
}

func handleWS(websocketID string, client net.Conn, server net.Conn, c ProxyChan) {
	go func() {
		for {
			msg, op, err := wsutil.ReadClientData(client)
			if err != nil {
				log.Errorf("Read client data fail %v", err)
				return
			}

			c.WsMsgChan <- WebsocketMessage{WebsocketID: websocketID, Message: msg, Opcode: op, Sender: "client"}

			err = wsutil.WriteClientMessage(server, op, msg)
			if err != nil {
				log.Errorf("Write server message fail %v", err)
				return
			}
		}
	}()

	for {
		msg, op, err := wsutil.ReadServerData(server)
		if err != nil {
			log.Errorf("Read client data fail %v", err)
			return
		}

		c.WsMsgChan <- WebsocketMessage{WebsocketID: websocketID, Message: msg, Opcode: op, Sender: "server"}

		err = wsutil.WriteServerMessage(client, op, msg)
		if err != nil {
			log.Errorf("Write server message fail %v", err)
			return
		}
	}
}
