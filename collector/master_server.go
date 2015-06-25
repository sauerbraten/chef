package main

import (
	"bufio"
	"log"
	"net"
	"strings"
	"time"
)

type masterServer struct {
	addr string
}

func newMasterServer(host string, port string) *masterServer {
	return &masterServer{addr: host + ":" + port}
}

func (ms *masterServer) getServerList() (servers map[string]*net.UDPAddr, err error) {
	conn, err := net.DialTimeout("tcp", ms.addr, 15*time.Second)
	if err != nil {
		log.Println("failed to connect to master server:", err)
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

		// 12.23.34.45 28785 → 12.23.34.45:28785
		msg = strings.Replace(strings.TrimSpace(msg), " ", ":", -1)

		addr, err = net.ResolveUDPAddr("udp", msg)
		if err != nil {
			return
		}

		servers[addr.String()] = addr
	}

	err = in.Err()

	return
}
