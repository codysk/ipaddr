package arp

import (
	"github.com/mdlayher/arp"
	"ipprovider/pkg/common"
	"log"
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
	log.Println("arp speaker starting...")
	packetCh := make(chan *arp.Packet)
	errCh := make(chan error)
	log.Println("arp speaker started")
	go speaker.recvPacket(packetCh, errCh)
	for {
		select {
		case packet := <- packetCh:
			go speaker.packetHandler(packet)
			break
		case err := <- errCh:
			log.Fatal("speaker down!")
			return err
		}
	}
}

func (speaker *Speaker) packetHandler(packet *arp.Packet) {

	if packet.Operation == arp.OperationReply {
		return
	}

	ip := packet.TargetIP.To4()
	if ip == nil {
		return
	}

	_, exist := common.AssignedIPv4[common.InetToN(ip)]
	if !exist {
		return
	}

	err := speaker.arpClient.Reply(packet, speaker._interface.HardwareAddr, ip)
	if err != nil {
		log.Print(err)
	}

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