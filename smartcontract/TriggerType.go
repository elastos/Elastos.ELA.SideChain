package smartcontract

type TriggerType byte

const (
	Verification TriggerType = iota
	Application
)