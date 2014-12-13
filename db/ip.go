package db

import (
	"net"
	"strconv"
	"strings"
)

// assumes 4-byte IPv4
func ipToInt(ip net.IP) (intIp int64) {
	for index, octet := range ip.To4() {
		intIp += int64(octet) << uint((3-index)*8)
	}

	return
}

// assumes 4-byte IPv4
func intToIP(intIp int64) net.IP {
	abcd := [4]byte{}

	for index, _ := range abcd {
		abcd[index] = byte(intIp >> uint((3-index)*8))
	}

	return net.IPv4(abcd[0], abcd[1], abcd[2], abcd[3])
}

// parses all of the following examples into valid IP ranges:
// 123.
// 184.29.39.193/16
// 12.304./8
// 29.43.223./13
// IP octets must be complete and end with a dot. prefix size is optional, a fitting prefix size will be chosen in the case it's omitted.
func getSubnet(cidr string) (ipNet *net.IPNet) {
	parts := strings.Split(cidr, "/")

	ipString := parts[0]
	prefixSize := 0

	if len(parts) == 2 {
		var err error
		prefixSize, err = strconv.Atoi(parts[1])
		if err != nil {
			prefixSize = 0
		} else if prefixSize < 0 || prefixSize > 31 {
			prefixSize = 0
		}
	}

	// not the most elegant way, but meh...
	switch strings.Count(ipString, ".") {
	case 1:
		if strings.HasSuffix(ipString, ".") {
			ipString += "0"
		}
		ipString += ".0.0"

		if prefixSize == 0 {
			prefixSize = 8
		}
	case 2:
		if strings.HasSuffix(ipString, ".") {
			ipString += "0"
		}
		ipString += ".0"

		if prefixSize == 0 {
			prefixSize = 16
		}
	case 3:
		if strings.HasSuffix(ipString, ".") {
			ipString += "0"
		}

		if prefixSize == 0 {
			prefixSize = 24
		}
	}

	// todo: maybe handle error?
	_, ipNet, _ = net.ParseCIDR(ipString + "/" + strconv.Itoa(prefixSize))

	return
}

// assumes IPv4
func getIPRange(ipNet *net.IPNet) (lower, upper int64) {
	var notMask int64

	for index, ipOctet := range ipNet.IP.To4() {
		maskOctet := ipNet.Mask[index]
		notMask += int64(^maskOctet) << uint((3-index)*8)

		octet := ipOctet & maskOctet
		lower += int64(octet) << uint((3-index)*8)
	}

	upper = lower + notMask

	return
}
