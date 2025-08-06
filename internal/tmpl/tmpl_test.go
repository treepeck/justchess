package tmpl

import (
	"testing"
)

type DummyWriter struct {
}

func (dw DummyWriter) Write(p []byte) (n int, err error) {
	return 0, nil
}

func BenchmarkExec(b *testing.B) {
	for b.Loop() {
		Exec(DummyWriter{}, "signup.html")
	}
}
