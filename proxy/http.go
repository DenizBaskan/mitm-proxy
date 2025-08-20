package proxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"http-proxy/cert"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/google/uuid"
	utls "github.com/refraction-networking/utls"
	log "github.com/sirupsen/logrus"
)

type HTTPRequest struct {
	ID       string
	Success  bool
	Hostname string
	URL      string
	TLS      bool
	Req      struct {
		Method  string
		Body    []byte
		Headers http.Header
	}
	Res struct {
		StatusCode int
		Body       []byte
		Headers    http.Header
	}
}

type ProxyChan struct {
	ReqChan   chan HTTPRequest
	WsChan    chan WebsocketConnection
	WsMsgChan chan WebsocketMessage
}

func HandleHTTP(clientReader *bufio.Reader, clientConn net.Conn, c ProxyChan) {
	req, err := http.ReadRequest(clientReader)
	if err != nil {
		log.Errorf("Read request fail %v", err)
		return
	}

	hostname := req.URL.Hostname()

	if hostname == "cert" && req.URL.Port() == "" {
		_, err = clientConn.Write([]byte("HTTP/1.1 200 OK\r\n" +
			"Content-Type: application/x-x509-ca-cert\r\n" +
			"Connection: close\r\n" +
			"\r\n"))
		if err != nil {
			log.Errorf("Write fail %v", err)
		}

		f, err := os.Open("data/root_ca/ca.crt")
		if err != nil {
			log.Printf("File open error: %v", err)
			return
		}

		io.Copy(clientConn, f)
		return
	}

	fail := HTTPRequest{ID: uuid.NewString(), Success: false, TLS: false, Hostname: hostname, URL: buildURL(req)}
	fail.Req.Method = req.Method
	fail.Req.Headers = req.Header

	var reqBuf bytes.Buffer
	if err = req.Write(&reqBuf); err != nil {
		c.ReqChan <- fail
		log.Errorf("Write fail %v", err)
		return
	}

	clientReader = bufio.NewReader(io.MultiReader(bytes.NewReader(reqBuf.Bytes()), clientReader))

	serverConn, err := net.Dial("tcp", hostname+":80")
	if err != nil {
		c.ReqChan <- fail
		log.Errorf("Dial fail %v", err)
		return
	}
	serverReader := bufio.NewReader(serverConn)

	handleRequest(hostname, false, clientReader, serverReader, clientConn, serverConn, c)
}

func HandleHTTPS(clientReader *bufio.Reader, clientConn net.Conn, c ProxyChan) {
	req, err := http.ReadRequest(clientReader)
	if err != nil {
		log.Errorf("Read request fail %v", err)
		return
	}

	hostname := req.URL.Hostname()
	if hostname == "" {
		hostname = req.Header.Get("Host")
	}

	port := req.URL.Port()
	if port == "" {
		port = "443"
	}

	fail := HTTPRequest{ID: uuid.NewString(), Success: false, TLS: true, Hostname: hostname, URL: hostname}
	fail.Req.Method = req.Method
	fail.Req.Headers = req.Header

	cert, err := cert.FetchCertificate(hostname)
	if err != nil {
		c.ReqChan <- fail
		log.Errorf("Could not fetch certificate %v", err)
		return
	}

	if _, err := clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n")); err != nil {
		c.ReqChan <- fail
		log.Errorf("Write fail %v", err)
		return
	}

	clientTLSConn := tls.Server(clientConn, &tls.Config{
		Certificates: []tls.Certificate{*cert},
		NextProtos:   []string{"http/1.1"},
	})
	defer clientTLSConn.Close()
	clientReader = bufio.NewReader(clientTLSConn)

	if err = clientTLSConn.Handshake(); err != nil {
		c.ReqChan <- fail
		log.Errorf("Handshake failed %s %v", hostname, err)
		return
	}

	serverConn, err := net.Dial("tcp", hostname+":"+port)
	if err != nil {
		c.ReqChan <- fail
		log.Errorf("Dial failed %v", err)
		return
	}
	defer serverConn.Close()

	serverTLSConn := utls.UClient(serverConn, &utls.Config{
		ServerName: hostname,
		NextProtos: []string{"http/1.1"},
	}, utls.HelloRandomized)
	defer serverTLSConn.Close()
	serverReader := bufio.NewReader(serverTLSConn)

	if err := serverTLSConn.Handshake(); err != nil {
		c.ReqChan <- fail
		log.Errorf("Handshake failed %s %v", hostname, err)
		return
	}

	handleRequest(hostname, true, clientReader, serverReader, clientTLSConn, serverTLSConn, c)
}
