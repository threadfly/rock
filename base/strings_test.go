package base

import (
	"testing"
)

var (
	str  = "I am working on a Java application which has a built in HTTP server, at the moment the server is implemented using ServerSocketChannel, it listens on port 1694 for requests"
	byts = []byte(str)

	bigStr  = "When copying from multiple sources, DistCp will abort the copy with an error message if two sources collide, but collisions at the destination are resolved per the options specified. By default, files already existing at the destination are skipped (i.e. not replaced by the source file). A count of skipped files is reported at the end of each job, but it may be inaccurate if a copier failed for some subset of its files, but succeeded on a later attempt.  It is important that each NodeManager can reach and communicate with both the source and destination file systems. For HDFS, both the source and destination must be running the same version of the protocol or use a backwards-compatible protocol; see Copying Between Versions.  After a copy, it is recommended that one generates and cross-checks a listing of the source and destination to verify that the copy was truly successful. Since DistCp employs both Map/Reduce and the FileSystem API, issues in or between any of the three could adversely and silently affect the copy. Some have had success running with -update enabled to perform a second pass, but users should be acquainted with its semantics before attempting this.  It’s also worth noting that if another client is still writing to a source file, the copy will likely fail. Attempting to overwrite a file being written at the destination should also fail on HDFS. If a source file is (re)moved before it is copied, the copy will fail with a FileNotFoundException.  Please refer to the detailed Command Line Reference for information on all the options available in DistCp."
	bigByts = []byte(bigStr)
)

func ForceStringToBytes(str string, b *testing.B) {
	var byts []byte
	for i := 0; i < b.N; i++ {
		byts = []byte(str)
	}
	b.Logf("byts's len:%d", len(byts))
}

func ForceBytesToString(byts []byte, b *testing.B) {
	var str string
	for i := 0; i < b.N; i++ {
		str = string(byts)
	}
	b.Logf("str's len:%d", len(str))
}

func BenchStringToBytes(str string, b *testing.B) {
	var byts []byte
	for i := 0; i < b.N; i++ {
		byts = StringToBytes(str)
	}
	b.Logf("byts's len:%d", len(byts))
}

func BenchBytesToString(byts []byte, b *testing.B) {
	var str string
	for i := 0; i < b.N; i++ {
		str = BytesToString(byts)
	}
	b.Logf("str's len:%d", len(str))
}

func Benchmark_Small_Graceful_StringToBytes(b *testing.B) {
	BenchStringToBytes(str, b)
}

func Benchmark_Small_Graceful_BytesToString(b *testing.B) {
	BenchBytesToString(byts, b)
}

func Benchmark_Small_Force_StringToBytes(b *testing.B) {
	ForceStringToBytes(str, b)
}

func Benchmark_Small_Force_BytesToString(b *testing.B) {
	ForceBytesToString(byts, b)
}

func Benchmark_Big_Graceful_StringToBytes(b *testing.B) {
	BenchStringToBytes(bigStr, b)
}

func Benchmark_Big_Graceful_BytesToString(b *testing.B) {
	BenchBytesToString(bigByts, b)
}

func Benchmark_Big_Force_StringToBytes(b *testing.B) {
	ForceStringToBytes(bigStr, b)
}

func Benchmark_Big_Force_BytesToString(b *testing.B) {
	ForceBytesToString(bigByts, b)
}
