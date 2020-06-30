package udp

import (
	"net"
	"time"

	"github.com/gorilla/websocket"
	"github.com/snsinfu/reverse-tunnel/config"
	"github.com/snsinfu/reverse-tunnel/server/service"
)

// connTimeout is the timeout used for checking websocket connection loss.
const connTimeout = 3 * time.Second

// Binder implements service.Binder for UDP tunneling service.
type Binder struct {
	isInitialized bool
	addr          *net.UDPAddr
	wsPool        *service.WSPool
	conn          *net.UDPConn
}

// Start binds to a UDP port and routes incoming packets to udp.Session objects.
func (binder *Binder) Start(ws *websocket.Conn, store *service.SessionStore) error {
	if !binder.isInitialized {
		var err error
		binder.conn, err = net.ListenUDP("udp", binder.addr)

		if err != nil {
			return err
		}

		binder.isInitialized = true
	}

	defer binder.conn.Close()

	go service.Watch(ws, connTimeout, func() error {
		binder.wsPool.Remove(ws)

		if binder.wsPool.IsEmpty() {
			binder.isInitialized = false

			return binder.conn.Close()
		} else {
			return nil
		}
	})

	buf := make([]byte, config.BufferSize)

	for {
		n, peer, err := binder.conn.ReadFromUDP(buf)

		if err != nil {
			return err
		}

		if sess, ok := store.Get(peer).(*Session); ok {
			sess.SendToAgent(buf[:n])
		} else {
			sess := NewSession(binder.conn, peer)
			id := store.Add(sess)

			nextWsSocket := binder.wsPool.Next()

			err = nextWsSocket.WriteJSON(service.BinderAcceptMessage{
				Event:       "accept",
				SessionID:   id,
				PeerAddress: peer.String(),
			})

			if err != nil {
				return err
			}

			// NOTE: The message is dropped here, which is acceptable since it
			// is UDP. But it makes a noticeable delay, for example, on a mosh
			// handshake. Maybe the message should be queued.
		}
	}
}
