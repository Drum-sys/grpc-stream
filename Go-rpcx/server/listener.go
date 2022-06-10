package server

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
)

var makeListeners = make(map[string]MakeListener) //建立多个连接，存到map

func init() {
	makeListeners["tcp"] = tcpMakeListener("tcp")
	makeListeners["tcp4"] = tcpMakeListener("tcp4")
	makeListeners["tcp6"] = tcpMakeListener("tcp6")
	makeListeners["http"] = tcpMakeListener("tcp")
	makeListeners["ws"] = tcpMakeListener("tcp")
	makeListeners["wss"] = tcpMakeListener("tcp")
}

type MakeListener func(s *Server, address string) (In net.Listener, err error)

// RegisterMakeListener registers a MakeListener for network.
func RegisterMakeListener(network string, ml MakeListener) {
	makeListeners[network] = ml
}

func (s *Server) makeListener(network, address string) (In net.Listener, err error) {
	ml := makeListeners[network]
	if ml == nil {
		return nil, fmt.Errorf("cna not make listener for %s", network)
	}

	if network == "wss" && s.tlsConfig == nil {
		return nil, errors.New("must set tls config for wss")
	}
	return ml(s, address)
}

// 创建TCP Listener
func tcpMakeListener(network string) MakeListener {
	return func(s *Server, address string) (In net.Listener, err error) {
		if s.tlsConfig == nil {
			In, err = net.Listen(network, address) // listen返回一个Listener
		} else {
			In, err = tls.Listen(network, address, s.tlsConfig)
		}
		return In, err
	}
}
