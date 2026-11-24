package main

import (
	"flag"
	"io"
	"log"
	"os"

	"git.sr.ht/~ser/godjot/v2/djot_html"
	"git.sr.ht/~ser/godjot/v2/djot_parser"
)

func main() {
	from := flag.String("from", "", "path to the input djot file (empty or '-' for stdin)")
	to := flag.String("to", "", "path to the output html file (empty or '-' for stdout)")
	overwrite := flag.Bool("overwrite", false, "overwrite output html file")
	flag.Parse()

	var inReader io.Reader
	var outWriter io.Writer
	if *from == "" || *from == "-" {
		inReader = os.Stdin
	} else {
		f, err := os.Open(*from)
		if err != nil {
			log.Fatalf("failed to open input file %v: %v", *from, err)
		}
		inReader = f
	}
	if *to == "" || *to == "-" {
		outWriter = os.Stdout
	} else {
		flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
		if !*overwrite {
			flags |= os.O_EXCL
		}
		f, err := os.OpenFile(*to, flags, 0660)
		if err != nil {
			log.Fatalf("failed to open output file %v: %v", *to, err)
		}
		outWriter = f
	}
	input, err := io.ReadAll(inReader)
	if err != nil {
		log.Fatalf("failed to read input file %v: %v", *from, err)
	}
	ast := djot_parser.BuildDjotAst(input)
	html := []byte(djot_html.New().ConvertDjot(&djot_html.HtmlWriter{}, ast...).String())
	for len(html) > 0 {
		n, err := outWriter.Write(html)
		if err != nil {
			log.Fatalf("failed to write output file %v: %v", *to, err)
		}
		html = html[n:]
	}
}
