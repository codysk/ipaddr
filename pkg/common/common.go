package common

import (
	"log"
	"net"
)

var AssigningIPv4 = make(map[uint32]string)

var AssignedIPv4 = make(map[uint32]string)

func InetToN(ip net.IP) uint32 {
	var ret = uint32(0)
	ret = ( uint32(ip[0]) << 24 ) | ( uint32(ip[1]) << 16 ) | ( uint32(ip[2]) << 8 ) | uint32(ip[3])
	log.Printf("%s -> %v", ip.String(), ret)
	return ret
}

