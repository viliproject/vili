package fireauth

import (
	"testing"
	"time"
)

func BenchmarkGenerateCreateToken(b *testing.B) {
	gen := New("some-secret")
	data := Data{"uid": "1"}
	opts := &Option{NotBefore: time.Now().Unix(), Expiration: time.Now().Add(time.Hour).Unix()}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.CreateToken(data, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSign(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = sign("message", "secret")
	}
}

func BenchmarkGenerateClaim(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := generateClaim(Data{"uid": "42"}, &Option{Admin: true}, 0)
		if err != nil {
			b.Fatal(err)
		}
	}
}
