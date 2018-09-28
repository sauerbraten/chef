package kidban

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sauerbraten/chef/internal/ips"
)

// runs after init() in config.go because of lexical order when files are passed to the compiler
func init() {
	go PeriodicallyUpdateKidbanRanges()
}

var (
	kidbannedNetworks []*net.IPNet
	lock              sync.RWMutex

	timeOfLastUpdate time.Time
)

func GetTimeOfLastUpdate() time.Time {
	return timeOfLastUpdate
}

func IsInKidbannedNetwork(ip net.IP) bool {
	lock.RLock()
	defer lock.RUnlock()
	for _, net := range kidbannedNetworks {
		if net.Contains(ip) {
			return true
		}
	}

	return false
}

func PeriodicallyUpdateKidbanRanges() {
	ticker := time.Tick(conf.UpdateInterval * time.Minute)
	timeOfLastUpdate = time.Now()

	for {
		networks, err := downloadKidbannedNetworks()
		if err != nil {
			log.Println("error fetching kidbanned networks:", err)
			<-ticker // don't set time of last update since this request failed
			continue
		}

		lock.Lock()
		kidbannedNetworks = networks
		lock.Unlock()

		log.Println("updated kidban subnets list")

		timeOfLastUpdate = <-ticker
	}
}

func downloadKidbannedNetworks() (downloadedNetworks []*net.IPNet, err error) {
	resp, err := http.Get(conf.KidbanRangesURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.New("non-200 status code: " + strconv.Itoa(resp.StatusCode))
	}

	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		downloadedNetworks = append(downloadedNetworks, ips.GetSubnet(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	return
}
