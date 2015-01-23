# Chef

A configurable Sauerbraten spy bot written in Go.

Chef collects all name-IP combinations it finds on servers and stores them in an SQLite database. You can access this data via a web interface or an IRC bot.

> *With great power comes great responsibility.*


## Collector

Chef retrieves the server list from the master server, adds manually specified servers, removes blacklisted servers, and then queries every server for client data via Sauerbraten's extinfo functionality (q.v. [my extinfo package for Go](http://github.com/sauerbraten/extinfo)). This happens periodically at a configurable interval. Chef stores the following data:

- information on the server
- name
- IP (only the first three octets since Sauerbraten doesn't give the full IP)
- sighting (i.e. IP bla was seen on server foo using name bar at this specific time)

To store these things, the database has four tables:

- `names`: stores the name as string
- `ips`: stores the IP as integer (to facilitate IP range checks)
- `servers`: stores server IP as string, server port as int, server description as string. An IP and a port together uniquely identify a server. If the description of a known server changes, the description is simply updated in this table.
- `sightings`: stores entries consisting of the current time and SQLite rowids referencing a name, an IP and a server.

For more information, see [`db/chef.sqlite.schema`](https://github.com/sauerbraten/chef/blob/master/db/chef.sqlite.schema).


## Web Interface

Chef offers a web interface to access the collected data. The interface lets you perform two types of lookups, a *direct lookup* and a *2-step lookup*. It has a simple frontpage with a query field, a drop-down to select the sorting (defaults to name frequency) and a checkbox to force a direct lookup when searching with a name. Additionally, there is a status page displaying the number of DB entries per table.

### Direct Lookup

A direct lookup is a simple lookup with only one step. If you give a name, it returns all IPs that have used this name; if you give an IP or IP range, it returns all names that have been used by the given IP (range). Results are displayed as distinct name-IP combinations with the timestamp and server information of their last sighting attached.

### 2-Step Lookup

A 2-step lookup only works on names: It first performs a direct lookup of the name to get all IPs that used the name, and then looks up all sightings by all those IPs. Like direct lookups, it returns distinct name-IP combinations with timestamp and server information of their last sighting.

### ASkidban Integration

The web interface periodically downloads a compiled list of „untrustworthy“ IP ranges, that is IP ranges that were identified by [pisto's ASkidban project](https://github.com/pisto/ASkidban) as ranges that belong to [Autonomous Systems](https://en.wikipedia.org/wiki/Autonomous_System) of hosting companies, businesses, etc., which is where most proxy and VPN servers will be located. Player IPs within one of the „kid“ ranges are colored orange-red on the results page.

The update URL and interval are specified in the configuration file; a sane interval is 60 minutes.

## IRC Bot

Lastly, there is an IRC bot. It has the following two commands:

- `.names <name, IP, or IP range>` (aliases: `.name`, `.nicks`, `.n`)
- `.lastseen <name, IP, or IP range>` (aliases: `.seen`, `.ls`, `.s`)

It uses a *direct lookup* when you give an IP and a *2-step lookup* when you give a name as argument.

## „Name, IP or IP range“

Chef does some smart regex matching and IP padding to improve the user experience. If the query text of a lookup is at least the first octet of an IP and the first dot, it is interpreted as an IP range. The prefix can be ommitted, in which case it will be deduced from the IP. The IP is padded with zeroes if neccessary, i.e. `123.` becomes `123.0.0.0`, `92.1` becomes `92.1.0.0`. The prefix size, if not specified, will be guessed from how many octets were given.

### Examples

- `34.` → `34.0.0.0/8` → results in `34.0.0.0 - 34.255.255.255`
- `177.40/16` → `177.40.0.0/16` → results in `177.40.0.0 – 177.40.255.255`
- `243.80.97` → `243.80.97.0/24` → results in `243.80.97.0 – 243.80.97.255`
- `183.29.64.0/9` → results in `183.0.0.0 – 183.127.255.255`
