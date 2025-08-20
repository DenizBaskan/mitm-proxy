package ui

import (
	"fmt"
	"http-proxy/proxy"
	"net/http"
	"strings"
	"unicode"

	"github.com/gobwas/ws"
)

func formatWebsocketURL(tls bool, url string) string {
	if tls {
		url = strings.Replace(url, "https://", "wss://", 1)
	} else {
		url = strings.Replace(url, "http://", "ws://", 1)
	}

	return url
}

func sanitizeBody(body string) string {
	result := make([]rune, 0, len(body))

	for _, r := range body {
		// some characters are falsely censored like '<'
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsPunct(r) || unicode.IsSpace(r) {
			result = append(result, r)
		} else {
			result = append(result, '�')
		}
	}

	return string(result)
}

func craftRequestLine(req proxy.HTTPRequest) string {
	status := "FAIL"
	if req.Success {
		status = fmt.Sprint(req.Res.StatusCode)
	}

	url := req.URL
	if len(url) > 50 {
		url = url[:50] + "..."
	}

	return fmt.Sprintf("[%s - %s] %s", req.Req.Method, status, url)
}

func craftInspect(request proxy.HTTPRequest) (string, string) {
	req := formatHeaders(request.Req.Headers)
	body := sanitizeBody(string(request.Req.Body))
	req += "\n\n[white]" + body + "\n"

	res := formatHeaders(request.Res.Headers)
	body = sanitizeBody(string(request.Res.Body))
	res += "\n\n[white]" + body + "\n"

	return req, res
}

func craftWebsocketMessage(msg proxy.WebsocketMessage) string {
	message := "[Server - "
	if msg.Sender == "client" {
		message = "[Client - "
	}

	switch msg.Opcode {
	case ws.OpContinuation:
		message += "Continuation"
	case ws.OpText:
		message += "Text"
	case ws.OpBinary:
		message += "Binary"
	case ws.OpClose:
		message += "Close"
	case ws.OpPing:
		message += "Ping"
	case ws.OpPong:
		message += "Pong"
	}

	message += "] "

	if msg.Opcode == ws.OpText {
		m := string(msg.Message)
		if len(m) > 50 {
			m = m[50:] + "..."
		}
		message += m
	}

	return message
}

func formatHeaders(h http.Header) string {
	var (
		s = ""
		i = 0
	)

	for name, values := range h {
		for _, value := range values {
			if i%2 == 0 {
				s += fmt.Sprintf("[green]%s: %s\n", name, value)
			} else {
				s += fmt.Sprintf("[blue]%s: %s\n", name, value)
			}

			i++
		}
	}

	return s
}
