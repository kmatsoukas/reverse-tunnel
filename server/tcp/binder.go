package tcp

import (
	"net"
	"time"

	"github.com/gorilla/websocket"
	"github.com/snsinfu/reverse-tunnel/server/service"
)

// Timeout value used for checking websocket connection loss.
const connTimeout = 3 * time.Second

// Binder implements service.Binder for TCP tunneling service.
type Binder struct {
	isInitialized bool
	addr          *net.TCPAddr
	wsPool        *service.WSPool
	ln            *net.TCPListener
}

// Start binds to a TCP port and creates tcp.Session for each client connection.
func (binder *Binder) Start(ws *websocket.Conn, store *service.SessionStore) error {
	if !binder.isInitialized {
		var err error
		binder.ln, err = net.ListenTCP("tcp", binder.addr)

		if err != nil {
			return err
		}

		binder.isInitialized = true
	}

	defer binder.ln.Close()

	binder.wsPool.Add(ws)

	go service.Watch(ws, connTimeout, func() error {
		binder.wsPool.Remove(ws)

		if binder.wsPool.IsEmpty() {
			binder.isInitialized = false

			return binder.ln.Close()
		} else {
			return nil
		}
	})

	for {
		conn, err := binder.ln.AcceptTCP()

		if err != nil {
			return err
		}

		sess := NewSession(conn)
		id := store.Add(sess)

		nextWsSocket := binder.wsPool.Next()

		err = nextWsSocket.WriteJSON(service.BinderAcceptMessage{
			Event:       "accept",
			SessionID:   id,
			PeerAddress: conn.RemoteAddr().String(),
		})

		if err != nil {
			return err
		}
	}
}
