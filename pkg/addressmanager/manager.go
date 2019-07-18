package addressmanager

import (
	"errors"
	"ipprovider/pkg/arp"
	"ipprovider/pkg/common"
	"log"
	"net"
)

type Manager struct {
	speaker *arp.Speaker
	//containerObj
}

func (manager *Manager) AssignIPForContainer(_ip net.IP, containerId string) error {
	ip := _ip.To4()
	if ip == nil {
		return errors.New("invalid ip format")
	}

	err := manager.speaker.AssignIP(ip)
	if err != nil {
		return err
	}

	common.AssignedIPv4[common.InetToN(ip)] = containerId
	log.Printf("assinged %v to %v", ip, containerId)

	return nil
}

func NewManager(speaker *arp.Speaker) *Manager {
	return &Manager{
		speaker: speaker,
	}
}
