package plans

//go:generate go run golang.org/x/tools/cmd/stringer -type ConditionType
type ConditionType int

const (
	InvalidCondition      ConditionType = 0
	ResourcePrecondition  ConditionType = 1
	ResourcePostcondition ConditionType = 2
	OutputPrecondition    ConditionType = 3
)
