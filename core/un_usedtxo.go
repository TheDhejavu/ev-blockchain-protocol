package blockchain

import (
	"encoding/hex"
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
		key, err := hex.DecodeString(txId)
		if err != nil {
			return err
		}
		key = append(utxoPrefix, key...)
		err = u.chain.crud.Save(key, outs.Serialize())
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *UnusedXTOSet) FindUnUsedTxOuputs(keyHash []byte) []TxOutput {
	var UTXOs []TxOutput

	u.chain.crud.ps.Seek(utxoPrefix, func(k, v []byte) {
		outs := DeSerializeOutputs(v)

		for _, out := range outs.Outputs {
			if out.IsLockWithKeyHash(keyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	})

	return UTXOs
}

func (u *UnusedXTOSet) FindUnUsedBallotOuputs(keyHash []byte) []TxOutput {
	var BallotUTXOs []TxOutput

	u.chain.crud.ps.Seek(utxoPrefix, func(k, v []byte) {
		outs := DeSerializeOutputs(v)

		for _, out := range outs.Outputs {
			if out.IsLockWithKeyHash(keyHash) && out.BallotTx.IsSet() {
				BallotUTXOs = append(BallotUTXOs, out)
			}
		}
	})

	return BallotUTXOs
}
