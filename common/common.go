package common

import (
	"bytes"
	"io"
	"errors"

	"github.com/elastos/Elastos.ELA.Utility/common"
	"github.com/elastos/Elastos.ELA.Utility/crypto"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"encoding/binary"
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

func ToCodeHash(code []byte) (common.Uint168, error) {
	temp := sha256.Sum256(code)
	md := ripemd160.New()
	io.WriteString(md, string(temp[:]))
	f := md.Sum(nil)

	hash ,err := common.Uint168FromBytes(f)
	if err != nil {
		return common.Uint168{}, errors.New("[Common] , ToCodeHash err.")
	}
	return *hash, nil
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