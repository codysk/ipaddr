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

func (manager *Manager) GetConnectedContainers() map[string]*common.ContainerIPInfo {
	return common.ConnectedContainer
}

func (manager *Manager)RevokeAssigning(ipStr string, containerId string) error {
	ip := net.ParseIP(ipStr).To4()
	if ip != nil {
		return manager.revokeAssigningWithIP(ip)
	}

	if containerId != "" {
		return manager.revokeAssigningWithContainerID(containerId)
	}

	return errors.New("invalid parameters")
}

func (manager *Manager) revokeAssigningWithIP(_ip net.IP) error {
	ip := _ip.To4()
	if ip == nil {
		return errors.New("invalid ip format")
	}
	if _, ok := common.AssignedIPv4[common.InetToN(ip)]; !ok {
		return errors.New("error: this ip not assigned")
	}

	externalIP := common.InetToN(ip)
	containerId := common.AssignedIPv4[externalIP].ContainerID
	_ = manager.dockerClient.DisconnectProviderNetwork(containerId)
	delete(common.AssignedIPv4, externalIP)
	delete(common.ConnectedContainer, containerId)

	return nil
}

func (manager *Manager)revokeAssigningWithContainerID(containerId string) error {
	if _, ok := common.ConnectedContainer[containerId]; !ok {
		return errors.New("error: this container not assigned ip")
	}

	externalIP := common.ConnectedContainer[containerId].ExternalIP
	_ = manager.dockerClient.DisconnectProviderNetwork(containerId)

	delete(common.AssignedIPv4, externalIP)
	delete(common.ConnectedContainer, containerId)

	return nil
}

func (manager *Manager) AssignIPForContainer(_ip net.IP, containerId string) error {
	ip := _ip.To4()
	if ip == nil {
		return errors.New("invalid ip format")
	}

	if _, ok := common.AssignedIPv4[common.InetToN(ip)]; ok {
		return errors.New("error: this ip has been assigned")
	}

	if _, ok := common.ConnectedContainer[containerId]; ok {
		err := manager.revokeAssigningWithContainerID(containerId)
		if err != nil {
			return err
		}
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
	containerInfo := &common.ContainerIPInfo{
		ContainerID: containerId,
		InternalIP: common.InetToN(net.ParseIP(containerIP).To4()),
		ExternalIP: common.InetToN(ip),
	}
	common.AssignedIPv4[common.InetToN(ip)] = containerInfo
	common.ConnectedContainer[containerId] = containerInfo

	log.Printf("assinged %v to %v", ip, containerId)

	return nil
}

func NewManager(speaker *arp.Speaker, dockerClient *container.DockerClient) *Manager {
	return &Manager{
		speaker: speaker,
		dockerClient: dockerClient,
	}
}
