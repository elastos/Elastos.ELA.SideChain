package service

import (
	"github.com/elastos/Elastos.ELA.SideChain/core"
	"github.com/elastos/Elastos.ELA.SideChain/servers"
	"github.com/elastos/Elastos.ELA.Utility/common"
)

type AssetInfo struct{
	Name 	   string
	Precision  byte
	AssetType  byte
	RecordType byte
}

func GetHeaderInfo(header *core.Header) *servers.BlockHead  {
	h := header.Hash()
	return &servers.BlockHead{
		Version:           header.Version,
		PrevBlockHash:     common.BytesToHexString(common.BytesReverse(header.Previous.Bytes())),
		TransactionsRoot:  common.BytesToHexString(common.BytesReverse(header.MerkleRoot.Bytes())),
		Timestamp:         header.Timestamp,
		Height:            header.Height,
		Nonce:             header.Nonce,
		Hash:              common.BytesToHexString(common.BytesReverse(h.Bytes())),
	}
}

func GetAssetInfo(asset *core.Asset) *AssetInfo {
	return &AssetInfo{
		Name:   	asset.Name,
		Precision:  asset.Precision,
		AssetType:  byte(asset.AssetType),
		RecordType: byte(asset.RecordType),
	}
}
