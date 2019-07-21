package containers

type Container struct {
	Id string
	Names []string
	Image string
	State string
	NetworkSettings NetworkSettings
}

type Containers []Container