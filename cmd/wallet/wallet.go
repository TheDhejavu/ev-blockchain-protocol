package wallet

import (
	"github.com/google/uuid"
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/workspace/evoting/ev-blockchain-protocol/wallet"
)

func NewCommands() *cobra.Command {
	var walletCommand = &cobra.Command{
		Use:   "wallet",
		Short: "Manage  wallets",
	}

	var userId string
	var createCommand = &cobra.Command{
		Use:   "create",
		Short: "Create a new wallet locally",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			id, err := uuid.NewUUID()
			if err != nil {
				logger.Panic(err)
			}
			if userId == "" {
				userId = id.String()
			}

			logger.Infof("WALLET ID: %s", userId)

			// Initialize system identity wallet
			wallets, _ := wallet.InitializeWallets()
			// Add new identity to the wallet with the User ID
			wallets.AddWallet(userId)
			wallets.Save()
			w, err := wallets.GetWallet(userId)
			if err != nil {
				logger.Panic(err)
			}
			logger.Info(w.String())
		},
	}

	createCommand.Flags().StringVar(&userId, "user", "", "Unique ID of user")

	walletCommand.AddCommand(
		createCommand,
	)

	return walletCommand
}
