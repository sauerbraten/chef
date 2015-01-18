package main

import (
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/sauerbraten/chef/db"
	"github.com/sauerbraten/extinfo"
)

var storage *db.DB

var privateNet *net.IPNet

func init() {
	_, privateNet, _ = net.ParseCIDR("192.168.0.0/16")
}

func main() {
	var err error
	storage, err = db.New()
	if err != nil {
		log.Fatal(err)
	}
	defer storage.Close()

	ms := newMasterServer(conf.MasterServerAddress, conf.MasterServerPort)

	for start := range time.Tick(conf.ScanIntervalSeconds * time.Second) {
		log.Println("refreshing server list after tick at", start.String())

		list, err := getCompleteServerList(ms)
		if err != nil {
			log.Println(err)
		}

		log.Println("running scan...")

		for _, serverAddress := range list {
			scanServer(serverAddress)
		}

		log.Printf("scan finished (took %v)", time.Since(start))
	}
}

// get master server list, add hidden servers, remove blacklisted servers
func getCompleteServerList(ms *masterServer) (list map[string]*net.UDPAddr, err error) {
	list, err = ms.getServerList()
	if err != nil {
		// still search hidden servers
		list = map[string]*net.UDPAddr{}
	}

	for _, serverAddress := range conf.ExtraServers {
		var addr *net.UDPAddr
		addr, err = net.ResolveUDPAddr("udp", serverAddress)
		if err != nil {
			return
		}

		// don't add servers that are already contained in master server list
		if _, ok := list[addr.IP.String()+":"+strconv.Itoa(addr.Port)]; ok {
			continue
		} else {
			list[addr.IP.String()+":"+strconv.Itoa(addr.Port)] = addr
		}
	}

	for _, serverAddress := range conf.BlacklistedServers {
		var deletePrefix string

		if strings.Count(serverAddress, ":") == 0 {
			var addr *net.IPAddr
			addr, err = net.ResolveIPAddr("ip", serverAddress)
			if err != nil {
				return
			}

			deletePrefix = addr.IP.String()
		} else {
			var addr *net.UDPAddr
			addr, err = net.ResolveUDPAddr("udp", serverAddress)
			if err != nil {
				return
			}

			deletePrefix = addr.IP.String() + ":" + strconv.Itoa(addr.Port)
		}

		for serverAddress, _ := range list {
			if strings.HasPrefix(serverAddress, deletePrefix) {
				delete(list, serverAddress)
			}
		}
	}

	return
}

func scanServer(serverAddress *net.UDPAddr) {
	s, err := extinfo.NewServer(serverAddress.IP.String(), serverAddress.Port, 1*time.Second)
	if err != nil {
		verbose(err)
		return
	}

	basicInfo, err := s.GetBasicInfo()
	if err != nil {
		verbose("error getting basic info from", serverAddress, ":", err)
		return
	}

	serverID := storage.GetServerId(serverAddress.IP.String(), serverAddress.Port, basicInfo.Description)

	playerInfos, err := s.GetAllClientInfo()
	if err != nil {
		verbose("error getting client info from", serverAddress, ":", err)
		return
	}

	if len(playerInfos) == 0 {
		return
	}

	log.Println("found", len(playerInfos), "players on", basicInfo.Description, serverAddress.String())

	for _, playerInfo := range playerInfos {
		ip := playerInfo.IP

		if ip.Equal(net.ParseIP("0.0.0.0")) || privateNet.Contains(ip) {
			// no useable IP → useless, don't save
			continue
		}

		storage.AddOrIgnoreSighting(playerInfo.Name, ip, serverID)
	}
}

func verbose(args ...interface{}) {
	if conf.Verbose {
		log.Println(args...)
	}
}
