package iptables

import (
	"errors"
	"github.com/coreos/go-iptables/iptables"
	"ipprovider/pkg/common"
	"log"
	"net"
	"time"
)

type Manager struct {
	ipt *iptables.IPTables
	_interface *net.Interface
	stopCh chan int
}

func (manager *Manager) ChainsMaintainer() error {

	err := manager.ipt.AppendUnique(
		common.IPTablesFilterTable,
		"FORWARD",
		"-i",
		manager._interface.Name,
		"-o",
		common.IPProviderNetworkName + "0",
		"-j",
		"ACCEPT",
	)
	if err != nil {
		return err
	}
	
	err = manager.ipt.NewChain(common.IPTablesNatTable, common.IPTablesNatTablePreRouteChain)
	eerr, eok := err.(*iptables.Error)
	switch {
	case err == nil:
	case eok && eerr.ExitStatus() == 1: // chain already exists.
		break
	default:
		return err
	}

	err = manager.ipt.AppendUnique(
		common.IPTablesNatTable,
		common.IPTablesNatTablePreRouteChain,
		"-i",
		"!"+manager._interface.Name,
		"-j",
		"RETURN",
	)
	if err != nil {
		return err
	}

	err = manager.ipt.AppendUnique(
		common.IPTablesNatTable,
		"PREROUTING",
		"-j",
		common.IPTablesNatTablePreRouteChain,
	)

	if err != nil {
		return err
	}

	return nil
}

func (manager *Manager) RemoveChains() {
	err := manager.ipt.Delete(
		common.IPTablesFilterTable,
		"FORWARD",
		"-i",
		manager._interface.Name,
		"-o",
		common.IPProviderNetworkName + "0",
		"-j",
		"ACCEPT",
	)
	if err != nil {
		log.Print(err)
	}
	err = manager.ipt.Delete(
		common.IPTablesNatTable,
		"PREROUTING",
		"-j",
		common.IPTablesNatTablePreRouteChain,
	)
	if err != nil {
		log.Print(err)
	}
	err = manager.ipt.ClearChain(common.IPTablesNatTable, common.IPTablesNatTablePreRouteChain)
	if err != nil {
		log.Print(err)
	}
	err = manager.ipt.DeleteChain(common.IPTablesNatTable, common.IPTablesNatTablePreRouteChain)
	if err != nil {
		log.Print(err)
	}
}

func (manager *Manager) RulesMaintainer() error {
	for externalIP, internalIP := range common.AssignedIPv4 {
		eIP := net.IP{
			byte((externalIP>>24) & 0xff),
			byte((externalIP>>16) & 0xff),
			byte((externalIP>>8) & 0xff),
			byte((externalIP>>0) & 0xff),
		}.To4()
		iIP := net.IP{
			byte((internalIP>>24) & 0xff),
			byte((internalIP>>16) & 0xff),
			byte((internalIP>>8) & 0xff),
			byte((internalIP>>0) & 0xff),
		}.To4()
		err := manager.ipt.AppendUnique(
			common.IPTablesNatTable,
			common.IPTablesNatTablePreRouteChain,
			"-d",
			eIP.String()+"/32",
			"-j",
			"DNAT",
			"--to-destination",
			iIP.String(),
		)
		if err != nil {
			log.Printf("maintainer return err: %v \n eip: %s iip: %s", err, eIP, iIP)
		}

	}
	return nil
}

func (manager *Manager) Serve() error {
	t := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-t.C:
			err := manager.ChainsMaintainer()
			if err != nil {
				log.Printf("chains maintainer err: %s", err)
			}
			err = manager.RulesMaintainer()
			if err != nil {
				log.Printf("rules maintainer err: %s", err)
			}
			break
		case <-manager.stopCh:
			return errors.New("maintainer received stop signal")
		}
	}
}

func (manager *Manager) Stop() {
	manager.stopCh <- 1
}

func NewManager(_int *net.Interface) (*Manager, error) {
	ipt, err := iptables.NewWithProtocol(iptables.ProtocolIPv4)
	if err != nil {
		return nil, err
	}
	return &Manager{
		ipt: ipt,
		_interface: _int,
		stopCh: make(chan int),
	}, nil
}
