package pattern_test

import (
	"testing"

	"github.com/oarkflow/filters/pattern"
)

func BenchmarkInsert(b *testing.B) {
	b.ResetTimer()
	a := 3
	bc := 15
	matcher := pattern.
		Match[int](a, bc).
		Case(func(args ...any) (int, error) {
			return 5, nil
		}, pattern.NONE, 15).
		Default(func(args ...any) (int, error) {
			return 2, nil
		})
	for i := 0; i < b.N; i++ {
		matcher.Result()
	}
}
