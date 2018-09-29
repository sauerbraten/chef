package main

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/sauerbraten/extinfo"

	"github.com/sauerbraten/chef/internal/db"
	"github.com/sauerbraten/chef/internal/ips"
)

var storage *db.Database

func main() {
	var err error
	storage, err = db.New(conf.DatabaseFilePath)
	if err != nil {
		log.Fatalln("could not initialize database:", err)
	}

	scanInterval, err := time.ParseDuration(conf.ScanInterval)
	if err != nil {
		log.Fatalln("could not parse scan interval:", err)
	}

	ms := newMasterServer(conf.MasterServerAddress)

	t := time.NewTicker(scanInterval)
	for start := time.Now(); true; start = <-t.C {
		log.Println("refreshing server list after tick at", start.String())

		resolveConfigServers()
		list := getExtendedServerList(ms)

		log.Println("running scan...")

		var wg sync.WaitGroup

		for _, serverAddress := range list {
			wg.Add(1)
			go func(serverAddress *net.UDPAddr) {
				defer wg.Done()
				scanServer(serverAddress)
			}(serverAddress)
		}

		wg.Wait()

		log.Printf("scan finished (took %v)", time.Since(start))
	}
}

func resolveConfigServers() {
	// clear old server lists
	conf.extraServers = make([]*net.UDPAddr, 0, len(conf.ExtraServers))

	// resolve extra servers
	for _, serverAddress := range conf.ExtraServers {
		addr, err := net.ResolveUDPAddr("udp", serverAddress)
		if err != nil {
			log.Println("error resolving "+serverAddress+":", err)
			continue
		}

		conf.extraServers = append(conf.extraServers, addr)
	}
}

// Returns the master server list, extended by manually specified extra servers
func getExtendedServerList(ms *masterServer) (list map[string]*net.UDPAddr) {
	var err error

	list, err = ms.getServerList()
	if err != nil {
		log.Println("error getting master server list:", err)
	}

	// make sure to have a non-nil map (list can be nil if the master server could not be reached, for example)
	if list == nil {
		list = map[string]*net.UDPAddr{}
	}

	for _, addr := range conf.extraServers {
		list[addr.String()] = addr
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

	serverMod, err := s.GetServerMod()
	if err != nil {
		verbose("error detecting server mod of", serverAddress, ":", err)
		return
	}

	playerInfos, err := s.GetAllClientInfo()
	if err != nil {
		verbose("error getting client info from", serverAddress, ":", err)
		return
	}

	if len(playerInfos) == 0 {
		return
	}

	verbose("found", len(playerInfos), "players on", basicInfo.Description, serverAddress.String())

	serverID := storage.GetServerID(serverAddress.IP.String(), serverAddress.Port, basicInfo.Description)

	for _, playerInfo := range playerInfos {
		// don't save bot sightings
		if playerInfo.ClientNum > extinfo.MaxPlayerCN {
			continue
		}

		// check for valid client IP
		if serverMod == "spaghettimod" || ips.IsInReservedBlock(playerInfo.IP) {
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
