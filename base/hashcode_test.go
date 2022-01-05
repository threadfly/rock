package base

import (
	"testing"
)

var (
	Str  = "I am working on a Java"
	Byts = []byte(Str)

	BigStr  = "When copying from multiple sources, DistCp will abort the copy with an error message if tw"
	BigByts = []byte(BigStr)
)

func BenchStringHashCode(s string, b *testing.B) {
	for i := 0; i < b.N; i++ {
		StringHashCode(s)
	}
}

func BenchStringHashCodeV2(s string, b *testing.B) {
	for i := 0; i < b.N; i++ {
		StringHashCodeV2(s)
	}
}

func Benchmark_StringHashCode_Small(b *testing.B) {
	BenchStringHashCode(Str, b)
}

func Benchmark_StringHashCode_Big(b *testing.B) {
	BenchStringHashCode(BigStr, b)
}

func Benchmark_StringHashCodeV2_Small(b *testing.B) {
	BenchStringHashCodeV2(Str, b)
}

func Benchmark_StringHashCodeV2_Big(b *testing.B) {
	BenchStringHashCodeV2(BigStr, b)
}
