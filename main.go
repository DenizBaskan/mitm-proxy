package main

import (
	"bufio"
	"http-proxy/proxy"
	"http-proxy/ui"
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	f, err := os.OpenFile("data/logs/errors.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(f)
	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	c := proxy.ProxyChan{
		ReqChan:   make(chan proxy.HTTPRequest),
		WsChan:    make(chan proxy.WebsocketConnection),
		WsMsgChan: make(chan proxy.WebsocketMessage),
	}

	ui := ui.NewUI(c)
	go ui.Run()

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Errorf("Accept fail %v", err)
			return
		}

		go func() {
			defer func() {
				conn.Close()

				if r := recover(); r != nil {
					log.Errorf("Panic %v", r)
				}
			}()

			reader := bufio.NewReader(conn)

			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			b, err := reader.Peek(7)
			if err != nil {
				log.Errorf("Peak fail %v", err)
				return
			}
			conn.SetReadDeadline(time.Time{})

			if string(b) == "CONNECT" {
				proxy.HandleHTTPS(reader, conn, c)
				return
			}

			proxy.HandleHTTP(reader, conn, c)
		}()
	}
}
