package transaction

import (
	"errors"

	. "github.com/elastos/Elastos.ELA.Utility/core/transaction"
	"github.com/elastos/Elastos.ELA.SideChain/core/transaction/payload"
)

const (
	IssueToken              TransactionType = 0x06
	TransferCrossChainAsset TransactionType = 0x08
)

type PayloadFactorySideNodeImpl struct {
	innerFactory *PayloadFactoryImpl
}

func (factor *PayloadFactorySideNodeImpl) Name(txType TransactionType) string {
	if name := factor.innerFactory.Name(txType); name != "Unknown" {
		return name
	}

	switch txType {
	case IssueToken:
		return "IssueToken"
	case TransferCrossChainAsset:
		return "TransferCrossChainAsset"
	default:
		return "Unknown"
	}
}

func (factor *PayloadFactorySideNodeImpl) Create(txType TransactionType) (Payload, error) {
	if p, _ := factor.innerFactory.Create(txType); p != nil {
		return p, nil
	}

	switch txType {
	case IssueToken:
		return new(payload.IssueToken), nil
	case TransferCrossChainAsset:
		return new(payload.TransferCrossChainAsset), nil
	default:
		return nil, errors.New("[NodeTransaction], invalid transaction type.")
	}
}

func init() {
	PayloadFactorySingleton = &PayloadFactorySideNodeImpl{innerFactory: &PayloadFactoryImpl{}}
}