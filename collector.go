package chef

import (
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/sauerbraten/extinfo"

	"github.com/sauerbraten/chef/db"
	"github.com/sauerbraten/chef/ips"
	"github.com/sauerbraten/chef/master"
)

type Collector struct {
	db           *db.Database
	ms           *master.Server
	scanInterval time.Duration
}

func NewCollector(
	db *db.Database,
	masterAddr string,
	scanInterval time.Duration,
) *Collector {
	return &Collector{
		db:           db,
		ms:           master.New(masterAddr, 15*time.Second),
		scanInterval: scanInterval,
	}
}

func (c *Collector) Run() {
	t := time.NewTicker(c.scanInterval)
	for start := time.Now(); true; start = <-t.C {
		slog.Info("refreshing server list from master", "master_addr", c.ms.Address())

		list, err := c.ms.ServerList()
		if err != nil {
			slog.Error("fetch master server list", "error", err)
			continue
		}

		slog.Info("starting scan", "num_servers", len(list))

		wg := new(sync.WaitGroup)
		wg.Add(len(list))
		for _, serverAddress := range list {
			go func() {
				defer wg.Done()
				c.scanServer(serverAddress)
			}()
		}
		wg.Wait()

		slog.Info("completed scan", "num_servers", len(list), "duration", time.Since(start))
	}
}

// scans a server and inserts sightings into the database
func (c *Collector) scanServer(serverAddress *net.UDPAddr) {
	s, err := extinfo.NewServer(*serverAddress, 2*time.Second)
	if err != nil {
		slog.Error("init extinfo for server", "server_addr", serverAddress, "error", err)
		return
	}

	basicInfo, err := s.GetBasicInfo()
	if err != nil {
		slog.Error("get basic info from server", "server_addr", serverAddress, "error", err)
		return
	}

	serverMod, err := s.GetServerMod()
	if err != nil {
		slog.Error("detect server mod of server", "server_addr", serverAddress, "error", err)
		return
	}

	playerInfos, err := s.GetAllClientInfo()
	if err != nil {
		slog.Error("get client info from server", "server_addr", serverAddress, "error", err)
		return
	}

	serverID := c.db.AddOrUpdateServer(serverAddress.IP.String(), serverAddress.Port, basicInfo.Description, serverMod)

	if len(playerInfos) == 0 {
		return
	}

	slog.Info("storing player info from server",
		"num_players", len(playerInfos), "server_desc", basicInfo.Description, "server_addr", serverAddress)

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
