package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/diamondburned/engode/impl/go/engode"
)

var dictFile = "engode/words1024.txt"
var dictEfficiency = false
var compressor = "zlib"

func init() {
	flag.StringVar(&dictFile, "dict", dictFile, "dictionary file")
	flag.BoolVar(&dictEfficiency, "dict-efficiency", dictEfficiency, "show dictionary usage")
	flag.StringVar(&compressor, "compressor", compressor, "compressor")
	flag.Parse()
}

var compressors = map[string]engode.Compressor{
	"none": engode.CompressNone,
	"lzma": engode.CompressLZMA,
	"zlib": engode.CompressZlib,
	"gzip": engode.CompressGzip,
}

func main() {
	comp, ok := compressors[compressor]
	if !ok {
		log.Fatalf("unknown compressor %q", compressor)
	}

	dict, err := os.ReadFile(dictFile)
	if err != nil {
		log.Fatalln("cannot read dict:", err)
	}

	f, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalln("stdin:", err)
	}

	enc, err := engode.NewCustomEncoder(strings.Fields(string(dict)))
	if err != nil {
		log.Fatalln("cannot create encoder:", err)
	}

	enc.Compressor = comp

	if dictEfficiency {
		log.Printf("dictionary efficiency: %.2f%%", enc.Efficiency()*100)
	}

	words, err := enc.Encode(f)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(strings.Join(words, " "))
}
