package addressmanager

import (
	"errors"
	"ipprovider/pkg/arp"
	"ipprovider/pkg/common"
	"ipprovider/pkg/container"
	"ipprovider/pkg/container/dockerjsonproto/containers"
	"log"
	"net"
	"strings"
)

type Manager struct {
	speaker *arp.Speaker
	dockerClient *container.DockerClient
}

func (manager *Manager) GetContainers() (*containers.Containers, error) {
	return manager.dockerClient.GetContainerList()
}

func (manager *Manager) AssignIPForContainer(_ip net.IP, containerId string) error {
	ip := _ip.To4()
	if ip == nil {
		return errors.New("invalid ip format")
	}

	if _, ok := common.AssignedIPv4[common.InetToN(ip)]; ok {
		return errors.New("error: this ip has been assigned")
	}

	err := manager.speaker.AssignIP(ip)
	if err != nil {
		return err
	}

	providerNetwork, err := manager.dockerClient.InspectProviderNetwork()
	if err != nil {
		return err
	}

	if _, ok := providerNetwork.Containers[containerId]; !ok {
		err = manager.dockerClient.ConnectProviderNetwork(containerId)
		if err != nil {
			return err
		}
		providerNetwork, err = manager.dockerClient.InspectProviderNetwork()
		if err != nil {
			return err
		}
	}

	containerIP := providerNetwork.Containers[containerId].IPv4Address
	containerIP = strings.Split(containerIP, "/")[0]

	log.Printf("%v -> %s[%v]", _ip, containerId, containerIP)
	common.AssignedIPv4[common.InetToN(ip)] = common.InetToN(net.ParseIP(containerIP).To4())

	log.Printf("assinged %v to %v", ip, containerId)

	return nil
}

func NewManager(speaker *arp.Speaker, dockerClient *container.DockerClient) *Manager {
	return &Manager{
		speaker: speaker,
		dockerClient: dockerClient,
	}
}
