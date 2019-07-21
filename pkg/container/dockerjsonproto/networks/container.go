package networks

type Container struct {
	Name string
	EndpointID string
	MacAddress string
	IPv4Address string
}

type Containers map[string]Container
