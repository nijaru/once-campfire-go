package compressionbench

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"testing"

	fastgzip "github.com/klauspost/compress/gzip"
)

type compressor interface {
	io.WriteCloser
	Reset(io.Writer)
}

// Use a full, validated room response captured by bench/application, not a
// synthetic repeated string. This module keeps the trial dependency out of the app.
func BenchmarkGzipMiss(b *testing.B) {
	body, err := os.ReadFile(os.Getenv("CAMPFIRE_BENCH_BODY"))
	if err != nil {
		b.Fatal("set CAMPFIRE_BENCH_BODY to a captured full identity response:", err)
	}
	implementations := []struct {
		name string
		new  func(io.Writer) compressor
	}{
		{"stdlib", func(out io.Writer) compressor { writer, _ := gzip.NewWriterLevel(out, 6); return writer }},
		{"klauspost", func(out io.Writer) compressor { writer, _ := fastgzip.NewWriterLevel(out, 6); return writer }},
	}
	if os.Getenv("BENCH_REVERSE") == "1" {
		implementations[0], implementations[1] = implementations[1], implementations[0]
	}
	for _, implementation := range implementations {
		b.Run(implementation.name, func(b *testing.B) {
			for _, cold := range []bool{false, true} {
				name := "reset"
				if cold {
					name = "new"
				}
				b.Run(name, func(b *testing.B) {
					var output bytes.Buffer
					writer := implementation.new(&output)
					writer.Write(body)
					writer.Close()
					reader, err := gzip.NewReader(bytes.NewReader(output.Bytes()))
					if err != nil {
						b.Fatal(err)
					}
					plain, err := io.ReadAll(reader)
					reader.Close()
					if err != nil || !bytes.Equal(plain, body) {
						b.Fatal("gzip roundtrip failed", err)
					}
					size := output.Len()
					b.SetBytes(int64(len(body)))
					b.ReportAllocs()
					b.ResetTimer()
					for b.Loop() {
						output.Reset()
						if cold {
							writer = implementation.new(&output)
						} else {
							writer.Reset(&output)
						}
						writer.Write(body)
						writer.Close()
					}
					b.ReportMetric(float64(size), "gzip-bytes")
				})
			}
		})
	}
}
