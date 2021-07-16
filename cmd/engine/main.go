package engine

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"

	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	blockchain "github.com/workspace/evoting/ev-blockchain-protocol/core"
	"github.com/workspace/evoting/ev-blockchain-protocol/database"
	"github.com/workspace/evoting/ev-blockchain-protocol/pkg/config"
	"github.com/workspace/evoting/ev-blockchain-protocol/pkg/crypto/multisig"
	"github.com/workspace/evoting/ev-blockchain-protocol/pkg/crypto/ringsig"
	"github.com/workspace/evoting/ev-blockchain-protocol/wallet"
)

const numOfKeys = 3

var (
	DefaultCurve = elliptic.P256()
	keyring      *ringsig.PublicKeyRing
	privKey      *ecdsa.PrivateKey
	signature    *ringsig.RingSign
	keyRingByte  [][]byte
	signers      [][]byte
	privKeys     []*ecdsa.PrivateKey
	candidates   [][]byte
	SigWitnesses [][]byte
	keyHash      = []byte("election_x")
	sysWallet    *wallet.WalletGroup
	sigCount     = 4
)

func GenerateMainWallet() {
	// Main Key
	keyring = ringsig.NewPublicKeyRing(numOfKeys)
	sysWallet = wallet.MakeWalletGroup()
	keyring.Add(sysWallet.Main.PrivateKey.PublicKey)
	keyRingByte = append(keyRingByte, sysWallet.Main.PublicKey)
}

func getStore() database.Store {
	store, err := database.NewStore("badgerdb", "4000")
	if err != nil {
		logger.Panic(err)
	}
	return store
}
func NewCommands() []*cobra.Command {
	GenerateMainWallet()

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

	var resetCommand = &cobra.Command{
		Use:   "reset",
		Short: "Reset Blockchain ",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			bc := blockchain.NewBlockchain(getStore(), config.Config{})
			bc = bc.ReInit()
			bc.ResetBlockchain("4000")

		},
	}

	var computeUtxoCommand = &cobra.Command{
		Use:   "utxo",
		Short: "Compute UTXO",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			bc := blockchain.NewBlockchain(getStore(), config.Config{})
			bc = bc.ReInit()
			bc.ComputeUnUsedTXOs()
		},
	}
	var start bool
	var stop bool

	var newElectionCommand = &cobra.Command{
		Use:   "election",
		Short: "manage elections",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			bc := blockchain.NewBlockchain(getStore(), config.Config{})
			bc = bc.ReInit()
			utxo := blockchain.NewUnusedXTOSet(bc)
			if start {
				var eTx *blockchain.Transaction

				var totalPeople int64
				totalPeople = 100
				for i := 0; i < 4; i++ {
					w := wallet.MakeWalletGroup()
					candidates = append(candidates, w.Main.PublicKey)
				}

				txOut := blockchain.NewElectionTxOutput(
					"Presidential Election",
					"President",
					keyHash,
					nil,
					nil,
					candidates,
					totalPeople,
				)

				mu := multisig.NewMultisig(sigCount)
				for i := 0; i < sigCount; i++ {
					// Initialize system identity wallet
					wallets, _ := wallet.InitializeWallets()
					// Add new identity to the wallet with the User ID
					userId := wallets.AddWallet(fmt.Sprintf("signers_%d", i))
					wallets.Save()
					w, err := wallets.GetWallet(userId)
					if err != nil {
						logger.Panic(err)
					}
					mu.AddSignature(
						txOut.ElectionTx.ToByte(),
						w.Main.PublicKey,
						w.Main.PrivateKey,
					)

					privKeys = append(privKeys, &w.Main.PrivateKey)
				}

				SigWitnesses = mu.Sigs
				signers = mu.PubKeys

				eTx, _ = blockchain.NewTransaction(
					blockchain.ELECTION_TX_TYPE,
					keyHash,
					blockchain.TxInput{},
					*txOut,
					utxo,
				)

				eTx.Output.ElectionTx.SigWitnesses = SigWitnesses
				eTx.Output.ElectionTx.Signers = signers

				block, err := bc.AddBlock([]*blockchain.Transaction{eTx})

				if err != nil {
					logger.Error("Add Block Error:", err)
				}
				fmt.Println("Block added  sucessfully: \n", block)
			} else {
				var electionTx *blockchain.Transaction
				txOut := bc.GetTransactionByKeyHash(keyHash)
				txIn := blockchain.NewElectionTxInput(
					keyHash,
					txOut.ID,
					signers,
					SigWitnesses,
				)
				mu := multisig.NewMultisig(sigCount)
				for i := 0; i < sigCount; i++ {
					// Initialize system identity wallet
					wallets, _ := wallet.InitializeWallets()
					userId := fmt.Sprintf("signers_%d", i)
					w, err := wallets.GetWallet(userId)
					if err != nil {
						logger.Panic(err)
					}
					mu.AddSignature(
						txIn.ElectionTx.ToByte(),
						w.Main.PublicKey,
						w.Main.PrivateKey,
					)
					privKeys = append(privKeys, &w.Main.PrivateKey)
				}

				electionTx, _ = blockchain.NewTransaction(
					blockchain.ELECTION_TX_TYPE,
					keyHash,
					*txIn,
					blockchain.TxOutput{},
					utxo,
				)

				electionTx.Input.ElectionTx.SigWitnesses = mu.Sigs
				electionTx.Input.ElectionTx.Signers = mu.PubKeys
				block, err := bc.AddBlock([]*blockchain.Transaction{electionTx})

				if err != nil {
					logger.Error("Add Block Error:", err)
				}
				fmt.Println("Block added  sucessfully: \n", block)
			}
		},
	}

	newElectionCommand.Flags().BoolVar(&start, "start", false, "Start Election")
	newElectionCommand.Flags().BoolVar(&stop, "stop", false, "Stop Election")

	return []*cobra.Command{
		mainCommand,
		printCommand,
		computeUtxoCommand,
		resetCommand,
		newElectionCommand,
	}
}
