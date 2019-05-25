package types

import (
	"bytes"
	"io"

	"github.com/elastos/Elastos.ELA.SideChain/auxpow"

	"github.com/elastos/Elastos.ELA/common"
)

type Header struct {
	Version     uint32
	Previous    common.Uint256
	MerkleRoot  common.Uint256
	Timestamp   uint32
	Bits        uint32
	Nonce       uint32
	Height      uint32
	SideAuxPow  auxpow.SideAuxPow
	ReceiptHash common.Uint256
	Bloom       []byte
}

func (header *Header) Serialize(w io.Writer) error {
	err := header.serializeNoAux(w)
	if err != nil {
		return err
	}

	err = header.SideAuxPow.Serialize(w)
	if err != nil {
		return err
	}

	w.Write([]byte{byte(1)})
	return nil
}

func (header *Header) Deserialize(r io.Reader) error {
	err := common.ReadElements(r,
		&header.Version,
		&header.Previous,
		&header.MerkleRoot,
		&header.Timestamp,
		&header.Bits,
		&header.Nonce,
		&header.Height,
		&header.ReceiptHash,
	)
	if err != nil {
		return err
	}

	header.Bloom, err = common.ReadVarBytes(r, 256, "Bloom" )
	if err != nil {
		return err
	}

	// SideAuxPow
	err = header.SideAuxPow.Deserialize(r)
	if err != nil {
		return err
	}

	r.Read(make([]byte, 1))

	return nil
}

func (header *Header) serializeNoAux(w io.Writer) error {
	err := common.WriteElements(w,
		header.Version,
		&header.Previous,
		&header.MerkleRoot,
		header.Timestamp,
		header.Bits,
		header.Nonce,
		header.Height,
		&header.ReceiptHash,
	)
	if err != nil {
		return err
	}
	return common.WriteVarBytes(w, header.Bloom)
	//return nil
}

func (header *Header) Hash() common.Uint256 {
	buf := new(bytes.Buffer)
	header.serializeNoAux(buf)
	return common.Sha256D(buf.Bytes())
}
