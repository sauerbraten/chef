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

func (s *Server) ServerList() (servers []string, err error) {
	conn, err := net.DialTimeout("tcp", s.addr, s.timeout)
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

	for in.Scan() {
		cmd := in.Text()
		if !strings.HasPrefix(cmd, "addserver ") || cmd == "\x00" {
			continue
		}

		addr := strings.TrimPrefix(cmd, "addserver ")
		addr = strings.TrimSpace(addr)

		// 12.23.34.45 28785 → 12.23.34.45:28785
		addr = strings.Replace(addr, " ", ":", -1)

		servers = append(servers, addr)
	}

	err = in.Err()

	return
}

func (s *Server) Address() string { return s.addr }
