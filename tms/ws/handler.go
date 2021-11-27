// Package ws provides a handler to maintain ws. It uses envelope interface.
package ws

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"

	"prisma/gogroup"
	"prisma/tms/envelope"

	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

// WebSocket Subscriber
type Handler struct {
	Streamer     Streamer
	ctxt         gogroup.GoGroup
	clients      map[*Client]bool
	clientsMutex sync.RWMutex
}

type Streamer interface {
	ToClient(ctxt gogroup.GoGroup, client *Client)
}

// constructor
func NewHandler(ctxt gogroup.GoGroup) *Handler {
	return &Handler{
		ctxt:         ctxt,
		clients:      make(map[*Client]bool),
		clientsMutex: sync.RWMutex{},
	}
}

// Subscriber
func (h *Handler) Publish(envelope envelope.Envelope) {
	sourceSessionId := envelope.Source
	// broadcast by default
	private := false
	// private check
	if nil != envelope.GetSession() {
		private = true
	}
	// do not expose others sessionId
	envelope.Source = ""
	marshaller := jsonpb.Marshaler{}
	payload := bytes.Buffer{}
	err := marshaller.Marshal(&payload, &envelope)
	if err == nil {
		h.clientsMutex.RLock()
		clients := h.clients
		h.clientsMutex.RUnlock()
		for wClient := range clients {
			go func(client *Client) {
				if private {
					if client.sessionId == sourceSessionId {
						select {
						case client.Send <- payload.String():
						default:
							close(client.Send)
							h.clientsMutex.Lock()
							delete(h.clients, client)
							h.clientsMutex.Unlock()
						}
						if envelope.Type == "Session/TERMINATE" || envelope.Type == "Session/IDLE" {
							close(client.Send)
							h.clientsMutex.Lock()
							delete(h.clients, client)
							h.clientsMutex.Unlock()
						}
					}
					// else ignore
				} else {
					// broadcast but do not send back to originating session
					if client.sessionId != sourceSessionId {
						select {
						case client.Send <- payload.String():
						default:
							close(client.Send)
							h.clientsMutex.Lock()
							delete(h.clients, client)
							h.clientsMutex.Unlock()
						}
					}
				}
			}(wClient)
		}
	}
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func (h *Handler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	conn, err := upgrader.Upgrade(responseWriter, request, nil)
	ctxt := h.ctxt.Child("")
	if err != nil {
		// FIXME handle error
		fmt.Println(err)
	}
	key := request.Header.Get("Sec-WebSocket-Key")
	sessionId, err := request.Cookie("id")
	if err == nil {
		// FIXME check session is valid
		// FIXME close old websocket associated with sessionId (C2 reloaded)
		//fmt.Println("creating socket:" + key)
		//fmt.Println("sessionId:" + sessionId.Value)
		client := &Client{
			key:       key,
			sessionId: sessionId.Value,
			Conn:      conn,
			Send:      make(chan string),
		}
		h.clientsMutex.Lock()
		h.clients[client] = true
		h.clientsMutex.Unlock()
		ctxt.Go(func() { client.init(ctxt) })
		if h.Streamer != nil {
			h.Streamer.ToClient(ctxt, client)
		}
	} else {
		fmt.Println("no sessionId, socket:" + key)
		conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(writeWait))
		conn.Close()
	}
}

type Client struct {
	key       string
	sessionId string
	Conn      *websocket.Conn
	Send      chan string
}

// readData is needed to determine the status of ws
func (client *Client) readData(ctx gogroup.GoGroup) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if _, _, err := client.Conn.NextReader(); err != nil {
			client.Conn.Close()
			break
		}
	}
}

func (client *Client) init(ctxt gogroup.GoGroup) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.Conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(writeWait))
		client.Conn.Close()
	}()
	client.
		Conn.
		SetPongHandler(func(string) error {
			client.Conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})
	ctxt.Go(func() { client.readData(ctxt) })
	for {
		select {
		case payload, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				ctxt.Cancel(errors.New("closed channel"))
				return
			}
			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				ctxt.Cancel(err)
				return
			}
			w.Write([]byte(payload))
			if err := w.Close(); err != nil {
				ctxt.Cancel(err)
				return
			}
		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				// FIXME handle error
				ctxt.Cancel(err)
				return
			}
		case <-ctxt.Done():
			return
		}
	}
}
