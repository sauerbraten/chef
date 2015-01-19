package main

import (
	"log"
	"net"
	"strconv"
	"time"

	"github.com/sauerbraten/chef/db"
	"github.com/sauerbraten/chef/ips"
	"github.com/sauerbraten/extinfo"
)

var storage *db.DB

func main() {
	var err error
	storage, err = db.New()
	if err != nil {
		log.Fatal(err)
	}
	defer storage.Close()

	// resolve addresses given in config
	err = finishConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	ms := newMasterServer(conf.MasterServerAddress, conf.MasterServerPort)

	for start := range time.Tick(conf.ScanIntervalSeconds * time.Second) {
		log.Println("refreshing server list after tick at", start.String())

		list := getServerList(ms)

		log.Println("running scan...")

		for _, serverAddress := range list {
			scanServer(serverAddress)
		}

		log.Printf("scan finished (took %v)", time.Since(start))
	}
}

func finishConfiguration() (err error) {
	// resolve extra servers
	for _, serverAddress := range conf.ExtraServers {
		var addr *net.UDPAddr

		addr, err = net.ResolveUDPAddr("udp", serverAddress)
		if err != nil {
			return
		}

		conf.extraServers = append(conf.extraServers, addr)
	}

	// resolve greylisted servers
	for _, serverAddress := range conf.GreylistedServers {
		var addr *net.IPAddr

		addr, err = net.ResolveIPAddr("ip", serverAddress)
		if err != nil {
			return
		}

		conf.greylistedServers[addr.IP.String()] = true
	}

	return
}

// get master server list and add manually specified extra servers
func getServerList(ms *masterServer) (list map[string]*net.UDPAddr) {
	var err error

	list, err = ms.getServerList()
	if err != nil {
		log.Println("error getting master server list:", err)
		// still search extra servers
		list = map[string]*net.UDPAddr{}
	}

	for _, addr := range conf.extraServers {
		list[addr.IP.String()+":"+strconv.Itoa(addr.Port)] = addr
	}

	return
}

func scanServer(serverAddress *net.UDPAddr) {
	s, err := extinfo.NewServer(serverAddress.IP.String(), serverAddress.Port, 2*time.Second)
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
		// don't save bot sightings
		if playerInfo.ClientNum > 127 {
			continue
		}

		// check IP
		if ips.IsInPrivateNetwork(playerInfo.IP) || conf.greylistedServers[serverAddress.IP.String()] {
			playerInfo.IP = net.ParseIP("0.0.0.0")
		}

		storage.AddOrIgnoreSighting(playerInfo.Name, playerInfo.IP, serverID)
	}
}

func verbose(args ...interface{}) {
	if conf.Verbose {
		log.Println(args...)
	}
}
