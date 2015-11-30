package main

import (
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/sauerbraten/chef/db"
	"github.com/sauerbraten/chef/ips"
	"github.com/sauerbraten/extinfo"
)

var storage *db.Database

func main() {
	var err error
	storage, err = db.New()
	if err != nil {
		log.Fatal(err)
	}
	defer storage.Close()

	ms := newMasterServer(conf.MasterServerAddress)

	for start := range time.Tick(conf.ScanIntervalSeconds * time.Second) {
		// resolve addresses given in config
		resolveConfigServers()

		log.Println("refreshing server list after tick at", start.String())

		list := getExtendedServerList(ms)

		log.Println("running scan...")

		var wg sync.WaitGroup

		for _, serverAddress := range list {
			wg.Add(1)
			go func(serverAddress *net.UDPAddr) {
				scanServer(serverAddress)
				wg.Done()
			}(serverAddress)
		}

		wg.Wait()

		log.Printf("scan finished (took %v)", time.Since(start))
	}
}

func resolveConfigServers() {
	// clear old server lists
	conf.extraServers = make([]*net.UDPAddr, 0, len(conf.ExtraServers))
	conf.greylistedServers = map[string]bool{}

	// resolve extra servers
	for _, serverAddress := range conf.ExtraServers {
		addr, err := net.ResolveUDPAddr("udp", serverAddress)
		if err != nil {
			log.Println("error resolving "+serverAddress+":", err)
			continue
		}

		conf.extraServers = append(conf.extraServers, addr)
	}

	// resolve greylisted servers
	for _, serverAddress := range conf.GreylistedServers {
		addr, err := net.ResolveIPAddr("ip", serverAddress)
		if err != nil {
			log.Println("error resolving "+serverAddress+":", err)
			continue
		}

		conf.greylistedServers[addr.IP.String()] = true
	}
}

// Returns the master server list, extended by manually specified extra servers
func getExtendedServerList(ms *masterServer) (list map[string]*net.UDPAddr) {
	var err error

	list, err = ms.getServerList()
	if err != nil {
		log.Println("error getting master server list:", err)
		// still going to search extra servers, so we need a valid map
		list = map[string]*net.UDPAddr{}
	}

	for _, addr := range conf.extraServers {
		list[addr.IP.String()+":"+strconv.Itoa(addr.Port)] = addr
	}

	return
}

// scans a server and inserts sightings into the database
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

	verbose("found", len(playerInfos), "players on", basicInfo.Description, serverAddress.String())

	for _, playerInfo := range playerInfos {
		// don't save bot sightings
		if playerInfo.ClientNum > extinfo.MAX_PLAYER_CN {
			continue
		}

		// check for valid client IP
		if conf.greylistedServers[serverAddress.IP.String()] || ips.IsInPrivateNetwork(playerInfo.IP) {
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
