package main

import "strings"

type User struct {
	Name    string   `json:"username"`
	Aliases []string `json:"aliases"`
	Host    string   `json:"host"`
}

func (u User) hasAlias(alias string) bool {
	for _, a := range u.Aliases {
		if a == alias {
			return true
		}
	}

	return false
}

func initializeUsers() {
	conf.usernameByHostname = map[string]string{}
	conf.hostnameByUsername = map[string]string{}

	for _, user := range conf.TrustedUsers {
		conf.usernameByHostname[user.Host] = user.Name
		conf.hostnameByUsername[strings.ToLower(user.Name)] = user.Host
	}
}

func getHostnameByAlias(alias string) (hostname string, ok bool) {
	for _, user := range conf.TrustedUsers {
		if user.hasAlias(alias) {
			hostname, ok = user.Host, true
		}
	}

	return
}

func getHostnameByUsername(username string) (hostname string, ok bool) {
	hostname, ok = conf.hostnameByUsername[username]
	return
}

func getUsernameByHostname(hostname string) (username string, ok bool) {
	mask := getMaskFromHostname(hostname)
	username, ok = conf.usernameByHostname[mask]
	return
}

func getMaskFromHostname(hostname string) string {
	parts := strings.Split(hostname, ".")
	return parts[0] + ".*." + parts[len(parts)-1]
}
