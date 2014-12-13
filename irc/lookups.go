package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sauerbraten/chef/db"
)

func nameLookUp(nameOrIP string) string {
	sightings := storage.LookUp(nameOrIP, db.ByNameFrequency)

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

	return fmt.Sprintf("%s – more at http://"+conf.WebInterfaceAddress+"/names/%s", strings.Join(topFiveNames, ", "), sanitize(nameOrIP))
}

func lastSeenLookUp(nameOrIP string) string {
	sightings := storage.LookUp(nameOrIP, db.ByLastSeen)

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

	return fmt.Sprintf("%s (%s) was last seen %s on %s – more at http://"+conf.WebInterfaceAddress+"/lastseen/%s", sightings[0].Name, sightings[0].IP, time.Unix(sightings[0].Timestamp, 0).UTC().Format("2006-01-02 15:04:05 MST"), serverString, sanitize(nameOrIP))
}
