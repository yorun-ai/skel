package formatter

import "testing"

var formatterBenchmarkSource = []byte(`domain benchmark.demo
@desc("User payload")
pub data User<T> {
    id: uuid
    values: list<map<string, T?>>
}
`)

func BenchmarkSource(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		if _, err := Source(formatterBenchmarkSource); err != nil {
			b.Fatal(err)
		}
	}
}
