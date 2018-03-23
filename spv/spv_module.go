package spv

import (
	tx "Elastos.ELA.SideChain/core/transaction"
	spvTx "SPVWallet/core/transaction"
	"SPVWallet/interface"
	"SPVWallet/p2p/msg"
	"bytes"
	"errors"
	"math/rand"
)

var spvService *_interface.SPVServiceImpl

func SpvInit() error {
	spvService, err := _interface.NewSPVService(uint64(rand.Int63()))
	if err != nil {
		return errors.New("[Error] " + err.Error())
	}
	spvService.Start()
	return nil
}

func VerifyTransaction(spvInfo []byte, txn *tx.Transaction) error {
	buf := new(bytes.Buffer)
	txn.Serialize(buf)
	txBytes := buf.Bytes()

	r := bytes.NewReader(txBytes)
	spvTxn := new(spvTx.Transaction)
	spvTxn.Deserialize(r)

	merkleBlock := new(msg.MerkleBlock)
	merkleBlock.Deserialize(spvInfo)
	if err := spvService.VerifyTransaction(*merkleBlock, []spvTx.Transaction{*spvTxn}); err != nil {
		return errors.New("SPV module verify transaction failed.")
	}

	return nil
}
