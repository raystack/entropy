package helm

type Status string

const (
	StatusUnknown Status = "unknown"
	StatusSuccess Status = "success"
	StatusFailed  Status = "failed"
)

func (x Status) String() string { return string(x) }
