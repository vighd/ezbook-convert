package main

import (
	"flag"
	"fmt"
	"os"
	"text/template"

	"ezbook-convert/cmd"
)

const version = "1.0.0"

const helpTemplate = `ezbook-convert - Convert K&H Bank exports to ezBookkeeping format

Usage:
  ezbook-convert <command> [flags]

Commands:
  convert        Convert K&H TSV to ezBookkeeping CSV
  update-config  Generate LLM prompt for updating categorization config
  version        Show version information
  help           Show this help message

Convert flags:
  --input        Input K&H TSV file path (required)
  --output       Output ezBookkeeping CSV file path (required)
  --account-name Account name for transactions (required)
  --config       YAML config file path (optional)

Update-config flags:
  --input        Input K&H TSV file path (required)
  --config       YAML config file path (default: categories.yaml)

Examples:
  ezbook-convert convert --input kh.csv --output ezbook.csv --account-name "K&H" --config categories.yaml
  ezbook-convert update-config --input kh.csv --config categories.yaml
`

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "convert":
		runConvert()
	case "update-config":
		runUpdateConfig()
	case "version":
		fmt.Printf("ezbook-convert version %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runConvert() {
	fs := flag.NewFlagSet("convert", flag.ExitOnError)
	inputPath := fs.String("input", "", "Input K&H TSV file path (required)")
	outputPath := fs.String("output", "", "Output ezBookkeeping CSV file path (required)")
	accountName := fs.String("account-name", "", "Account name for transactions (required)")
	configPath := fs.String("config", "", "YAML config file path (optional)")

	fs.Parse(os.Args[2:])

	if *inputPath == "" || *outputPath == "" || *accountName == "" {
		fmt.Fprintf(os.Stderr, "Error: --input, --output, and --account-name are required\n\n")
		fs.PrintDefaults()
		os.Exit(1)
	}

	if err := cmd.ConvertCmd(*inputPath, *outputPath, *accountName, *configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runUpdateConfig() {
	fs := flag.NewFlagSet("update-config", flag.ExitOnError)
	inputPath := fs.String("input", "", "Input K&H TSV file path (required)")
	configPath := fs.String("config", "categories.yaml", "YAML config file path")

	fs.Parse(os.Args[2:])

	if *inputPath == "" {
		fmt.Fprintf(os.Stderr, "Error: --input is required\n\n")
		fs.PrintDefaults()
		os.Exit(1)
	}

	if err := cmd.UpdateConfigCmd(*inputPath, *configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	tmpl := template.Must(template.New("help").Parse(helpTemplate))
	tmpl.Execute(os.Stdout, nil)
}
