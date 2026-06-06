package core

// array of 4 maps (in Redis you can have up to 16 databases)
var KeySpaceStat [4]map[string]int

func UpdateDBStat(num int, metric string, value int) {
	KeySpaceStat[num][metric] = value
}
