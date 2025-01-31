package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"html/template"
	"log"
	"os"
)

type Doc struct {
	Title      string
	Stylesheet string
	Body       template.HTML
}

var (
	stylesheetFilename string
	outputFilename     string
	documentTitle      string
	//go:embed template.gohtml
	outputTemplateSource string
	outputTemplate       *template.Template
)

func main() {
	flag.StringVar(&outputFilename, "o", "", "output filename")
	flag.StringVar(&stylesheetFilename, "s", "", "include link tag to stylesheet")
	flag.StringVar(&documentTitle, "t", "", "document title")

	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("missing markdown file")
	}

	outputTemplate = template.Must(template.New("template.gohtml").Parse(outputTemplateSource))

	var md = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(true),
				),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	var convertedHtml bytes.Buffer
	for _, inputFilename := range flag.Args() {
		if inputBytes, err := os.ReadFile(inputFilename); err == nil {
			if err := md.Convert(inputBytes, &convertedHtml); err != nil {
				log.Fatal(err)
			}
		}
	}

	var finalHtml bytes.Buffer
	if err := outputTemplate.Execute(&finalHtml, Doc{documentTitle, stylesheetFilename, template.HTML(convertedHtml.String())}); err != nil {
		log.Fatal(err)
	}
	if outputFilename != "" {
		if err := os.WriteFile(outputFilename, finalHtml.Bytes(), 0644); err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println(finalHtml.String())
	}
}
