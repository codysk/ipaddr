package common

import (
	"net"
)

// var AssigningIPv4 = make(map[uint32]string)

var AssignedIPv4 = make(map[uint32]uint32)

var Subnet = "10.50.0.0/16"
var IPRange = "10.50.0.0/17"
var Gateway = "10.50.255.254"

const IPProviderNetworkName = "provider_net"

const IPTablesNatTable = "nat"
const IPTablesFilterTable = "filter"
const IPTablesNatTablePreRouteChain = "IPPROVIDER_PREROUTE"

func InetToN(ip net.IP) uint32 {
	var ret = uint32(0)
	ret = ( uint32(ip[0]) << 24 ) | ( uint32(ip[1]) << 16 ) | ( uint32(ip[2]) << 8 ) | uint32(ip[3])
	// log.Printf("%s -> %v", ip.String(), ret)
	return ret
}

