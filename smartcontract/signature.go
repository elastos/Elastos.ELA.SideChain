package smartcontract

import (
	"io"

	"github.com/elastos/Elastos.ELA.SideChain/vm/interfaces"
	"github.com/elastos/Elastos.ELA.SideChain/core"
)

//SignableData describe the data need be signed.
type SignableData interface {
	interfaces.IDataContainer

	SetPrograms([]*core.Program)

	GetPrograms() []*core.Program

	//TODO: add SerializeUnsigned
	SerializeUnsigned(io.Writer) error
}
