package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"net"
)

var RootCmd = &cobra.Command{
	Use: "ipProvider",
	Short: "ipProvider",
	Long: "ipProvider",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hello ip")
		_interface, _ := getFirstBoardcastInterface()
		fmt.Println("interface: ", _interface.Name)
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