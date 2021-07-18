package server

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/workspace/evoting/ev-blockchain-protocol/rpc"
)

func NewCommands() *cobra.Command {
	var rpcPort string
	var rpcCommand = &cobra.Command{
		Use:   "rpc",
		Short: "Manage RPC Server",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(rpcPort)
			rpc.StartServer(rpcPort)
		},
	}

	rpcCommand.Flags().StringVar(&rpcPort, "port", "4000", "RPC server Port")

	return rpcCommand
}
