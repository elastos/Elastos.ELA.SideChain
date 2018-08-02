package contract

func ContractParameterTypeToByte(c []ContractParameterType) []byte {
	size := len(c)
	b := make([]byte, size)
	for i := 0; i < size; i++ {
		b[i] = byte(c[i])
	}
	return b
}

func ByteToContractParameterType(b []byte) []ContractParameterType {
	size := len(b)
	c := make([]ContractParameterType, size)
	for i := 0; i < size; i++ {
		c[i] = ContractParameterType(b[i])
	}

	return c
}