package ips

import (
	"bufio"
	"log"
	"net"
	"strings"
)

var privateNetworks []*net.IPNet

const privateNetworkRanges = `0.0.0.0/8
10.0.0.0/8
100.64.0.0/10
127.0.0.0/8
169.254.0.0/16
172.16.0.0/12
192.0.0.0/24
192.0.0.0/29
192.0.0.170/32
192.0.0.171/32
192.0.2.0/24
192.31.196.0/24
192.52.193.0/24
192.88.99.0/24
192.168.0.0/16
198.18.0.0/15
198.51.100.0/24
203.0.113.0/24
240.0.0.0/4
255.255.255.255/32`

func init() {
	scanner := bufio.NewScanner(strings.NewReader(privateNetworkRanges))

	for scanner.Scan() {
		_, privateNet, err := net.ParseCIDR(scanner.Text())
		if err != nil {
			log.Println(err)
			continue
		}

		privateNetworks = append(privateNetworks, privateNet)
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}

func IsInPrivateNetwork(ip net.IP) bool {
	for _, net := range privateNetworks {
		if net.Contains(ip) {
			return true
		}
	}

	return false
}
