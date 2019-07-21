package container

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"ipprovider/pkg/common"
	"ipprovider/pkg/container/dockerjsonproto/containers"
	"ipprovider/pkg/container/dockerjsonproto/networks"
	"log"
	"net"
	"net/http"
	"reflect"
)

const ApiVersion = "v1.24"

type DockerClient struct {
	unixSocket string
	host string
	httpClient http.Client
}

func (client *DockerClient)getApiUrl(path string) string {
	return client.host + "/" + ApiVersion + path
}

func (client *DockerClient)GetContainerList() (*containers.Containers, error) {
	resp, err := client.httpClient.Get(client.getApiUrl("/containers/json?all=1"))
	if err != nil {
		// log.Print(err)
		return nil, err
	}
	// log.Printf("resp: %v", resp)
	defer resp.Body.Close()

	var _containers containers.Containers
	err = json.NewDecoder(resp.Body).Decode(&_containers)
	if err != nil {
		log.Printf("error on decode resp body: %v", err)
		return nil, err
	}

	return &_containers, err

}

func (client *DockerClient) InitProviderNetwork() error {
	// Todo: delete provider network then create new one
	err := client.RemoveProviderNetwork()
	if err != nil {
		return err
	}

	err = client.CreateProviderNetwork()
	if err != nil {
		return err
	}
	return nil
}

func (client *DockerClient) CreateProviderNetwork() error {
	log.Println("creating Network " + common.IPProviderNetworkName)
	type ReqIPAM struct {
		Driver string
		Config []map[string]string
		Options map[string]string
	}
	type CreateReq struct {
		Name string
		CheckDuplicate bool
		Driver string
		EnableIPv6 bool
		IPAM *ReqIPAM
		Internal bool
		Options map[string]string
		Labels map[string]string
	}
	createReq := &CreateReq{
		Name: common.IPProviderNetworkName,
		CheckDuplicate: true,
		Driver: "bridge",
		EnableIPv6: false,
		IPAM: &ReqIPAM{
			Driver: "default",
			Config: []map[string]string{
				{
					"Subnet": common.Subnet,
					"IPRange": common.IPRange,
					"Gateway": common.Gateway,
				},
			},
		},
		Internal: false,
		Options: map[string]string{
			"com.docker.network.bridge.name": common.IPProviderNetworkName+"0",
			"com.docker.network.bridge.enable_ip_masquerade": "true",
			"com.docker.network.bridge.enable_icc": "true",
			"com.docker.network.bridge.host_binding_ipv4": "0.0.0.0",
			"com.docker.network.driver.mtu": "1500",
		},
		Labels: map[string]string{
			"manage-by": reflect.TypeOf(DockerClient{}).PkgPath(),
		},
	}

	reqBodyBuf := new(bytes.Buffer)
	err := json.NewEncoder(reqBodyBuf).Encode(createReq)
	if err != nil {
		return err
	}
	reqString := reqBodyBuf.String()

	resp, err := client.httpClient.Post(
		client.getApiUrl("/networks/create"),
		"application/json",
		reqBodyBuf,
	)
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		return errors.New(
			fmt.Sprintf(
				"create network request return code: %d \n"+
					"request: %v",
				resp.StatusCode,
				reqString,
			),
		)
	}
	return nil
}

func (client *DockerClient) RemoveProviderNetwork() error {
	network, err := client.InspectProviderNetwork()
	if err != nil {
		return err
	}

	_containers := network.Containers

	for id := range _containers {
		err := client.DisconnectProviderNetwork(id)
		if err != nil {
			log.Print(err)
		}
	}

	req, _ := http.NewRequest("DELETE",
		client.getApiUrl("/networks/" + common.IPProviderNetworkName),
		nil,
	)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		log.Printf("remove network request return code: %d", resp.StatusCode)
	}

	log.Println("Network " + common.IPProviderNetworkName + " removed")
	return nil
}

func (client *DockerClient) DisconnectProviderNetwork(containerId string) error {
	type DisconnectReq struct {
		Container string
		Force bool
	}
	disconnectReq := &DisconnectReq{
		Container: containerId,
		Force: true,
	}

	reqBodyBuf := new(bytes.Buffer)
	err := json.NewEncoder(reqBodyBuf).Encode(disconnectReq)
	if err != nil {
		return err
	}
	resp, err := client.httpClient.Post(
		client.getApiUrl("/networks/" + common.IPProviderNetworkName + "/disconnect"),
		"application/json",
		reqBodyBuf,
	)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(
			fmt.Sprintf(
				"disconnect container [%s] return status code: %d",
				containerId,
				resp.StatusCode,
			),
		)
	}
	return nil
}

func (client *DockerClient) InspectProviderNetwork() (*networks.Network, error) {
	resp, err := client.httpClient.Get(client.getApiUrl("/networks/" + common.IPProviderNetworkName))
	if err != nil {
		// log.Print(err)
		return nil, err
	}
	defer resp.Body.Close()

	var network networks.Network

	err = json.NewDecoder(resp.Body).Decode(&network)
	if err != nil {
		log.Printf("error on decode resp body: %v", err)
		return nil, err
	}

	return &network, nil
}

func NewDockerClient(socketPath string) *DockerClient {
	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (conn net.Conn, e error) {
				return net.Dial("unix", socketPath)
			},
		},
	}
	return &DockerClient{
		unixSocket: socketPath,
		httpClient: client,
		host: "http://unix",
	}
}
