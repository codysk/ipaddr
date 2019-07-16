package arp

import (
	"github.com/mdlayher/arp"
	"net"
)

type Speaker struct {
	_interface *net.Interface
	arpClient *arp.Client
}

func (speaker *Speaker) recvPacket (ch chan *arp.Packet, errCh chan error) {
	for {
		packet, _, err := speaker.arpClient.Read()
		if err != nil {
			errCh <- err
			break
		}
		ch <- packet
	}
}

func (speaker *Speaker)ListenAndServe() error {
	packetCh := make(chan *arp.Packet)
	errCh := make(chan error)
	go speaker.recvPacket(packetCh, errCh)
	for {
		select {
		case packet := <- packetCh:
			go speaker.packetHandler(packet)
			break
		case err := <- errCh:
			return err
		}
	}
}

func (speaker *Speaker) packetHandler(packet *arp.Packet) {

}

func NewArpSpeaker(_interface string) (*Speaker, error) {
	__interface, err := net.InterfaceByName(_interface)
	if err != nil {
		return nil, err
	}

	client, err := arp.Dial(__interface)
	if err != nil {
		return nil, err
	}

	speaker := &Speaker{
		_interface: __interface,
		arpClient: client,
	}

	return speaker, nil
}