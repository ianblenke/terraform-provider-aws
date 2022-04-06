//go:build generate
// +build generate

package main

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"
)

const filename = `../../../website/docs/guides/custom-service-endpoints.html.md`

//go:embed custom_endpoints_header.tmpl
var header string

//go:embed custom_endpoints_footer.tmpl
var footer string

type ServiceDatum struct {
	ProviderPackage string
	Aliases         []string
}

type TemplateData struct {
	Services []ServiceDatum
}

const (
	// column indices of CSV
	//awsCLIV2Command         = 0
	//awsCLIV2CommandNoDashes = 1
	//goV1Package             = 2
	//goV2Package             = 3
	//providerPackageActual   = 4
	//providerPackageCorrect  = 5
	//splitPackageRealPackage = 6
	//aliases                 = 7
	//providerNameUpper       = 8
	//goV1ClientName          = 9
	//skipClientGenerate      = 10
	//sdkVersion              = 11
	//resourcePrefixActual    = 12
	//resourcePrefixCorrect   = 13
	//filePrefix              = 14
	//docPrefix               = 15
	//humanFriendly           = 16
	//brand                   = 17
	//exclude                 = 18
	//allowedSubcategory      = 19
	//deprecatedEnvVar        = 20
	//envVar                  = 21
	//note                    = 22
	providerPackageActual  = 4
	providerPackageCorrect = 5
	aliases                = 7
	exclude                = 18
)

func main() {
	f, err := os.Open("../../../names/names_data.csv")
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	csvReader := csv.NewReader(f)

	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	td := TemplateData{}

	for i, l := range data {
		if i < 1 { // no header
			continue
		}

		if l[exclude] != "" {
			continue
		}

		if l[providerPackageActual] == "" && l[providerPackageCorrect] == "" {
			continue
		}

		p := l[providerPackageCorrect]

		if l[providerPackageActual] != "" {
			p = l[providerPackageActual]
		}

		sd := ServiceDatum{
			ProviderPackage: p,
		}

		if l[aliases] != "" {
			sd.Aliases = strings.Split(l[aliases], ";")
		}

		td.Services = append(td.Services, sd)
	}

	sort.SliceStable(td.Services, func(i, j int) bool {
		return td.Services[i].ProviderPackage < td.Services[j].ProviderPackage
	})

	writeTemplate(header+tmpl+footer, "website", td)
}

func writeTemplate(body string, templateName string, td TemplateData) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file (%s): %s", filename, err)
	}

	tplate, err := template.New(templateName).Parse(body)
	if err != nil {
		log.Fatalf("error parsing template: %s", err)
	}

	var buffer bytes.Buffer
	err = tplate.Execute(&buffer, td)
	if err != nil {
		log.Fatalf("error executing template: %s", err)
	}

	if _, err := f.Write(buffer.Bytes()); err != nil {
		f.Close()
		log.Fatalf("error writing to file (%s): %s", filename, err)
	}

	if err := f.Close(); err != nil {
		log.Fatalf("error closing file (%s): %s", filename, err)
	}
}

var tmpl = `
<!-- markdownlint-disable no-inline-html -->
<!--
    The division splits this long list into multiple columns without manually
    maintaining a table. The terraform.io Markdown parser previously allowed
    for Markdown within HTML elements, however the Terraform Registry parser
    is more accurate/strict, so we use raw HTML to maintain this list.
-->
<div style="column-width: 14em;">
<ul>
{{- range .Services }}
  {{- if .Aliases }}
  <li><code>{{ .ProviderPackage }}</code> ({{ range $i, $e := .Aliases }}{{ if gt $i 0 }} {{ end }}or <code>{{ $e }}</code>{{ end }})</li>
  {{- else }}
  <li><code>{{ .ProviderPackage }}</code></li>
  {{- end }}
{{- end }}
</ul>
</div>
<!-- markdownlint-enable no-inline-html -->
`
