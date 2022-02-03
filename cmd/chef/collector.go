package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/sauerbraten/chef/internal/db"
	"github.com/sauerbraten/chef/pkg/extinfo"
	"github.com/sauerbraten/chef/pkg/ips"
	"github.com/sauerbraten/chef/pkg/master"
)

type Collector struct {
	db              *db.Database
	ms              *master.Server
	p               *extinfo.Pinger
	refreshInterval time.Duration
	scanInterval    time.Duration
	serverList      []string
	servers         map[string]*extinfo.Server
	verbose         bool
}

func NewCollector(
	db *db.Database,
	ms *master.Server,
	refreshInterval, scanInterval time.Duration,
	verbose bool,
) *Collector {
	p, err := extinfo.NewPinger("")
	if err != nil {
		panic(err)
	}
	return &Collector{
		db:              db,
		ms:              ms,
		p:               p,
		refreshInterval: refreshInterval,
		scanInterval:    scanInterval,
		serverList:      []string{},
		servers:         map[string]*extinfo.Server{},
		verbose:         verbose,
	}
}

func (c *Collector) Run() {
	c.refreshServerList()
	c.scanServers()

	shouldScan := time.NewTicker(c.scanInterval)
	shouldRefresh := time.NewTicker(c.refreshInterval)

	for {
		select {
		case <-shouldRefresh.C:
			c.refreshServerList()
		case <-shouldScan.C:
			c.scanServers()
		}
	}
}

// Returns the master server list, extended by manually specified extra servers
func (c *Collector) refreshServerList() {
	log.Printf("refreshing server list from %s", c.ms.Address())

	var err error
	c.serverList, err = c.ms.ServerList()
	if err != nil {
		log.Println("collector: error getting master server list:", err)
	}
}

func (c *Collector) scanServers() {
	start := time.Now()

	log.Println("running scan...")

	var wg sync.WaitGroup

	for _, addr := range c.serverList {
		s, ok := c.servers[addr]
		if !ok {
			host, port, err := hostAndPort(addr)
			if err != nil {
				c.logf("parsing %s: %v", addr, err)
				return
			}
			s, _ = extinfo.NewServer(c.p, host, port, 5*time.Second) // c.p is never nil
			c.servers[addr] = s
		}

		wg.Add(1)
		go func(s *extinfo.Server) {
			defer wg.Done()
			// c.logf("scanning %s", addr)
			c.scanServer(s)
			// c.logf("scanning %s completed", addr)
		}(s)
	}

	wg.Wait()

	log.Printf("scan finished (took %v)", time.Since(start))
}

// scans a server and inserts sightings into the database
func (c *Collector) scanServer(s *extinfo.Server) {
	basicInfo, err := s.GetBasicInfo()
	if err != nil {
		c.logf("error getting basic info from %s: %v", s.Addr(), err)
		return
	}

	serverMod, err := s.GetServerMod()
	if err != nil {
		c.logf("error detecting server mod of %s: %v", s.Addr(), err)
		return
	}

	clientInfos, err := s.GetClientInfo(-1)
	if err != nil {
		c.logf("error getting client info from %s: %v", s.Addr(), err)
		return
	}

	hasPlayers := false
	for _, i := range clientInfos {
		if i.State != extinfo.StateSpectator {
			hasPlayers = true
			break
		}
	}

	if !hasPlayers {
		return
	}

	c.logf("found %d clients on %s (%s)", len(clientInfos), basicInfo.Description, s.Addr())

	serverID := c.db.GetServerID(s.Host(), s.Port(), basicInfo.Description, int8(serverMod), basicInfo.ProtocolVersion)
	c.db.UpdateServerLastSeen(serverID)

	var gameID int64
	if basicInfo.GameMode != extinfo.GameModeCoopEdit {
		gameID = c.db.GetGameID(int8(basicInfo.MasterMode), int8(basicInfo.GameMode), basicInfo.Map, serverID, basicInfo.SecsLeft, int(conf.scanInterval/time.Second))
		c.db.UpdateGameLastRecordedAt(gameID)
		if basicInfo.SecsLeft == 0 {
			c.db.SetGameEnded(gameID)
		}

		if basicInfo.GameMode.HasTeams() {
			if t, err := s.GetTeamScores(); err == nil {
				for _, s := range t.Scores {
					c.db.UpdateScore(gameID, s.Name, s.Score)
				}
			} else {
				c.logf("fetching team scores from %s: %v", s.Addr(), err)
			}
		}
	}

	for _, ci := range clientInfos {
		// ignore bots
		if ci.ClientNum >= 128 {
			continue
		}

		combinationID := c.db.GetCombinationID(ci.Name, ci.IP)

		if basicInfo.GameMode != extinfo.GameModeCoopEdit && ci.State != extinfo.StateSpectator {
			// save game stats
			c.db.AddOrUpdateStats(combinationID, gameID, ci.Team, ci.Frags, ci.Deaths, ci.Accuracy, ci.Teamkills, ci.Flags)
		}

		// check for valid client IP
		if serverMod == extinfo.ServerModSpaghetti || ips.IsInReservedBlock(ci.IP) {
			ci.IP = net.ParseIP("0.0.0.0")
		}

		c.db.AddOrIgnoreSighting(combinationID, serverID)
	}
}

func (c *Collector) log(args ...interface{}) {
	if c.verbose {
		log.Println(args...)
	}
}

func (c *Collector) logf(format string, args ...interface{}) {
	if c.verbose {
		log.Printf(format, args...)
	}
}

func hostAndPort(addr string) (string, int, error) {
	host, _port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", -1, fmt.Errorf("parsing '%s' as host:port tuple: %v", addr, err)
	}

	port, err := strconv.Atoi(_port)
	if err != nil {
		return "", -1, fmt.Errorf("error converting port '%s' to int: %v", _port, err)
	}

	return host, port, nil
}
