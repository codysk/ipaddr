package arp

import (
	"errors"
	"github.com/mdlayher/arp"
	"ipprovider/pkg/common"
	"log"
	"net"
	"time"
)

type Speaker struct {
	_interface *net.Interface
	arpClient *arp.Client

	gratuitousArpRespChs map[uint32]chan *arp.Packet

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
			go speaker.assignHandler(packet)
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
		// log.Printf("%v not in %v", ip, common.AssignedIPv4)
		return
	}

	err := speaker.arpClient.Reply(packet, speaker._interface.HardwareAddr, ip)
	if err != nil {
		log.Print(err)
	}

}

func (speaker *Speaker) assignHandler(packet *arp.Packet) {
	for assigningIp, ch := range speaker.gratuitousArpRespChs {
		ip := packet.SenderIP
		mac := packet.SenderHardwareAddr
		if common.InetToN(ip) != assigningIp || mac.String() == speaker._interface.HardwareAddr.String() {
			continue
		}
		select {
		case ch <- packet:
			continue
		default:
		}
	}
}

func (speaker *Speaker) AssignIP(ip net.IP) error {
	freeRecv := make(chan *arp.Packet)

	speaker.gratuitousArpRespChs[common.InetToN(ip)] = freeRecv
	go speaker.sendGratuitousARP(ip)

	select {
	case <- freeRecv:
		delete(speaker.gratuitousArpRespChs, common.InetToN(ip))
		return errors.New("assign failed. ")
	case <- time.After(5*time.Second):
	}

	delete(speaker.gratuitousArpRespChs, common.InetToN(ip))
	return nil
}

func (speaker *Speaker) sendGratuitousARP(ip net.IP)  {
	packet, _ := arp.NewPacket(
		arp.OperationRequest,
		speaker._interface.HardwareAddr,
		ip,
		net.HardwareAddr{0xff,0xff,0xff,0xff,0xff,0xff},
		ip,
	)
	for i := 0; i < 3; i += 1 {
		err := speaker.arpClient.WriteTo(packet, net.HardwareAddr{0xff,0xff,0xff,0xff,0xff,0xff})
		if err != nil {
			log.Println("sendGratuitousARP err: ")
			log.Println(err)
		}
		time.Sleep(1 * time.Second)
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

		gratuitousArpRespChs: make(map[uint32]chan *arp.Packet),
	}

	return speaker, nil
}