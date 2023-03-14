package helm

const (
	StatusUnknown Status = "unknown"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
)

type Status string

func (x Status) String() string { return string(x) }
