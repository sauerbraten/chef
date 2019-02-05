package main

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/sauerbraten/extinfo"

	"github.com/sauerbraten/chef/internal/db"
	"github.com/sauerbraten/chef/pkg/ips"
	"github.com/sauerbraten/chef/pkg/master"
)

type Collector struct {
	db           *db.Database
	ms           *master.Server
	scanInterval time.Duration
	extraServers []string
	verbose      bool
}

func NewCollector(
	db *db.Database,
	ms *master.Server,
	scanInterval time.Duration,
	extraServers []string,
	verbose bool,
) *Collector {
	return &Collector{
		db:           db,
		ms:           ms,
		scanInterval: scanInterval,
		extraServers: extraServers,
		verbose:      verbose,
	}
}

func (c *Collector) Run() {
	t := time.NewTicker(c.scanInterval)
	for start := time.Now(); true; start = <-t.C {
		log.Println("refreshing server list after tick at", start.String())

		list := c.fetchExtendedServerList()

		log.Println("running scan...")

		var wg sync.WaitGroup

		for _, serverAddress := range list {
			wg.Add(1)
			go func(serverAddress *net.UDPAddr) {
				defer wg.Done()
				c.scanServer(serverAddress)
			}(serverAddress)
		}

		wg.Wait()

		log.Printf("scan finished (took %v)", time.Since(start))
	}
}

// Returns the master server list, extended by manually specified extra servers
func (c *Collector) fetchExtendedServerList() (list map[string]*net.UDPAddr) {
	var err error

	list, err = c.ms.ServerList()
	if err != nil {
		log.Println("collector: error getting master server list:", err)
	}

	// make sure to have a non-nil map (list can be nil if the master server could not be reached, for example)
	if list == nil {
		list = map[string]*net.UDPAddr{}
	}

	for _, _addr := range c.extraServers {
		if _, ok := list[_addr]; ok {
			// server from config was already on the master list
			continue
		}

		addr, err := net.ResolveUDPAddr("udp", _addr)
		if err != nil {
			log.Println("collector: error resolving "+_addr+":", err)
			continue
		}

		list[_addr] = addr
	}

	return
}

// scans a server and inserts sightings into the database
func (c *Collector) scanServer(serverAddress *net.UDPAddr) {
	s, err := extinfo.NewServer(*serverAddress, 2*time.Second)
	if err != nil {
		c.log(err)
		return
	}

	basicInfo, err := s.GetBasicInfo()
	if err != nil {
		c.log("error getting basic info from", serverAddress, ":", err)
		return
	}

	serverMod, err := s.GetServerMod()
	if err != nil {
		c.log("error detecting server mod of", serverAddress, ":", err)
		return
	}

	playerInfos, err := s.GetAllClientInfo()
	if err != nil {
		c.log("error getting client info from", serverAddress, ":", err)
		return
	}

	if len(playerInfos) == 0 {
		return
	}

	c.log("found", len(playerInfos), "players on", basicInfo.Description, serverAddress.String())

	serverID := c.db.GetServerID(serverAddress.IP.String(), serverAddress.Port, basicInfo.Description)
	c.db.UpdateServerLastSeen(serverID)

	for _, playerInfo := range playerInfos {
		// don't save bot sightings
		if playerInfo.ClientNum > extinfo.MaxPlayerCN {
			continue
		}

		// check for valid client IP
		if serverMod == "spaghettimod" || ips.IsInReservedBlock(playerInfo.IP) {
			playerInfo.IP = net.ParseIP("0.0.0.0")
		}

		c.db.AddOrIgnoreSighting(playerInfo.Name, playerInfo.IP, serverID)
	}
}

func (c *Collector) log(args ...interface{}) {
	if c.verbose {
		log.Println(args...)
	}
}
