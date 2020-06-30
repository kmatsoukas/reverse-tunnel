package service

import (
	"sync"

	"github.com/gorilla/websocket"
)

// WSPool holds an array of web socket connections for a binder
type WSPool struct {
	sync.Mutex

	pool      []*websocket.Conn
	nextIndex int
}

// IsEmpty returns true if the pool is empty
func (wsPool *WSPool) IsEmpty() bool {
	wsPool.Lock()
	defer wsPool.Unlock()

	poolSize := len(wsPool.pool)

	if poolSize == 0 {
		return true
	} else {
		return false
	}
}

// Add a web socket connection to the pool.
func (wsPool *WSPool) Add(ws *websocket.Conn) error {
	wsPool.Lock()
	defer wsPool.Unlock()

	wsPool.pool = append(wsPool.pool, ws)

	return nil
}

// Remove a web socket connection from the pool
func (wsPool *WSPool) Remove(ws *websocket.Conn) {
	wsPool.Lock()
	defer wsPool.Unlock()

	wsIndex := -1

	for i, w := range wsPool.pool {
		if w == ws {
			wsIndex = i
		}
	}

	if wsIndex != -1 {
		wsPool.pool = append(wsPool.pool[:wsIndex], wsPool.pool[wsIndex+1:]...)
	}
}

// Next returns the next web socket connection to use from the pool.
func (wsPool *WSPool) Next() *websocket.Conn {
	wsPool.Lock()
	defer wsPool.Unlock()

	poolLength := len(wsPool.pool)
	nextIndex := wsPool.nextIndex

	if poolLength > 1 {
		if nextIndex+1 >= poolLength {
			wsPool.nextIndex = 0
		} else {
			wsPool.nextIndex++
		}
	}

	return wsPool.pool[nextIndex]
}

// NewWSPool returns a new instance of WSPool
func NewWSPool() WSPool {
	pool := []*websocket.Conn{}

	return WSPool{
		pool:      pool,
		nextIndex: 0,
	}
}
