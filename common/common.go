package common

import (
	"bytes"
	"encoding/binary"

	"github.com/elastos/Elastos.ELA.Utility/common"
	"github.com/elastos/Elastos.ELA.Utility/crypto"
)

func GetGenesisAddress(genesisHash common.Uint256) (string, error) {
	programHash, err := GetGenesisProgramHash(genesisHash)
	if err != nil {
		return "", err
	}
	return programHash.ToAddress()
}

func GetGenesisProgramHash(genesisHash common.Uint256) (*common.Uint168, error) {
	buf := new(bytes.Buffer)
	buf.WriteByte(byte(len(genesisHash.Bytes())))
	buf.Write(genesisHash.Bytes())
	buf.WriteByte(byte(common.CROSSCHAIN))

	return crypto.ToProgramHash(buf.Bytes())
}

func IntToBytes(n int) []byte {
	tmp := int32(n)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.LittleEndian, tmp)
	return buffer.Bytes()
}

func BytesToInt(b []byte) []int {
	i := make([]int, len(b))
	for k, v := range b {
		i[k] = int(v)
	}
	return i
}