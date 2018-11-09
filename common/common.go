package common

import (
	"bytes"
	"encoding/binary"

	"github.com/elastos/Elastos.ELA.Utility/common"
	"github.com/elastos/Elastos.ELA.Utility/crypto"

	"github.com/elastos/Elastos.ELA.SideChain/vm"
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

func UInt168ToUInt160(hash *common.Uint168) []byte {
	hashBytes := make([]byte, len(hash) - 1)
	data := hash.Bytes();
	copy(hashBytes, data[1 : len(hash)])
	return hashBytes
}

func IsMultiSigContract(script []byte) bool {
	m := 0
	n := 0
	i := 0
	if len(script) < 37 {
		return false
	}
	if script[i] > vm.PUSH16 {
		return false;
	}

	if script[i] < vm.PUSH1 && script[i] != 1 && script[i] != 2 {
		return false
	}
	switch script[i] {
	case 1:
		i++
		m = int(script[i])
		i++
	case 2:
		i++
		m = int(script[i])
		i += 2
	default:
		m = int(script[i]) - 80
		i++
	}
	if m < 1 || m > 1024 {
		return false
	}
	for script[i] == 33 {
		i += 34
		if len(script) <= i {
			return false
		}
		n++
	}
	if n < m || n > 1024 {
		return false
	}
	switch script[i] {
	case 1:
		i++
		if n != int(script[i]) {
			return false
		}
		i++
	case 2:
		if len(script) < i + 3 {
			return false
		} else {
			i++
			if n != int(script[i]) {
				return false
			}
		}
	  i += 2
	default:
		if n != int(script[i]) - 80 {
			i++
			return false
		}
	  	i++
	}

	if script[i] != vm.CHECKMULTISIG {
		i++
		return false
	}
	i++
	if len(script) != i {
		return false
	}
	
	return true
}

func IsSignatureCotract(script []byte) bool {
	if len(script) != 35 {
		return false
	}
	if script[0] != 33 || script[34] != vm.CHECKSIG {
		return false
	}
	return true
}