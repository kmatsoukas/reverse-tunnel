package udp

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/snsinfu/reverse-tunnel/config"
	"github.com/snsinfu/reverse-tunnel/ports"
	"github.com/snsinfu/reverse-tunnel/server/service"
)

// ErrUnauthorizedKey is returned when a key is not authorized.
var ErrUnauthorizedKey = errors.New("unauthorized key")

// ErrInsufficientScope is returned when a key is not allowed to bind to a
// requested port.
var ErrInsufficientScope = errors.New("insufficient scope")

// Service implements service.Service for UDP tunneling service.
type Service struct {
	authorities map[string]ports.Set
	bindings    map[string]*Binder
}

// NewService creates a udp.Service with given server configuration.
func NewService(conf config.Server) Service {
	auths := map[string]ports.Set{}
	bindings := make(map[string]*Binder)

	for _, agent := range conf.Agents {
		set := ports.Set{}

		for _, np := range agent.Ports {
			if np.Protocol == "udp" {
				set.Add(np.Port)
			}
		}

		auths[agent.AuthKey] = set
	}

	return Service{authorities: auths, bindings: bindings}
}

// GetBinder returns a udp.Binder for an agent with given authorization key and
// given UDP port.
func (serv *Service) GetBinder(key string, port int) (service.Binder, error) {
	set, ok := serv.authorities[key]

	if !ok {
		return nil, ErrUnauthorizedKey
	}

	if !set.Has(port) {
		return nil, ErrInsufficientScope
	}

	bindKey := fmt.Sprintf("tcp/%d", port)

	if binder, ok := serv.bindings[bindKey]; ok {
		return binder, nil
	}

	addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))

	if err != nil {
		return nil, err
	}

	pool := service.NewWSPool()
	binder := &Binder{addr: addr, isInitialized: false, wsPool: &pool}

	serv.bindings[bindKey] = binder

	return binder, nil
}
