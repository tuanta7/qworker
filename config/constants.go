package config

const (
	CriticalQueue = "critical"
	DefaultQueue  = "default"
	LowQueue      = "low"
)

var (
	Queues = map[string]int{
		CriticalQueue: 3,
		DefaultQueue:  2,
		LowQueue:      1,
	}
)
