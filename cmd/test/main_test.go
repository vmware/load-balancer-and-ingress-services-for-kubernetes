package main

import "testing"

// func BenchmarkLogger(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		Logger()
// 	}
// }

func BenchmarkLoggerWithContext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LoggerWithContext()
	}
}
