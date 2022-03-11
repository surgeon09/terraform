package addrs

// FIXME: comment
type Checkable interface {
	checkableSigil()

	String() string
}

// The following are all of the possible Checkable address types:
var (
	_ Checkable = AbsResourceInstance{}
	_ Checkable = AbsOutputValue{}
)

type checkable struct {
}

func (c checkable) checkableSigil() {
}
