package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"ipprovider/pkg/addressmanager"
	"ipprovider/pkg/arp"
	"ipprovider/pkg/container"
	"ipprovider/pkg/http"
	"ipprovider/pkg/iptables"
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
		_interface, _ := getFirstBroadcastInterface()
		log.Println("interface: ", _interface.Name)

		log.Println("init test data")
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

		manager := addressmanager.NewManager(speaker, dockerClient)

		ipManager, err := iptables.NewManager(_interface)
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			speaker.ListenAndServe()
			log.Println("speaker exited")
			sigCh <- syscall.SIGTERM
		}()

		go func() {
			log.Print(http.NewHttpServer(":8088", manager).StartHttpServer())
			log.Println("http server exited")
			sigCh <- syscall.SIGTERM
		}()

		go func() {
			log.Print(ipManager.Serve())
			log.Println("ipManager server exited")
			sigCh <- syscall.SIGTERM
		}()


		log.Println("listening system signal")
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		log.Printf("signal: %v",<-sigCh)

		// PreStop Action
		ipManager.Stop()
		_ = dockerClient.RemoveProviderNetwork()
		ipManager.RemoveChains()
	},
}

func getFirstBroadcastInterface() (*net.Interface, error) {
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