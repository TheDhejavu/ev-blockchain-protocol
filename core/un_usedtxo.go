package blockchain

import (
	"encoding/hex"
	"fmt"
)

var (
	utxoPrefix  = []byte("utxo-")
	prefiLength = len(utxoPrefix)
)

type UnusedXTOSet struct {
	chain *Blockchain
}

func NewUnusedXTOSet(chain *Blockchain) *UnusedXTOSet {
	return &UnusedXTOSet{chain}
}
func (u *UnusedXTOSet) CountUnusedTxOutputs() int {
	counter := u.chain.crud.CountByPrefix(utxoPrefix)

	return counter
}

func (u *UnusedXTOSet) Compute() error {
	u.chain.crud.DeleteByPrefix(utxoPrefix)

	UTXO, err := u.chain.FindUnUsedTXO()
	if err != nil {
		return err
	}
	for txId, outs := range UTXO {
		fmt.Printf("TX_ID: %x \n", txId)
		key, err := hex.DecodeString(txId)
		if err != nil {
			return err
		}

		key = append(utxoPrefix, key...)
		err = u.chain.crud.Save(key, outs.Serialize())
		for i := 0; i < len(outs.Outputs); i++ {
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *UnusedXTOSet) FindUnUsedAccreditationTxOuputs(pubKey []byte) map[string]TxOutput {
	var utxos = make(map[string]TxOutput)

	u.chain.crud.ps.Seek(utxoPrefix, func(k, v []byte) {
		outs := DeSerializeOutputs(v)
		txId := hex.EncodeToString(k[len(utxoPrefix):])

		for _, out := range outs.Outputs {
			if out.IsLockWithKeyHash(pubKey) && out.AccreditationTx.IsSet() {
				utxos[txId] = out
			}
		}
	})

	return utxos
}

func (u *UnusedXTOSet) FindUnUsedVotingTxOuputs(pubKey []byte) map[string]TxOutput {
	var utxos = make(map[string]TxOutput)

	u.chain.crud.ps.Seek(utxoPrefix, func(k, v []byte) {
		outs := DeSerializeOutputs(v)
		txId := hex.EncodeToString(k[len(utxoPrefix):])

		for _, out := range outs.Outputs {
			if out.IsLockWithKeyHash(pubKey) && out.VotingTx.IsSet() {
				utxos[txId] = out
			}
		}
	})

	return utxos
}

func (u *UnusedXTOSet) FindUnUsedBallotTxOuputs(pubKey []byte) map[string]TxOutput {
	var utxos = make(map[string]TxOutput)

	u.chain.crud.ps.Seek(utxoPrefix, func(k, v []byte) {
		outs := DeSerializeOutputs(v)
		txId := hex.EncodeToString(k[len(utxoPrefix):])

		for _, out := range outs.Outputs {
			if out.IsLockWithKeyHash(pubKey) && out.BallotTx.IsSet() {
				utxos[txId] = out
			}
		}
	})

	return utxos
}

func (u *UnusedXTOSet) FindUnUsedElectionTxOuputs(pubKey []byte) map[string]TxOutput {
	var utxos = make(map[string]TxOutput)

	u.chain.crud.ps.Seek(utxoPrefix, func(k, v []byte) {
		outs := DeSerializeOutputs(v)
		txId := hex.EncodeToString(k[len(utxoPrefix):])

		for _, out := range outs.Outputs {
			if out.IsLockWithKeyHash(pubKey) && out.ElectionTx.IsSet() {
				utxos[txId] = out
			}
		}
	})

	return utxos
}
