# Chef

A configurable Sauerbraten spy bot written in Go.

Chef collects all name-IP-combinations it finds on servers and stores them in a SQLite database. You can access this data via web interface or IRC bot.

> *With great power comes great responsibility*

## Collector

Chef retrieves the server list from the master server, adds manually specified servers, removes blacklisted servers, and then queries every server for client data via Sauerbraten's extinfo functionality (q.v. [my extinfo package for Go](http://github.com/sauerbraten/extinfo)). This happens periodically at a configurable interval. Chef stores the following data:

- information on the server
- name
- IP
- sighting (i.e. IP bla was seen on server foo using name bar at this specific time)

To store these things, the database has four tables:

- `names` table: stores the name as string
- `ips` table: stores the IP as integer (to enable IP range checks)
- `servers` table: stores server IP as string, server port as int, server description as string. IP and port uniquely identify a server. If the description changes it is simply updated in this table.
- `sightings` table: stores entries consisting of the current time and SQLite rowids referencing a name, an IP and a server.

For more information, see [`db/chef.sqlite.schema`](https://github.com/sauerbraten/chef/blob/master/db/chef.sqlite.schema).


## Web Interface

Chef offers a web interface to access the collected data. The interface uses API-like URLs, but is intended to be used by humans (formatted text replies etc.). It has two endpoints, both of which perform the same DB query, but with different sorting for the results:

- `/names/<name, IP, or IP range>`: sorts by name frequency
- `/lastseen/<name, IP, or IP range>`: sorts by date and time, most recent sighting first

Additionally, there is a `/status` page displaying the number of DB entries per table.


## IRC Bot

Lastly, there is an IRC bot. It performs the same two lookups as the web interface:

- `.names <name, IP, or IP range>`: aliases: `.name`, `.nicks`, `.n`
- `.lastseen <name, IP, or IP range>`: aliases: `.seen`, `.ls`, `.s`


## „Name, IP or IP range“

Chef does some smart regex matching and IP padding to improve the user experience. If the argument to the lookup commands is at least the first octet of an IP and the first dot, it is interpreted as an IP range. The prefix can be ommitted, in which case it will be deduced from the IP. The IP is padded with zeroes if neccessary, i.e. `123.` becomes `123.0.0.0`, `92.1` becomes `92.1.0.0`. The prefix size will be guessed from how many octets were given if it is omitted.

### Examples

- `177.40/16` → `177.40.0.0/16 → results in 177.40.0.0 - 177.40.255.255
- `243.80.97` → `243.80.97.0/24` → results in 243.80.97.0 - 243.80.97.255
- `183.29.64.0/9` → results in 183.0.0.0 - 183.127.255.255