package engine

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"log"
	"time"

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
	DefaultCurve   = elliptic.P256()
	keyring        *ringsig.PublicKeyRing
	privKey        *ecdsa.PrivateKey
	signature      *ringsig.RingSign
	keyRingByte    [][]byte
	signers        [][]byte
	privKeys       []*ecdsa.PrivateKey
	candidates     [][]byte
	SigWitnesses   [][]byte
	electionPubkey = []byte("2_election_12345678")
	sysWallet      *wallet.WalletGroup
	sigCount       = 4
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
	var queryResultCommand = &cobra.Command{
		Use:   "result",
		Short: "Manage election results",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			bc := blockchain.NewBlockchain(getStore(), config.Config{})
			bc = bc.ReInit()
			fmt.Println(bc.QueryResult(electionPubkey))
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
	var castBallot bool
	var getBallot bool

	var electionCommand = &cobra.Command{
		Use:   "election",
		Short: "manage elections",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			bc := blockchain.NewBlockchain(getStore(), config.Config{})
			bc = bc.ReInit()

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
					electionPubkey,
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
					electionPubkey,
					blockchain.TxInput{},
					*txOut,
				)

				eTx.Output.ElectionTx.SigWitnesses = SigWitnesses
				eTx.Output.ElectionTx.Signers = signers

				block, err := bc.AddBlock([]*blockchain.Transaction{eTx})

				if err != nil {
					logger.Error("Add Block Error:", err)
				}
				fmt.Println("Block added  sucessfully: \n", block)
			}

			if stop {
				var electionTx *blockchain.Transaction
				txOut, _ := bc.FindTxWithElectionOutByPubkey(electionPubkey)
				txIn := blockchain.NewElectionTxInput(
					electionPubkey,
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
					electionPubkey,
					*txIn,
					blockchain.TxOutput{},
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
	var accreditationCommand = &cobra.Command{
		Use:   "ac",
		Short: "manage accreditation txs",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			bc := blockchain.NewBlockchain(getStore(), config.Config{})
			bc = bc.ReInit()

			if start {
				var eaTx *blockchain.Transaction

				txOut, _ := bc.FindTxWithElectionOutByPubkey(electionPubkey)
				if txOut.ID == nil {
					logger.Fatal("Error: No correspoding election TxOut")
				}
				txAccreditationOut := blockchain.NewAccreditationTxOutput(
					electionPubkey,
					txOut.ID,
					nil,
					nil,
					time.Now().Unix(),
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
						txAccreditationOut.AccreditationTx.ToByte(),
						w.Main.PublicKey,
						w.Main.PrivateKey,
					)
					privKeys = append(privKeys, &w.Main.PrivateKey)
				}

				eaTx, _ = blockchain.NewTransaction(
					blockchain.ACCREDITATION_TX_TYPE,
					electionPubkey,
					blockchain.TxInput{},
					*txAccreditationOut,
				)

				eaTx.Output.AccreditationTx.SigWitnesses = mu.Sigs
				eaTx.Output.AccreditationTx.Signers = mu.PubKeys

				block, err := bc.AddBlock([]*blockchain.Transaction{eaTx})

				if err != nil {
					logger.Error("Add Block Error:", err)
				}
				fmt.Println("Block added  sucessfully: \n", block)
			}

			if stop {
				var acTx *blockchain.Transaction
				txElectionOut, _ := bc.FindTxWithElectionOutByPubkey(electionPubkey)
				txAcOut, _ := bc.FindTxWithAcOutByPubkey(electionPubkey)
				fmt.Printf("%x", txAcOut.ID)
				// return
				txAcIn := blockchain.NewAccreditationTxInput(
					electionPubkey,
					txElectionOut.ID,
					txAcOut.ID,
					nil,
					nil,
					100,
					time.Now().Unix(),
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
						txAcIn.AccreditationTx.ToByte(),
						w.Main.PublicKey,
						w.Main.PrivateKey,
					)
					privKeys = append(privKeys, &w.Main.PrivateKey)
				}

				acTx, _ = blockchain.NewTransaction(
					blockchain.ACCREDITATION_TX_TYPE,
					electionPubkey,
					*txAcIn,
					blockchain.TxOutput{},
				)

				acTx.Input.AccreditationTx.SigWitnesses = mu.Sigs
				acTx.Input.AccreditationTx.Signers = mu.PubKeys

				block, err := bc.AddBlock([]*blockchain.Transaction{acTx})

				if err != nil {
					logger.Error("Add Block Error:", err)
				}
				fmt.Println("Block added  sucessfully: \n", block)
			}
		},
	}

	var votingCommand = &cobra.Command{
		Use:   "voting",
		Short: "manage voting txs",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			bc := blockchain.NewBlockchain(getStore(), config.Config{})
			bc = bc.ReInit()

			if start {
				var vtTx *blockchain.Transaction

				txOut, _ := bc.FindTxWithElectionOutByPubkey(electionPubkey)
				if txOut.ID == nil {
					logger.Fatal("Error: No correspoding election TxOut")
				}
				txVotingOut := blockchain.NewVotingTxOutput(
					electionPubkey,
					txOut.ID,
					nil,
					nil,
					time.Now().Unix(),
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
						txVotingOut.VotingTx.ToByte(),
						w.Main.PublicKey,
						w.Main.PrivateKey,
					)
					privKeys = append(privKeys, &w.Main.PrivateKey)
				}

				vtTx, _ = blockchain.NewTransaction(
					blockchain.VOTING_TX_TYPE,
					electionPubkey,
					blockchain.TxInput{},
					*txVotingOut,
				)

				vtTx.Output.VotingTx.SigWitnesses = mu.Sigs
				vtTx.Output.VotingTx.Signers = mu.PubKeys

				block, err := bc.AddBlock([]*blockchain.Transaction{vtTx})

				if err != nil {
					logger.Error("Add Block Error:", err)
				}
				fmt.Println("Block added  sucessfully: \n", block)
			}

			if stop {
				var vTx *blockchain.Transaction
				txElectionOut, _ := bc.FindTxWithElectionOutByPubkey(electionPubkey)
				txVotingOut, _ := bc.FindTxWithVotingOutByPubkey(electionPubkey)
				fmt.Printf("%x", txVotingOut.ID)
				// return
				txVotingIn := blockchain.NewVotingTxInput(
					electionPubkey,
					txElectionOut.ID,
					txVotingOut.ID,
					nil,
					nil,
					time.Now().Unix(),
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
						txVotingIn.VotingTx.ToByte(),
						w.Main.PublicKey,
						w.Main.PrivateKey,
					)
					privKeys = append(privKeys, &w.Main.PrivateKey)
				}

				vTx, _ = blockchain.NewTransaction(
					blockchain.VOTING_TX_TYPE,
					electionPubkey,
					*txVotingIn,
					blockchain.TxOutput{},
				)

				vTx.Input.VotingTx.SigWitnesses = mu.Sigs
				vTx.Input.VotingTx.Signers = mu.PubKeys

				block, err := bc.AddBlock([]*blockchain.Transaction{vTx})

				if err != nil {
					logger.Error("Add Block Error:", err)
				}
				fmt.Println("Block added  sucessfully: \n", block)
			}
		},
	}

	var ballotCommand = &cobra.Command{
		Use:   "ballot",
		Short: "manage ballot txs",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			bc := blockchain.NewBlockchain(getStore(), config.Config{})
			bc = bc.ReInit()

			if getBallot {
				logger.Info("Get Ballot!!!!!!!!")
				var bTx *blockchain.Transaction
				secretMessage := []byte("This is my ballot secret message")
				msg, _ := sysWallet.View.Encrypt(secretMessage)
				txElectionOut, err := bc.FindTxWithElectionOutByPubkey(electionPubkey)
				if err != nil {
					logger.Error("Error:", err)
				}

				bTxOut := blockchain.NewBallotTxOutput(
					electionPubkey,
					msg,
					txElectionOut.ID,
					nil,
					nil,
					nil,
					time.Now().Unix(),
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
						bTxOut.BallotTx.ToByte(),
						w.Main.PublicKey,
						w.Main.PrivateKey,
					)
					privKeys = append(privKeys, &w.Main.PrivateKey)
				}

				bTx, _ = blockchain.NewTransaction(
					blockchain.BALLOT_TX_TYPE,
					electionPubkey,
					blockchain.TxInput{},
					*bTxOut,
				)

				bTx.Output.BallotTx.SigWitnesses = mu.Sigs
				bTx.Output.BallotTx.Signers = mu.PubKeys

				// Generate Decoy keys
				for i := 0; i < numOfKeys-1; i++ {
					w := wallet.MakeWalletGroup()
					// add the public key part to the ring
					keyring.Add(w.Main.PrivateKey.PublicKey)
					keyRingByte = append(keyRingByte, w.Main.PublicKey)
				}
				bTx.Output.BallotTx.PubKeys = keyRingByte

				block, err := bc.AddBlock([]*blockchain.Transaction{bTx})

				if err != nil {
					logger.Error("Add Block Error:", err)
				}
				fmt.Println("Block added  sucessfully: \n", block)

			}

			if castBallot {
				logger.Info("Cast Ballot!!!!!!!!")
				var bTx *blockchain.Transaction

				txElectionOut, _ := bc.FindTxWithElectionOutByPubkey(electionPubkey)
				txBallotOut, _ := bc.GetBallotTxByPubkey(electionPubkey)

				bTxIn := blockchain.NewBallotTxInput(
					electionPubkey,
					txElectionOut.Output.ElectionTx.Candidates[0],
					txElectionOut.ID,
					txBallotOut.ID,
					nil,
					nil,
					time.Now().Unix(),
				)

				bTx, _ = blockchain.NewTransaction(
					blockchain.BALLOT_TX_TYPE,
					electionPubkey,
					*bTxIn,
					blockchain.TxOutput{},
				)

				// Sign message
				signature, err := ringsig.Sign(
					&sysWallet.Main.PrivateKey,
					keyring,
					bTxIn.BallotTx.ToByte(),
				)
				if err != nil {
					log.Panic(err)
				}

				bTx.Input.BallotTx.Signature = signature.ToByte()
				bTx.Input.BallotTx.PubKeys = keyRingByte

				block, err := bc.AddBlock([]*blockchain.Transaction{bTx})

				if err != nil {
					logger.Error("Add Block Error:", err)
				}
				fmt.Println("Block added  sucessfully: \n", block)

			}
		},
	}

	electionCommand.Flags().BoolVar(&start, "start", false, "Start Election")
	electionCommand.Flags().BoolVar(&stop, "stop", false, "Stop Election")

	accreditationCommand.Flags().BoolVar(&start, "start", false, "Start Accreditation")
	accreditationCommand.Flags().BoolVar(&stop, "stop", false, "Stop Accreditation")

	votingCommand.Flags().BoolVar(&start, "start", false, "Start Voting")
	votingCommand.Flags().BoolVar(&stop, "stop", false, "Stop Voting")

	ballotCommand.Flags().BoolVar(&getBallot, "get", false, "Get ballot")
	ballotCommand.Flags().BoolVar(&castBallot, "cast", false, "Cast Ballot")

	return []*cobra.Command{
		mainCommand,
		printCommand,
		computeUtxoCommand,
		resetCommand,
		electionCommand,
		accreditationCommand,
		votingCommand,
		ballotCommand,
		queryResultCommand,
		ballotCommand,
	}
}
