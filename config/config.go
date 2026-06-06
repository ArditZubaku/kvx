package config

var Host string
var Port int

// On a production implementation I would've constraint the memory
var KeysLimit int

var AOF_FILE string = "./kvx-master.aof"
