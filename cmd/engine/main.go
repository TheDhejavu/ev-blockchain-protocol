package engine

import (
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	blockchain "github.com/workspace/evoting/ev-blockchain-protocol/core"
	"github.com/workspace/evoting/ev-blockchain-protocol/database"
	"github.com/workspace/evoting/ev-blockchain-protocol/pkg/config"
)

func getStore() database.Store {
	store, err := database.NewStore("badgerdb", "4000")
	if err != nil {
		logger.Panic(err)
	}
	return store
}
func NewCommands() []*cobra.Command {

	var mainCommand = &cobra.Command{
		Use:   "init",
		Short: "Initialize the blockchain and create the genesis block",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			bc := blockchain.NewBlockchain(getStore(), config.Config{})
			bc.Init()
		},
	}

	var printCommand = &cobra.Command{
		Use:   "print",
		Short: "Print the blockchain data",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			bc := blockchain.NewBlockchain(getStore(), config.Config{})
			bc = bc.ReInit()
			bc.PrintBlockchain()
		},
	}

	return []*cobra.Command{
		mainCommand,
		printCommand,
	}
}
