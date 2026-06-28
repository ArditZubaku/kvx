package server

var EngineStatus = struct {
	Waiting      int32 // TODO: Swap with atomic.Int32
	Busy         int32
	ShuttingDown int32
}{
	Waiting:      1 << 1,
	Busy:         1 << 2,
	ShuttingDown: 1 << 3,
}
