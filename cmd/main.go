package main

import (
	"github.com/spf13/cobra"
	"github.com/workspace/evoting/ev-blockchain-protocol/cmd/engine"
	"github.com/workspace/evoting/ev-blockchain-protocol/cmd/wallet"
)

func main() {
	var app = &cobra.Command{
		Use: "ev",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	engine := engine.NewCommands()
	app.AddCommand(engine...)
	app.AddCommand(wallet.NewCommands())
	app.Execute()
}
