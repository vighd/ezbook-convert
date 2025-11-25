package cmd

import (
	"fmt"
	"os"

	"ezbook-convert/internal/categorizer"
	"ezbook-convert/internal/config"
	"ezbook-convert/internal/converter"
	"ezbook-convert/internal/parser"
)

// ConvertCmd executes the convert command
func ConvertCmd(inputPath, outputPath, accountName, configPath string) error {
	// Load config
	cfg, err := loadConfigOrDefault(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Parse K&H export
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	khTransactions, err := parser.ParseKHExport(inputFile)
	if err != nil {
		return fmt.Errorf("failed to parse K&H export: %w", err)
	}

	fmt.Printf("Parsed %d transactions from K&H export\n", len(khTransactions))

	// Convert to ezBookkeeping format
	cat := categorizer.New(cfg)
	conv := converter.New(cat, accountName)

	ezTransactions, convErrors := conv.Convert(khTransactions)

	// Report conversion errors
	if len(convErrors) > 0 {
		fmt.Fprintf(os.Stderr, "\nWarning: %d transactions failed to convert:\n", len(convErrors))
		for _, err := range convErrors {
			fmt.Fprintf(os.Stderr, "  - %v\n", err)
		}
	}

	fmt.Printf("Successfully converted %d transactions\n", len(ezTransactions))

	// Write output
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	if err := converter.WriteCSV(outputFile, ezTransactions); err != nil {
		return fmt.Errorf("failed to write CSV: %w", err)
	}

	fmt.Printf("\n✓ Conversion complete! Output written to: %s\n", outputPath)

	return nil
}

func loadConfigOrDefault(configPath string) (*config.Config, error) {
	if configPath == "" {
		// No config provided, use empty config
		return &config.Config{
			KnownPartners: []string{},
			Categories:    make(map[string]*config.Category),
		}, nil
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: Config file not found, using empty categories\n")
			return &config.Config{
				KnownPartners: []string{},
				Categories:    make(map[string]*config.Category),
			}, nil
		}
		return nil, err
	}

	return cfg, nil
}
