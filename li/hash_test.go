package li

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"testing"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
)

func benchmarkHashBlake2b(n int, b *testing.B) {
	input := bytes.Repeat([]byte("foo"), n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h, _ := blake2b.New256(nil)
		_, err := h.Write(input)
		ce(err)
	}
}

func BenchmarkHashBlake2b1(b *testing.B)    { benchmarkHashBlake2b(1, b) }
func BenchmarkHashBlake2b4(b *testing.B)    { benchmarkHashBlake2b(4, b) }
func BenchmarkHashBlake2b16(b *testing.B)   { benchmarkHashBlake2b(16, b) }
func BenchmarkHashBlake2b64(b *testing.B)   { benchmarkHashBlake2b(64, b) }
func BenchmarkHashBlake2b256(b *testing.B)  { benchmarkHashBlake2b(256, b) }
func BenchmarkHashBlake2b1024(b *testing.B) { benchmarkHashBlake2b(1024, b) }
func BenchmarkHashBlake2b4096(b *testing.B) { benchmarkHashBlake2b(4096, b) }

func benchmarkHashBlake2s(n int, b *testing.B) {
	input := bytes.Repeat([]byte("foo"), n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h, _ := blake2s.New256(nil)
		_, err := h.Write(input)
		ce(err)
	}
}

func BenchmarkHashBlake2s1(b *testing.B)    { benchmarkHashBlake2s(1, b) }
func BenchmarkHashBlake2s4(b *testing.B)    { benchmarkHashBlake2s(4, b) }
func BenchmarkHashBlake2s16(b *testing.B)   { benchmarkHashBlake2s(16, b) }
func BenchmarkHashBlake2s64(b *testing.B)   { benchmarkHashBlake2s(64, b) }
func BenchmarkHashBlake2s256(b *testing.B)  { benchmarkHashBlake2s(256, b) }
func BenchmarkHashBlake2s1024(b *testing.B) { benchmarkHashBlake2s(1024, b) }
func BenchmarkHashBlake2s4096(b *testing.B) { benchmarkHashBlake2s(4096, b) }

func benchmarkHashSha256(n int, b *testing.B) {
	input := bytes.Repeat([]byte("foo"), n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := sha256.New()
		_, err := h.Write(input)
		ce(err)
	}
}

func BenchmarkHashSha2561(b *testing.B)    { benchmarkHashSha256(1, b) }
func BenchmarkHashSha2564(b *testing.B)    { benchmarkHashSha256(4, b) }
func BenchmarkHashSha25616(b *testing.B)   { benchmarkHashSha256(16, b) }
func BenchmarkHashSha25664(b *testing.B)   { benchmarkHashSha256(64, b) }
func BenchmarkHashSha256256(b *testing.B)  { benchmarkHashSha256(256, b) }
func BenchmarkHashSha2561024(b *testing.B) { benchmarkHashSha256(1024, b) }
func BenchmarkHashSha2564096(b *testing.B) { benchmarkHashSha256(4096, b) }

func benchmarkHashSha512(n int, b *testing.B) {
	input := bytes.Repeat([]byte("foo"), n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := sha512.New()
		_, err := h.Write(input)
		ce(err)
	}
}

func BenchmarkHashSha5121(b *testing.B)    { benchmarkHashSha512(1, b) }
func BenchmarkHashSha5124(b *testing.B)    { benchmarkHashSha512(4, b) }
func BenchmarkHashSha51216(b *testing.B)   { benchmarkHashSha512(16, b) }
func BenchmarkHashSha51264(b *testing.B)   { benchmarkHashSha512(64, b) }
func BenchmarkHashSha512512(b *testing.B)  { benchmarkHashSha512(512, b) }
func BenchmarkHashSha5121024(b *testing.B) { benchmarkHashSha512(1024, b) }
func BenchmarkHashSha5124096(b *testing.B) { benchmarkHashSha512(4096, b) }
