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

	"github.com/sauerbraten/chef/pkg/ips"
)

type Checker struct {
	source          string
	refreshInterval time.Duration

	kidbannedNetworks []*net.IPNet
	timeOfLastUpdate  time.Time

	lock sync.RWMutex
}

func NewChecker(source string, refreshInterval time.Duration) (*Checker, error) {
	c := &Checker{
		source:          source,
		refreshInterval: refreshInterval,
	}

	go c.periodicallyUpdateKidbanRanges()

	return c, nil
}

func (c *Checker) TimeOfLastUpdate() time.Time {
	return c.timeOfLastUpdate
}

func (c *Checker) IsBanned(ip net.IP) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	for _, net := range c.kidbannedNetworks {
		if net.Contains(ip) {
			return true
		}
	}

	return false
}

func (c *Checker) periodicallyUpdateKidbanRanges() {
	ticker := time.Tick(c.refreshInterval)
	c.timeOfLastUpdate = time.Now()

	for {
		networks, err := c.downloadKidbannedNetworks()
		if err != nil {
			log.Println("error fetching kidbanned networks:", err)
			<-ticker // don't set time of last update since this request failed
			continue
		}

		c.lock.Lock()
		c.kidbannedNetworks = networks
		c.lock.Unlock()

		log.Println("updated kidban subnets list")

		c.timeOfLastUpdate = <-ticker
	}
}

func (c *Checker) downloadKidbannedNetworks() (downloadedNetworks []*net.IPNet, err error) {
	resp, err := http.Get(c.source)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.New("non-200 status code: " + strconv.Itoa(resp.StatusCode))
	}

	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		subnet := ips.GetSubnet(scanner.Text())
		if subnet != nil {
			downloadedNetworks = append(downloadedNetworks, subnet)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	return
}
