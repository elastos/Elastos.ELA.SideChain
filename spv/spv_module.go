package spv

import (
	"bytes"
	"errors"
	"math/rand"

	spvtx "github.com/elastos/Elastos.ELA.SPV/core/transaction"
	"github.com/elastos/Elastos.ELA.SPV/interface"
	spv "github.com/elastos/Elastos.ELA.SPV/interface"
	"github.com/elastos/Elastos.ELA.SideChain/common/config"
	"github.com/elastos/Elastos.ELA.SideChain/core/transaction/payload"
	tx "github.com/elastos/Elastos.ELA/core/transaction"
)

var spvService *_interface.SPVServiceImpl

func SpvInit() error {
	spvService := _interface.NewSPVService(uint64(rand.Int63()), config.SideParameters.SpvSeedList)
	spvService.Start()
	return nil
}

func VerifyTransaction(txn *tx.NodeTransaction) error {
	proof := new(spv.Proof)

	switch object := txn.Payload.(type) {
	case *payload.IssueToken:
		proofBuf := new(bytes.Buffer)
		if err := object.Proof.Serialize(proofBuf); err != nil {
			return errors.New("IssueToken payload serialize failed")
		}
		if err := proof.Deserialize(proofBuf); err != nil {
			return errors.New("IssueToken payload deserialize failed")
		}
	default:
		return errors.New("Invalid payload")
	}

	txBuf := new(bytes.Buffer)
	txn.Serialize(txBuf)

	spvtxn := new(spvtx.Transaction)
	spvtxn.Deserialize(txBuf)

	if err := spvService.VerifyTransaction(*proof, *spvtxn); err != nil {
		return errors.New("SPV module verify transaction failed.")
	}

	return nil
}
