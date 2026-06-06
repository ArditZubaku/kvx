package config

var Host string
var Port int

// On a production implementation I would've constraint the memory
var KeysLimit int

var EvictionRatio float64
var EvictionStrategy string

var AOF_FILE string
