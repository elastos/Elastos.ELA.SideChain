package ledger

import (
	"errors"
	"fmt"
	"math"

	tx "github.com/elastos/Elastos.ELA.Core/core/transaction"
	side_tx "github.com/elastos/Elastos.ELA.SideChain/core/transaction"
	core_ledger "github.com/elastos/Elastos.ELA.Core/core/ledger"
	"github.com/elastos/Elastos.ELA.Utility/common"
	"github.com/elastos/Elastos.ELA.SideChain/core/transaction/payload"
)

type SideTransactionValidatorImpl struct {
	core_ledger.TransactionValidatorImpl
}

//validate the transaction of duplicate UTXO input
func (txValiator *SideTransactionValidatorImpl) CheckTransactionInput(txn *tx.NodeTransaction) error {
	var zeroHash common.Uint256
	if txn.IsCoinBaseTx() {
		if len(txn.UTXOInputs) != 1 {
			return errors.New("coinbase must has only one input")
		}
		coinbaseInputHash := txn.UTXOInputs[0].ReferTxID
		coinbaseInputIndex := txn.UTXOInputs[0].ReferTxOutputIndex
		//TODO :check sequence
		if coinbaseInputHash.CompareTo(zeroHash) != 0 || coinbaseInputIndex != math.MaxUint16 {
			return errors.New("invalid coinbase input")
		}

		return nil
	}

	if txValiator.isIssueTokenTx(txn) {
		return nil
	}

	if len(txn.UTXOInputs) <= 0 {
		return errors.New("transaction has no inputs")
	}
	for i, utxoin := range txn.UTXOInputs {
		referTxnHash := utxoin.ReferTxID
		referTxnOutIndex := utxoin.ReferTxOutputIndex
		if (referTxnHash.CompareTo(zeroHash) == 0) && (referTxnOutIndex == math.MaxUint16) {
			return errors.New("invalid transaction input")
		}
		for j := 0; j < i; j++ {
			if referTxnHash == txn.UTXOInputs[j].ReferTxID && referTxnOutIndex == txn.UTXOInputs[j].ReferTxOutputIndex {
				return errors.New("duplicated transaction inputs")
			}
		}
	}

	return nil
}

func (txValiator *SideTransactionValidatorImpl) CheckTransactionSignature(txn *tx.NodeTransaction) error {
	flag, err := VerifySignature(txn)
	if flag && err == nil {
		return nil
	} else {
		return err
	}
}

func (txValiator *SideTransactionValidatorImpl) CheckTransactionUTXOLock(txn *tx.NodeTransaction) error {
	if txn.IsCoinBaseTx() {
		return nil
	}
	if txValiator.isIssueTokenTx(txn) {
		return nil
	}
	if len(txn.UTXOInputs) <= 0 {
		return errors.New("Transaction has no inputs")
	}
	referenceWithUTXO_Output, err := txn.GetReference()
	if err != nil {
		return errors.New(fmt.Sprintf("GetReference failed: %x", txn.Hash()))
	}
	for input, output := range referenceWithUTXO_Output {

		if output.OutputLock == 0 {
			//check next utxo
			continue
		}
		if input.Sequence != math.MaxUint32-1 {
			return errors.New("Invalid input sequence")
		}
		if txn.LockTime < output.OutputLock {
			return errors.New("UTXO output locked")
		}
	}
	return nil
}

func (txValiator *SideTransactionValidatorImpl) CheckTransactionBalance(Tx *tx.NodeTransaction) error {
	if txValiator.isIssueTokenTx(Tx) {
		return nil
	}
	return txValiator.TransactionValidatorImpl.CheckTransactionBalance(Tx)
}

func (txValiator *SideTransactionValidatorImpl) CheckTransactionPayload(Tx *tx.NodeTransaction) error {
	if err := txValiator.TransactionValidatorImpl.CheckTransactionPayload(Tx); err == nil {
		return nil
	}

	switch Tx.Payload.(type) {
	case *payload.IssueToken:
	case *payload.TransferCrossChainAsset:
	default:
		return errors.New("[txValidator],invalidate transaction payload type.")
	}
	return nil
}

func (txValiator *SideTransactionValidatorImpl) isIssueTokenTx(Tx *tx.NodeTransaction) bool {
	return Tx.TxType == side_tx.IssueToken
}

func init() {
	core_ledger.Validator = &SideTransactionValidatorImpl{
		core_ledger.TransactionValidatorImpl{}}
}