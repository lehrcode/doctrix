package main

import (
	"bytes"
	_ "embed"
	"flag"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
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
	outputFormat         string
	documentMargin       float64
)

func main() {
	flag.StringVar(&outputFilename, "o", "", "output filename")
	flag.StringVar(&stylesheetFilename, "s", "", "include link tag to stylesheet")
	flag.StringVar(&documentTitle, "t", "", "document title")
	flag.StringVar(&outputFormat, "f", "html", "output format")
	flag.Float64Var(&documentMargin, "m", 1.5, "document margin in cm")

	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("missing markdown file")
	}

	outputTemplate = template.Must(template.New("template.gohtml").Parse(outputTemplateSource))

	if outputFilename == "" {
		outputFilename = strings.TrimSuffix(flag.Arg(0), filepath.Ext(flag.Arg(0))) + "." + outputFormat
	}

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
	if strings.EqualFold(outputFormat, "html") {
		if err := os.WriteFile(outputFilename, finalHtml.Bytes(), 0644); err != nil {
			log.Fatal(err)
		}
	} else if strings.EqualFold(outputFormat, "pdf") {
		var finalPdf bytes.Buffer
		if err := Html2pdf(finalHtml.String(), &finalPdf, documentMargin*0.3937007874); err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(outputFilename, finalPdf.Bytes(), 0644); err != nil {
			log.Fatal(err)
		}
	}
}
