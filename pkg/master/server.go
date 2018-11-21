package master

import (
	"bufio"
	"net"
	"strings"
	"time"
)

type Server struct {
	addr    string
	timeout time.Duration
}

func New(addr string, timeout time.Duration) *Server {
	return &Server{
		addr:    addr,
		timeout: timeout,
	}
}

func (ms *Server) ServerList() (servers map[string]*net.UDPAddr, err error) {
	conn, err := net.DialTimeout("tcp", ms.addr, ms.timeout)
	if err != nil {
		return
	}
	defer conn.Close()

	in := bufio.NewScanner(conn)
	out := bufio.NewWriter(conn)

	// request list

	_, err = out.WriteString("list\n")
	if err != nil {
		return
	}

	err = out.Flush()
	if err != nil {
		return
	}

	// receive list

	var addr *net.UDPAddr
	servers = map[string]*net.UDPAddr{}

	for in.Scan() {
		msg := in.Text()
		if msg == "\x00" {
			// end of list
			continue
		}

		msg = strings.TrimPrefix(msg, "addserver ")
		msg = strings.TrimSpace(msg)

		// 12.23.34.45 28785 → 12.23.34.45:28785
		msg = strings.Replace(msg, " ", ":", -1)

		addr, err = net.ResolveUDPAddr("udp", msg)
		if err != nil {
			return
		}

		servers[addr.String()] = addr
	}

	err = in.Err()

	return
}
