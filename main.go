package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

var templateFile string

func main() {

	var rootCmd = &cobra.Command{
		Use: "tmplgen",
		Run: run,
	}

	rootCmd.Flags().StringVarP(&templateFile, "template-file", "t", "", "template in text/template format")

	rootCmd.Execute()
}

func exitWithError(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%v: %v\n", msg, err)
	os.Exit(1)
}

func run(cmd *cobra.Command, args []string) {

	if templateFile == "" {
		exitWithError("parameter error", fmt.Errorf("template-file not specified"))
	}

	var data csvData
	var funcs = template.FuncMap{
		"field": data.lookupValue,
	}

	t, err := template.New("main").Funcs(funcs).ParseFiles(templateFile)
	if err != nil {
		exitWithError("error parsing template", err)
	}

	rdr := csv.NewReader(os.Stdin)
	rdr.Comment = '#'
	rdr.LazyQuotes = true
	rdr.TrimLeadingSpace = true

	records, err := rdr.ReadAll()
	if err != nil {
		exitWithError("error reading csv input", err)
	}

	err = data.init(records)
	if err != nil {
		exitWithError("error processing input data", err)
	}

	err = t.Templates()[0].Execute(os.Stdout, data)
	if err != nil {
		exitWithError("error executing template", err)
	}
}

type csvData struct {
	Fields  []string
	Records [][]string

	fieldMap map[string]int
}

func (d *csvData) init(records [][]string) error {
	if len(records) == 0 {
		return fmt.Errorf("no records provided")
	}

	d.Fields = records[0]
	d.Records = records[1:]

	d.fieldMap = make(map[string]int)
	for i, f := range d.Fields {
		d.fieldMap[f] = i
	}

	return nil
}

func (d *csvData) lookupValue(record []string, field string) (string, error) {
	if idx, ok := d.fieldMap[field]; ok {
		return record[idx], nil
	} else {
		return "", fmt.Errorf("could not find field %q", field)
	}
}
