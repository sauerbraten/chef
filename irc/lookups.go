package main

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sauerbraten/chef/db"
)

func nameLookup(nameOrIP string) string {
	finishedLookup := storage.Lookup(nameOrIP, db.ByNameFrequency, false)

	if len(finishedLookup.Results) == 0 {
		return "nothing found!"
	}

	topFiveNames := []string{}

	for i, sighting := range finishedLookup.Results {
		if i < 5 {
			alreadyIncluded := false
			for _, includedName := range topFiveNames {
				if includedName == sighting.Name {
					alreadyIncluded = true
				}
			}

			if !alreadyIncluded {
				topFiveNames = append(topFiveNames, sighting.Name)
			}
		} else {
			break
		}
	}

	return fmt.Sprintf("%s – more at http://"+conf.WebInterfaceAddress+"/lookup?q=%s&sorting=name_frequency", strings.Join(topFiveNames, ", "), url.QueryEscape(nameOrIP))
}

func lastSeenLookup(nameOrIP string) string {
	finishedLookup := storage.Lookup(nameOrIP, db.ByLastSeen, false)

	if len(finishedLookup.Results) == 0 {
		return "nothing found!"
	}

	lastSighting := finishedLookup.Results[0]
	serverString := lastSighting.ServerDescription
	if lastSighting.ServerDescription != "" {
		serverString += " ("
	}
	serverString += lastSighting.ServerIP + ":" + strconv.Itoa(lastSighting.ServerPort)
	if lastSighting.ServerDescription != "" {
		serverString += ")"
	}

	return fmt.Sprintf("%s (%s) was last seen %s on %s – more at http://"+conf.WebInterfaceAddress+"/lookup?q=%s&sorting=last_seen", lastSighting.Name, lastSighting.IP, time.Unix(lastSighting.Timestamp, 0).UTC().Format("2006-01-02 15:04:05 MST"), serverString, url.QueryEscape(nameOrIP))
}
