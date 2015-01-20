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
)

var kidbannedNetworks []*net.IPNet
var lock sync.RWMutex

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

func PeriodicallyUpdateKidbanRanges(url string) {
	ticker := time.Tick(1 * time.Hour)

	for {
		networks, err := downloadKidbannedNetworks(url)
		if err != nil {
			log.Println("error fetching kidbanned networks:", err)
			<-ticker
			continue
		}

		lock.Lock()
		kidbannedNetworks = networks
		lock.Unlock()

		<-ticker
	}
}

func downloadKidbannedNetworks(url string) (downloadedNetworks []*net.IPNet, err error) {
	// hit URL
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.New("non-200 status code: " + strconv.Itoa(resp.StatusCode))
	}

	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		_, network, err := net.ParseCIDR(scanner.Text())
		if err != nil {
			log.Println(err)
			continue
		}

		downloadedNetworks = append(downloadedNetworks, network)
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	return
}
