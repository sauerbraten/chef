package main

import (
	"fmt"
	"strconv"
	"strings"
	"net/url"
	"time"

	"github.com/sauerbraten/chef/db"
)

func nameLookup(nameOrIP string) string {
	sightings := storage.Lookup(nameOrIP, db.ByNameFrequency, false)

	if len(sightings) == 0 {
		return "nothing found!"
	}

	topFiveNames := []string{}

	for i, sighting := range sightings {
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
	sightings := storage.Lookup(nameOrIP, db.ByLastSeen, false)

	if len(sightings) == 0 {
		return "nothing found!"
	}

	serverString := sightings[0].ServerDescription
	if sightings[0].ServerDescription != "" {
		serverString += " ("
	}
	serverString += sightings[0].ServerIP + ":" + strconv.Itoa(sightings[0].ServerPort)
	if sightings[0].ServerDescription != "" {
		serverString += ")"
	}

	return fmt.Sprintf("%s (%s) was last seen %s on %s – more at http://"+conf.WebInterfaceAddress+"/lookup?q=%s&sorting=last_seen", sightings[0].Name, sightings[0].IP, time.Unix(sightings[0].Timestamp, 0).UTC().Format("2006-01-02 15:04:05 MST"), serverString, url.QueryEscape(nameOrIP))
}
