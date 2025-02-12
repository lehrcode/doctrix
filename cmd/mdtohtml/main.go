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
	"github.com/yuin/goldmark/renderer/html"
	"html/template"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Document struct {
	Title  string
	Values url.Values
	Body   template.HTML
}

var (
	outputFilename string
	document       Document

	//go:embed template.gohtml
	outputTemplateSource   string
	outputTemplateFilename string
	outputTemplate         *template.Template
)

func main() {
	var valuesString string
	flag.StringVar(&valuesString, "e", "lang=en", "extra values")
	flag.StringVar(&outputFilename, "o", "", "output filename")
	flag.StringVar(&document.Title, "t", "", "document title")
	flag.StringVar(&outputTemplateFilename, "m", "", "html template")

	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("missing markdown file")
	}

	if v, err := url.ParseQuery(valuesString); err != nil {
		log.Fatal(err)
	} else {
		document.Values = v
		log.Printf("Extra values: %#v\n", document.Values)
	}

	if outputTemplateFilename != "" {
		log.Printf("Output template filename: %s\n", outputTemplateFilename)
		if t, err := template.ParseFiles(outputTemplateFilename); err != nil {
			log.Fatal(err)
		} else {
			outputTemplate = t
		}
	} else {
		outputTemplate = template.Must(template.New("template.gohtml").Parse(outputTemplateSource))
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
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
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
	document.Body = template.HTML(convertedHtml.String())
	if err := outputTemplate.Execute(&finalHtml, document); err != nil {
		log.Fatal(err)
	}
	if outputFilename == "" {
		outputFilename = strings.TrimSuffix(filepath.Base(flag.Arg(0)), filepath.Ext(flag.Arg(0))) + ".html"
	}

	if err := os.WriteFile(outputFilename, finalHtml.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}
