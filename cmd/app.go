package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"ipprovider/pkg/arp"
	"ipprovider/pkg/common"
	"ipprovider/pkg/http"
	"log"
	"net"
)

var RootCmd = &cobra.Command{
	Use: "ipProvider",
	Short: "ipProvider",
	Long: "ipProvider",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("hello ip")
		_interface, _ := getFirstBoardcastInterface()
		log.Println("interface: ", _interface.Name)

		log.Println("init test data")
		common.AssignedIPv4[common.InetToN(net.IP{192, 168, 153, 233})] = "test"
		speaker, err := arp.NewArpSpeaker(_interface.Name)
		if err != nil {
			log.Print("get arp speaker failed.")
			log.Fatal(err)
		}
		go speaker.ListenAndServe()

		log.Fatal(http.NewHttpServer(":8088").StartHttpServer())

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