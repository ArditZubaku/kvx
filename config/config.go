package config

var (
	Host string
	Port int
)

// On a production implementation I would've constraint the memory
var KeysLimit int

var (
	EvictionRatio    float64
	EvictionStrategy string
)

var AofFile string
