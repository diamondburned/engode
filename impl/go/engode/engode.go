package engode

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	_ "embed"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pedroalbanese/lzma"
)

//go:embed words1024.txt
var WordList1024 string

//go:embed words512.txt
var WordList512 string

//go:embed words256.txt
var WordList256 string

// CompressLZMA compresses the given input using LZMA.
func CompressLZMA(in []byte) ([]byte, error) {
	var buf bytes.Buffer
	wc := lzma.NewWriterSizeLevel(&buf, int64(len(in)), lzma.BestSpeed)
	if _, err := wc.Write(in); err != nil {
		return nil, err
	}
	if err := wc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// CompressZlib compresses the given input using zlib.
func CompressZlib(in []byte) ([]byte, error) {
	var buf bytes.Buffer
	wc, err := zlib.NewWriterLevel(&buf, zlib.BestCompression)
	if err != nil {
		panic(err)
	}
	if _, err := wc.Write(in); err != nil {
		return nil, err
	}
	if err := wc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// CompressGzip compresses the given input using gzip.
func CompressGzip(in []byte) ([]byte, error) {
	var buf bytes.Buffer
	wc, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		panic(err)
	}
	if _, err := wc.Write(in); err != nil {
		return nil, err
	}
	if err := wc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// CompressNone compresses the given input using no compression.
func CompressNone(in []byte) ([]byte, error) {
	return in, nil
}

// Compressor is a function that compresses a byte slice.
type Compressor func([]byte) ([]byte, error)

// Encoder encodes strings into a string of words.
type Encoder struct {
	words       []string
	bitsPerWord uint

	// Compressor is the compression algorithm to use. Defaults to zlib.
	Compressor Compressor
}

// NewDefaultEncoder returns a new Encoder that encodes strings into a string of
// words.
func NewDefaultEncoder() *Encoder {
	e, err := NewCustomEncoder(strings.Fields(WordList1024))
	if err != nil {
		panic(err)
	}
	return e
}

// NewCustomEncoder returns a new Encoder that encodes strings into a string of
// words.
func NewCustomEncoder(words []string) (*Encoder, error) {
	// Calculate the number of bits per dictionary word given the number of
	// words in the dictionary.
	bitsPerWord := int(math.Floor(math.Log2(float64(len(words)))))
	if bitsPerWord < 1 {
		return nil, fmt.Errorf("invalid number of words: %d (minimum 2)", len(words))
	}
	if bitsPerWord > 64 {
		return nil, fmt.Errorf("too many words: %d (maximum 64)", len(words))
	}

	return &Encoder{
		words:       words,
		bitsPerWord: uint(bitsPerWord),

		Compressor: CompressZlib,
	}, nil
}

// Efficiency returns the percentage of dictionary words actually used.
func (e *Encoder) Efficiency() float64 {
	return math.Pow(2, float64(e.bitsPerWord)) / float64(len(e.words))
}

// Encode encodes the given string into a string of words.
func (e *Encoder) Encode(input []byte) ([]string, error) {
	iz, err := e.Compressor(input)
	if err != nil {
		return nil, err
	}

	// spew.Dump(iz)

	// log.Printf("input: %d bytes", len(iz))
	// log.Println("bits per word:", e.bitsPerWord)

	return e.encode(iz)
}

func (e *Encoder) encode(input []byte) ([]string, error) {
	words := make([]string, 0, len(input)*8/int(e.bitsPerWord))

	var lastWord string
	var lastCount int

	flushLast := func() {
		if lastCount >= 4 {
			words = append(words, strconv.Itoa(lastCount), lastWord)
		} else {
			for i := 0; i < lastCount; i++ {
				words = append(words, lastWord)
			}
		}
	}

	r := bitReader(input, e.bitsPerWord)
	for {
		bits, ok := r()
		if !ok {
			break
		}

		word := e.words[bits]
		if lastWord == word {
			lastCount++
			continue
		}

		flushLast()

		lastWord = word
		lastCount = 1
	}

	flushLast()

	return words, nil
}

// func (e *Encoder) Decode(input []string) ([]byte, error) {
// 	var buf bytes.Buffer
// 	w := newBitWriter(&buf)

// 	for _, word := range input {
// 		i := indexOf(e.words, word)
// 		if i < 0 {
// 			return nil, fmt.Errorf("unknown word: %q", word)
// 		}
// 		w.WriteBits64(uint64(i), e.bitsPerWord)
// 	}

// 	if err := w.Err(); err != nil {
// 		return nil, err
// 	}

// 	return gunzip(buf.Bytes())
// }

func compress(in []byte) ([]byte, error) {
	var buf bytes.Buffer
	wc := lzma.NewWriterSizeLevel(&buf, int64(len(in)), lzma.BestSpeed)
	if _, err := wc.Write(in); err != nil {
		return nil, err
	}
	if err := wc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decompress(in []byte) ([]byte, error) {
	cmd := exec.Command("gunzip", "-c")
	cmd.Stdin = bytes.NewReader(in)

	o, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gunzip: %w", err)
	}

	return o, nil
}

// bitReader is the minimal version of bitReader.
// It is taken from https://go.dev/src/compress/bzip2/bit_reader.go.
func bitReader(in []byte, want uint) func() (uint64, bool) {
	var n uint64
	var bits uint
	return func() (uint64, bool) {
		for want > bits {
			if len(in) == 0 {
				return 0, false
			}

			b := in[0]
			in = in[1:]

			n <<= 8
			n |= uint64(b)
			bits += 8
		}
		n = (n >> (bits - want)) & ((1 << want) - 1)
		bits -= want
		return n, true
	}
}

/*
const (
	minASCII  = ' '
	maxASCII  = '~'
	numASCIIs = maxASCII - minASCII + 1
)

// glomp is a combination of printable ASCII characters. If we have 2 gumps per
// word, we can represent "hello" as 3 pairs of gumps: [he], [ll], [o'\0'].
//
// This simplifies our work of encoding a list of gumps into a list of words.
type glomp string

// Encoder is a type that can encode a string into a string of words.
type Encoder struct {
	glompSize int
	words     []string
}

// NewDefaultEncoder returns a new Encoder that encodes strings into a string of
// words.
func NewDefaultEncoder() *Encoder {
	e, err := NewCustomEncoder(strings.Fields(wordListFile))
	if err != nil {
		panic(err)
	}
	return e
}

// NewCustomEncoder returns a new Encoder that encodes strings into a string of
// words.
func NewCustomEncoder(words []string) (*Encoder, error) {
	// We have 95 possible combinations of printable ASCII characters. We can
	// calculate the number of glomps we need to represent each word. Say for
	// example our dictionary is 200 words long. 95*2 is 190, so we need 2
	// glomps per word.
	glompSize := len(words) / numASCIIs
	if glompSize == 0 {
		return nil, fmt.Errorf("not enough words to use (min 95, got %d)", len(words))
	}

	return &Encoder{
		glompSize: glompSize,
		words:     words,
	}, nil
}

// Encode encodes the given string into a string of words.
func (e *Encoder) Encode(input string) (string, error) {
	// Given the glomp size and a string:
	//
	//    - 1 glomp per word, index is just the ASCII value
	//    - 2 glomps per word, index is the ASCII value * 95 + the next ASCII value
	//    - 3 glomps per word, index is the ASCII value * 95^2 + the next ASCII value * 95 + the next ASCII value
	//    - ...
	//
	// So we can calculate the index by multiplying the ASCII value by 95^i for
	// each glomp, where i is the glomp's index.

	var buf strings.Builder
	for input != "" {
		// Get the next glomp.
		var g glomp
		if len(input) >= e.glompSize {
			g = glomp(input[:e.glompSize])
			input = input[e.glompSize:]
		} else {
			g = glomp(input)
			input = ""
		}

		// Calculate the index.
		var index int
		for i, c := range g {
			n := c - minASCII
			if n > maxASCII {
				return "", fmt.Errorf("invalid character %q", c)
			}

			log.Println(int(c-minASCII), "*", pow(numASCIIs, i))
			index += int(c-minASCII) * pow(numASCIIs, i)
		}

		// Write the word.
		buf.WriteString(e.words[index])
		buf.WriteByte(' ')
	}

	return strings.TrimSuffix(buf.String(), " "), nil
}

// conv8bits_7bits converts a byte slice to a slice of concatenated 7-bit
// values. The last byte of the output slice will be padded with 0s on the
// right.
func conv8bitsTo7bits(in []byte) []byte {
	out := make([]byte, (len(in)*8+6)/7)
	for i, b := range in {
		out[i*7/8] |= b << uint(i*7%8)
		out[i*7/8+1] |= b >> uint(8-i*7%8)
	}
	return out
}
*/
