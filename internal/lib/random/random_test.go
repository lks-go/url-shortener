package random_test

import (
	"testing"

	"github.com/lks-go/url-shortener/internal/lib/random"
)

func BenchmarkNewString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		random.NewString(6)
	}
}
