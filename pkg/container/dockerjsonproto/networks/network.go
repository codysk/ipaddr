package networks

type Network struct {
	Name string
	Id string
	Driver string
	EnableIPv6 bool
	Containers Containers
	Labels Labels
}

type Networks []Network

type Labels map[string]string
