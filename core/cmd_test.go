package core

import "testing"

// Helper to create a dummy slice of values
func makeValueSlice(size int) []RedisCmd {
	slice := make([]RedisCmd, size)
	for i := range size {
		slice[i] = RedisCmd{Cmd: "SET", Args: []string{"key", "value"}}
	}
	return slice
}

// Helper to create a dummy slice of pointers
func makePointerSlice(size int) []RedisCmd {
	slice := make([]RedisCmd, size)
	for i := range size {
		slice[i] = RedisCmd{Cmd: "SET", Args: []string{"key", "value"}}
	}
	return slice
}

const sliceSize = 10000

// Benchmark for Slice of Values
func BenchmarkValueSlice(b *testing.B) {
	cmds := makeValueSlice(sliceSize)
	b.ResetTimer() // Don't count the setup time

	for b.Loop() {
		var totalLen int
		for j := range cmds {
			// Simulating reading data
			totalLen += len(cmds[j].Cmd)
		}
	}
}

// Benchmark for Slice of Pointers
func BenchmarkPointerSlice(b *testing.B) {
	cmds := makePointerSlice(sliceSize)
	b.ResetTimer()

	for b.Loop() {
		var totalLen int
		for j := range cmds {
			// Simulating reading data via pointer indirection
			totalLen += len(cmds[j].Cmd)
		}
	}
}
