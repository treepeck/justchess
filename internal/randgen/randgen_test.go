package randgen

import "testing"

func BenchmarkGenId(b *testing.B) {
	for b.Loop() {
		GenId(SessionIdLen)
	}
}
