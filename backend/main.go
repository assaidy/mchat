package main

import (
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

type Message struct {
	from    *Client
	payload []byte
}

type Server struct {
	listenAddr string
	ln         net.Listener
	quitch     chan struct{}
	msgch      chan *Message
	clients    map[*Client]struct{}
	mu         sync.Mutex
}

type Client struct {
	conn net.Conn
    lastActive time.Time // TODO: set rate limit, set deadline
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan struct{}),
		msgch:      make(chan *Message, 10),
		clients:    map[*Client]struct{}{},
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	s.ln = ln

	go s.acceptLoop()
	go s.broadcast()

	<-s.quitch
	close(s.msgch)

	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		log.Println("client connected:", conn.RemoteAddr())

		client := Client{conn: conn}
		s.mu.Lock()
		s.clients[&client] = struct{}{}
		s.mu.Unlock()

		go s.readLoop(&client)
	}
}

func (s *Server) readLoop(c *Client) {
	defer func() {
		log.Println("client disconnected:", c.conn.RemoteAddr())
		s.mu.Lock()
		delete(s.clients, c)
		s.mu.Unlock()
		c.conn.Close()
	}()

	buf := make([]byte, 3*1000)
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			if err == io.EOF { // client disconnected
				return
			}
			log.Println("read error:", err)
			continue
		}
		if m := strings.TrimSpace(string(buf[:n])); m == "" || !utf8.ValidString(m) {
			// ignore whitespace/empty messages
			continue
		}
		msg := Message{
			from:    c,
			payload: buf[:n],
		}
		s.msgch <- &msg
	}
}

func (s *Server) broadcast() {
	for msg := range s.msgch {
		log.Printf("new msg [%s]: %s", msg.from.conn.RemoteAddr(), string(msg.payload))
		for client := range s.clients {
			if client != msg.from {
				_, err := client.conn.Write(msg.payload)
				if err != nil {
					log.Printf("error sending msg: %s from: %s to: %s", string(msg.payload), msg.from.conn.RemoteAddr(), client.conn.RemoteAddr())
				}
			}
		}
	}
}

func main() {
	serverAddr := "localhost:6868"
	log.Println("starting server at:", serverAddr)
	server := NewServer(serverAddr)
	log.Fatal(server.Start())
}
