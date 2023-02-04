package engode

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	_ "embed"
)

//go:embed test.dat
var testCpp []byte

//go:embed test_minified.dat
var testMinifiedCpp []byte

func BenchmarkEncode(b *testing.B) {
	dicts := []struct {
		name string
		data string
	}{
		{"words1024", WordList1024},
		{"words512", WordList512},
		{"words256", WordList256},
	}

	inputs := []struct {
		name string
		data []byte
	}{
		{"normal", testCpp},
		{"minified", testMinifiedCpp},
	}

	compressors := []struct {
		name     string
		compress func([]byte) ([]byte, error)
	}{
		{"none", func(in []byte) ([]byte, error) { return in, nil }},
		{"lzma", CompressLZMA},
		{"zlib", CompressZlib},
		{"gzip", CompressGzip},
	}

	for _, in := range inputs {
		for _, c := range compressors {
			for _, d := range dicts {
				enc, err := NewCustomEncoder(strings.Fields(d.data))
				if err != nil {
					b.Fatal(err)
				}
				enc.Compressor = c.compress

				b.Run("encode-"+in.name+"-"+c.name+"-"+d.name, func(b *testing.B) {
					var out []string
					for i := 0; i < b.N; i++ {
						o, err := enc.Encode(in.data)
						if err != nil {
							b.Fatal(err)
						}
						out = o
					}
					b.ReportMetric(float64(len(strings.Join(out, " "))), "chars")
					b.ReportMetric(float64(len(out)), "words")
				})
			}
		}
	}
}

func TestEncode(t *testing.T) {
	type test struct {
		in  string
		out []string
	}

	tests := []test{
		{"hello world", []string{
			"lot", "now", "sky", "said", "south", "plain", "full", "salt",
			"danger", "written", "enter", "act", "a", "all", "your", "so",
			"bad", "rope",
		}},
	}

	enc := NewDefaultEncoder()

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			out, err := enc.Encode([]byte(tt.in))
			if err != nil {
				t.Errorf("%q: %v", tt.in, err)
			}

			if !reflect.DeepEqual(out, tt.out) {
				t.Fatalf("unexpected Encode(%q):\n"+
					"got      %#v\n"+
					"expected %#v", tt.in, out, tt.out)
			}
		})
	}
}
