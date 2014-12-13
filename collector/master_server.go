package main

import (
	"bufio"
	"log"
	"net"
	"strconv"
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

		msgParts := strings.Split(msg, " ")

		addr, err = net.ResolveUDPAddr("udp", msgParts[1]+":"+msgParts[2])
		if err != nil {
			return
		}

		servers[addr.IP.String()+":"+strconv.Itoa(addr.Port)] = addr
	}

	err = in.Err()

	return
}
