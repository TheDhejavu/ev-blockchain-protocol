package main

import (
	"github.com/spf13/cobra"
	"github.com/thedhejavu/ev-blockchain-protocol/cmd/engine"
	"github.com/thedhejavu/ev-blockchain-protocol/cmd/server"
	"github.com/thedhejavu/ev-blockchain-protocol/cmd/wallet"
)

func main() {
	var app = &cobra.Command{
		Use: "ev",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	engine := engine.NewCommands()
	app.AddCommand(engine...)
	app.AddCommand(
		wallet.NewCommands(),
		server.NewCommands(),
	)
	app.Execute()
}
