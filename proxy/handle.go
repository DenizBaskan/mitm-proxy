package proxy

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

func handleRequest(hostname string, tls bool, clientReader, serverReader *bufio.Reader, client, server net.Conn, c ProxyChan) {
	for {
		req, err := http.ReadRequest(clientReader)
		if err != nil {
			if err != io.EOF {
				log.Errorf("Read request fail %v", err)
			}
			return
		}
		defer req.Body.Close()

		request := HTTPRequest{ID: uuid.NewString(), Success: false, TLS: tls, Hostname: hostname, URL: buildURL(req)}
		request.Req.Method = req.Method
		request.Req.Headers = req.Header

		var reqBodyBuf bytes.Buffer
		tee := io.TeeReader(req.Body, &reqBodyBuf)
		req.Body = io.NopCloser(tee)

		if err = req.Write(server); err != nil {
			c.ReqChan <- request
			log.Errorf("Write fail %v", err)
			return
		}

		reqBody, err := decodeBody(req.Header.Get("Content-Encoding"), reqBodyBuf)
		if err != nil {
			c.ReqChan <- request
			log.Errorf("Decode body fail %v", err)
			return
		}
		request.Req.Body = reqBody

		res, err := http.ReadResponse(serverReader, req)
		if err != nil {
			c.ReqChan <- request
			log.Errorf("Read response fail %v", err)
			return
		}
		defer res.Body.Close()

		request.Res.StatusCode = res.StatusCode
		request.Res.Headers = res.Header

		var resBodyBuf bytes.Buffer
		tee = io.TeeReader(res.Body, &resBodyBuf)
		res.Body = io.NopCloser(tee)

		if err = res.Write(client); err != nil {
			c.ReqChan <- request
			log.Errorf("Write fail %v", err)
			return
		}

		resBody, err := decodeBody(res.Header.Get("Content-Encoding"), resBodyBuf)
		if err != nil {
			c.ReqChan <- request
			log.Errorf("Decode body fail %v", err)
			return
		}

		request.Res.Body = resBody
		request.Success = true
		c.ReqChan <- request

		if websocket.IsWebSocketUpgrade(req) && res.StatusCode == 101 {
			id := uuid.NewString()
			c.WsChan <- WebsocketConnection{ID: id, Hostname: hostname, URL: buildURL(req), TLS: tls}
			handleWS(id, client, server, c)

			return
		}
	}
}
