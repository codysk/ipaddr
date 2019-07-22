package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"ipprovider/pkg/addressmanager"
	"ipprovider/pkg/arp"
	"ipprovider/pkg/common"
	"ipprovider/pkg/container"
	"ipprovider/pkg/http"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var RootCmd = &cobra.Command{
	Use: "ipProvider",
	Short: "ipProvider",
	Long: "ipProvider",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("hello ip")
		sigCh := make(chan os.Signal)
		_interface, _ := getFirstBoardcastInterface()
		log.Println("interface: ", _interface.Name)

		log.Println("init test data")
		common.AssignedIPv4[common.InetToN(net.IP{192, 168, 153, 233})] = "test"
		speaker, err := arp.NewArpSpeaker(_interface.Name)
		if err != nil {
			log.Print("get arp speaker failed.")
			log.Fatal(err)
		}

		dockerClient := container.NewDockerClient("/var/run/docker.sock")
		err = dockerClient.InitProviderNetwork()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			speaker.ListenAndServe()
			log.Println("speaker exited")
			sigCh <- syscall.SIGTERM
		}()

		manager := addressmanager.NewManager(speaker)
		go func() {
			log.Print(http.NewHttpServer(":8088", manager, dockerClient).StartHttpServer())
			log.Println("http server exited")
			sigCh <- syscall.SIGTERM
		}()


		log.Println("listening system signal")
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		log.Printf("signal: %v",<-sigCh)

		// PreStop Action
		_ = dockerClient.RemoveProviderNetwork()
	},
}

func getFirstBoardcastInterface() (*net.Interface, error) {
	_interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, _interface := range _interfaces {
		if (_interface.Flags & 0x13) == 0x13 {
			return &_interface, nil
		}
	}
	return nil, errors.New("no available interface")
}