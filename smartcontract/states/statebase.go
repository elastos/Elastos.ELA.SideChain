package states

import (
	"io"
	"github.com/elastos/Elastos.ELA.Utility/common"
	"errors"
)

type StateBase struct {
	StateVersion byte
}

func (stateBase *StateBase) Serialize(w io.Writer) error {
	common.WriteUint8(w, stateBase.StateVersion)
	return nil;
}

func (stateBase *StateBase) Deserialize(r io.Reader) error {
	stateVersion, err := common.ReadUint8(r)
	if err != nil {
		return errors.New("StateBase StateVersion Deserialize fail.")
	}
	stateBase.StateVersion = stateVersion
	return nil
}