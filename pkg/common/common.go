package common

import (
	"net"
)

// var AssigningIPv4 = make(map[uint32]string)

type ContainerIPInfo struct {
	InternalIP uint32
	ExternalIP uint32
	ContainerID string
}


var AssignedIPv4 = make(map[uint32]*ContainerIPInfo)
var ConnectedContainer = make(map[string]*ContainerIPInfo)

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

