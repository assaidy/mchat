package main

// TODO: implement token authentication

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

const Port = ":3000"

type Server struct {
	Clients  map[*websocket.Conn]struct{}
	Messages []Message
	MsgCh    chan Message
}

type MessageRequest struct {
	Text string `json:"text"`
}

func NewServer() *Server {
	server := &Server{
		Clients:  make(map[*websocket.Conn]struct{}),
		Messages: make([]Message, 0),
		MsgCh:    make(chan Message),
	}
	go server.broadcast()

	return server
}

func (s *Server) handleWS(ws *websocket.Conn) {
	log.Printf("new connection from client: %s", ws.RemoteAddr())
	s.Clients[ws] = struct{}{}
	// sync messages history
	for _, msg := range s.Messages {
		if err := websocket.JSON.Send(ws, msg.Response()); err != nil {
			log.Printf("error sending msg (%s): %v", ws.RemoteAddr(), err)
		}
	}
	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	name := getClientName(ws)
	s.MsgCh <- &MessageJoinChat{Name: name}

	defer func() {
		ws.Close()
		log.Printf("connection closed with client: %s", ws.RemoteAddr())
		s.MsgCh <- &MessageLeaveChat{Name: name}
	}()

	for {
		buf := make([]byte, 1024)
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				delete(s.Clients, ws)
				return
			}
			log.Printf("error reading msg (%s): %v", ws.RemoteAddr(), err)
			continue
		}

		var msg MessageRequest
		err = json.Unmarshal(buf[:n], &msg)
		if err != nil {
			log.Printf("error unmarshal msg (%s): %v", ws.RemoteAddr(), err)
			continue
		}

		log.Printf("new message (%s): %v", ws.RemoteAddr(), msg.Text)

		s.MsgCh <- &MessageChat{Name: name, Text: msg.Text}
	}
}

func (s *Server) broadcast() {
	for msg := range s.MsgCh {
		s.Messages = append(s.Messages, msg)
		for ws := range s.Clients {
			go func(ws *websocket.Conn) {
				if err := websocket.JSON.Send(ws, msg.Response()); err != nil {
					log.Printf("error sending msg (%s): %v", ws.RemoteAddr(), err)
				}
			}(ws)
		}
	}
}

type Message interface {
	Response() map[string]string
}

type MessageJoinChat struct {
	Name string
}

func (m *MessageJoinChat) Response() map[string]string {
	return map[string]string{
		"type": "join",
		"text": fmt.Sprintf("%s joined chat", m.Name),
	}
}

type MessageLeaveChat struct {
	Name string
}

func (m *MessageLeaveChat) Response() map[string]string {
	return map[string]string{
		"type": "leave",
		"text": fmt.Sprintf("%s left chat", m.Name),
	}
}

type MessageChat struct {
	Name string
	Text string
}

func (m *MessageChat) Response() map[string]string {
	return map[string]string{
		"type":   "chat",
		"sender": m.Name,
		"text":   m.Text,
	}
}

func main() {
	s := NewServer()
	http.Handle("/ws", websocket.Handler(s.handleWS))

	log.Printf("starting server on port %v", Port)
	http.ListenAndServe(Port, nil)
}

func getClientName(ws *websocket.Conn) string {
	buf := make([]byte, 1024)
	n, err := ws.Read(buf)
	if err != nil {
		log.Printf("error reading client name (%s): %v", ws.RemoteAddr(), err)
		return "UNKNOWN"
	}

	msg := struct {
		Name string `json:"name"`
	}{}

	err = json.Unmarshal(buf[:n], &msg)
	if err != nil {
		log.Printf("error unmarshal client name (%s): %v", ws.RemoteAddr(), err)
		return "UNKNOWN"
	}

	return msg.Name
}
