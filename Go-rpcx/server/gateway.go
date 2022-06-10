package server

import (
	"github.com/smallnest/rpcx/protocol"
	"github.com/soheilhy/cmux"
	"io"
	"net"
)

func (s *Server) startGateway(network string, ln net.Listener) net.Listener {
	if network != "tcp" && network != "tcp4" && network != "tcp6" && network != "reuseport" {
		// log.Infof("network is not tcp/tcp4/tcp6 so can not start gateway")
		return ln
	}

	m := cmux.New(ln)

	rpcxLn := m.Match(rpcxPrefixByteMatcher())

	// mux Plugins
	if s.Plugins != nil {
		s.Plugins.MuxMatch(m)
	}

	go m.Serve()

	return rpcxLn
}

func rpcxPrefixByteMatcher() cmux.Matcher {
	magic := protocol.MagicNumber()
	return func(r io.Reader) bool {
		buf := make([]byte, 1)
		n, _ := r.Read(buf)
		return n == 1 && buf[0] == magic
	}
}
