package formatter

import "testing"

func FuzzSourceIdempotent(f *testing.F) {
	f.Add(formatterBenchmarkSource)
	f.Add([]byte("domain fuzz\next service StorageService { method ping {} }\n"))
	f.Add([]byte("domain fuzz\r\ndata User{id:string}\r\n"))
	f.Add([]byte("// comment\n@desc(\"value\")\ndata User { value: list<string?> }"))
	f.Add([]byte("0/*\n  */"))
	f.Add([]byte("{\n0/*\n  */"))
	f.Add([]byte("0\"\"\"\n  \"\"\""))
	f.Add([]byte("/ (\n0"))
	f.Fuzz(func(t *testing.T, source []byte) {
		first, err := Source(source)
		if err != nil {
			return
		}
		second, err := Source(first)
		if err != nil {
			t.Fatal(err)
		}
		if string(first) != string(second) {
			t.Fatalf("formatter is not idempotent: first=%q second=%q", first, second)
		}
	})
}
