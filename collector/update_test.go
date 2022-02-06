package collector

import (
	"testing"
)

func BenchmarkUpdateCollector(b *testing.B) {
	benchmarkCollector(b, "update", NewUpdateCollector)
}
