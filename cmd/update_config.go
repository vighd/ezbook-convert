package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"ezbook-convert/internal/anonymizer"
	"ezbook-convert/internal/categorizer"
	"ezbook-convert/internal/config"
	"ezbook-convert/internal/parser"
	"gopkg.in/yaml.v3"
)

const llmPromptTemplate = `=== PROMPT FOR LLM ===
Copy the text below and paste it into ChatGPT/Gemini:

---
I have a transaction categorization config in YAML format for my personal finance app.
I received new transactions that contain unknown merchants/partners.

CURRENT CONFIG:
---
{{.CurrentConfigYAML}}
---

NEW UNCATEGORIZED MERCHANTS:
{{- if .OwnerNames}}
{{.OwnerNamesIndex}}. [OWNER_NAME] ({{len .OwnerNames}} occurrences) - Account owner's transfers, suggested: Miscellaneous / Other Income or Other Expense
{{- end}}
{{- if .TransferPartners}}
{{.TransferPartnersIndex}}. [TRANSFER_PARTNER] ({{len .TransferPartners}} occurrences) - Personal transfers detected, suggested: Miscellaneous / Other Income or Other Expense
{{- end}}
{{- if .AccountNumbers}}
{{.AccountNumbersIndex}}. [ACCOUNT_NUMBER] ({{len .AccountNumbers}} occurrences) - Bank account transfers, suggested: General Transfer / Bank Transfer
{{- end}}
{{- range $i, $merchant := .Businesses}}
{{add $i $.BusinessStartIndex}}. "{{$merchant}}"
{{- end}}

IMPORTANT INSTRUCTIONS:
1. RESEARCH each merchant carefully:
   - Use internet search to identify the business type
   - Look for keywords in the merchant name (e.g., "Pekseg" = bakery, "Patika" = pharmacy)
   - Many names contain the business category directly (e.g., "ABC" = grocery, "Kft" = company)
   - Hungarian business names often include type hints (pékség, patika, étterem, benzinkút, etc.)

2. CATEGORIZATION RULES:
   - Assign appropriate categories based on merchant business type
   - For [OWNER_NAME], [TRANSFER_PARTNER], and [ACCOUNT_NUMBER], use the suggested categories above
   - Add keywords to existing categories that match the business type
   - Only create new category if absolutely necessary (prefer existing ones)

3. YAML FORMATTING - CRITICAL:
   - Each keyword MUST be on a separate line with a dash (-)
   - DO NOT use nested lists like ["item1", "item2"] 
   - DO NOT use inline arrays
   - Each category MUST have a 'subcategory' field
   
   CORRECT format:
     keywords:
       - keyword1
       - keyword2
       - keyword3
   
   WRONG format (DO NOT DO THIS):
     keywords:
       - ["keyword1", "keyword2"]

4. OUTPUT FORMAT:
   - Your ENTIRE response must be valid YAML (no explanations before/after)
   - Start your response with: known_partners:
   - Put the YAML in a code block using triple backticks:
     ` + "```yaml" + `
     known_partners:
       - .....
     ` + "```" + `
   - This allows easy copy-paste or download

5. COMPLETENESS:
   - Include ALL existing categories from CURRENT CONFIG above
   - Add new merchants to known_partners list (use placeholders like [TRANSFER_PARTNER])
   - Add new keywords to appropriate categories (one per line!)

AVAILABLE CATEGORY NAMES (from ezBookkeeping defaults):
{{.AvailableCategories}}
---

=== END OF PROMPT ===

Found {{.TotalMerchants}} new merchants ({{.AnonymizedCount}} anonymized for privacy).

📋 Next steps:
1. Copy the prompt above (everything between the --- lines)
2. Paste into ChatGPT or Gemini
3. The LLM will return YAML in a code block - click the copy button on the code block
4. VERIFY the YAML format - check that keywords are NOT in nested arrays
5. Save the copied YAML to categories.yaml
6. Run the convert command with the updated config
`

// UpdateConfigCmd executes the update-config command
func UpdateConfigCmd(inputPath, configPath string) error {
	// Load existing config
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

	// Extract all partner names and transaction types
	var partnerNames []string
	var transactionTypes []string
	for _, t := range khTransactions {
		if t.PartnerName != "" {
			partnerNames = append(partnerNames, t.PartnerName)
			transactionTypes = append(transactionTypes, t.Type)
		}
	}

	// Find uncategorized partners
	cat := categorizer.New(cfg)
	uncategorized := cat.GetUncategorizedPartners(partnerNames)

	if len(uncategorized) == 0 {
		fmt.Println("✓ All merchants are already in the known_partners list!")
		fmt.Println("No new categorization needed.")
		return nil
	}

	// Ask for account owner name for privacy protection
	fmt.Println("\n🔒 Privacy Protection Setup")
	fmt.Println("To protect personal data, we need to know the account owner's name.")
	fmt.Print("Enter account owner name (e.g., 'Vigh Dániel') or press Enter to skip: ")
	
	var ownerName string
	fmt.Scanln(&ownerName)
	ownerName = strings.TrimSpace(ownerName)

	// Prepare anonymization config
	anonCfg := &anonymizer.Config{
		OwnerName: ownerName,
	}

	// Build map of partner to transaction type for anonymization
	partnerTypeMap := make(map[string]string)
	for i, partner := range partnerNames {
		if i < len(transactionTypes) {
			partnerTypeMap[partner] = transactionTypes[i]
		}
	}

	// Anonymize merchant names (protect personal data)
	var anonymized []anonymizer.AnonymizationResult
	for _, partner := range uncategorized {
		transType := partnerTypeMap[partner]
		result := anonymizer.Anonymize(partner, transType, anonCfg)
		anonymized = append(anonymized, result)
	}

	// Show user what will be anonymized
	personalDataCount := 0
	for _, result := range anonymized {
		if result.IsPersonal {
			personalDataCount++
		}
	}

	if personalDataCount > 0 {
		fmt.Printf("\n🔒 %d personal data item(s) detected and will be anonymized:\n", personalDataCount)
		for _, result := range anonymized {
			if result.IsPersonal {
				fmt.Printf("  • \"%s\" → %s (%s)\n", result.Original, result.Anonymized, result.DetectionType)
			}
		}
		fmt.Println()
	}

	// Generate LLM prompt
	generateLLMPrompt(cfg, anonymized)

	return nil
}

func generateLLMPrompt(cfg *config.Config, anonymized []anonymizer.AnonymizationResult) {
	// Group by anonymization type
	ownerNames := []string{}
	transferPartners := []string{}
	accountNumbers := []string{}
	businesses := []string{}
	
	for _, result := range anonymized {
		switch result.DetectionType {
		case "owner_name":
			ownerNames = append(ownerNames, result.Anonymized)
		case "transfer_partner":
			transferPartners = append(transferPartners, result.Anonymized)
		case "account_number":
			accountNumbers = append(accountNumbers, result.Anonymized)
		default:
			businesses = append(businesses, result.Anonymized)
		}
	}
	
	// Calculate indices for prompt
	idx := 1
	ownerNamesIndex := 0
	transferPartnersIndex := 0
	accountNumbersIndex := 0
	businessStartIndex := 1
	
	if len(ownerNames) > 0 {
		ownerNamesIndex = idx
		idx++
	}
	if len(transferPartners) > 0 {
		transferPartnersIndex = idx
		idx++
	}
	if len(accountNumbers) > 0 {
		accountNumbersIndex = idx
		idx++
	}
	businessStartIndex = idx

	// Serialize current config to YAML
	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error serializing config: %v\n", err)
		return
	}

	// Prepare template data
	data := struct {
		CurrentConfigYAML     string
		OwnerNames            []string
		TransferPartners      []string
		AccountNumbers        []string
		Businesses            []string
		OwnerNamesIndex       int
		TransferPartnersIndex int
		AccountNumbersIndex   int
		BusinessStartIndex    int
		TotalMerchants        int
		AnonymizedCount       int
		AvailableCategories   string
	}{
		CurrentConfigYAML:     string(yamlData),
		OwnerNames:            ownerNames,
		TransferPartners:      transferPartners,
		AccountNumbers:        accountNumbers,
		Businesses:            businesses,
		OwnerNamesIndex:       ownerNamesIndex,
		TransferPartnersIndex: transferPartnersIndex,
		AccountNumbersIndex:   accountNumbersIndex,
		BusinessStartIndex:    businessStartIndex,
		TotalMerchants:        len(anonymized),
		AnonymizedCount:       len(ownerNames) + len(transferPartners) + len(accountNumbers),
		AvailableCategories:   getAvailableCategories(),
	}

	// Create template with helper functions
	tmpl := template.Must(template.New("prompt").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b + 1 },
	}).Parse(llmPromptTemplate))

	// Execute template
	if err := tmpl.Execute(os.Stdout, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
	}
}

func getAvailableCategories() string {
	categories := []string{
		"Food & Drink",
		"Clothing & Appearance",
		"Housing & Houseware",
		"Transportation",
		"Communication",
		"Entertainment",
		"Education & Studying",
		"Medical & Healthcare",
		"Gift & Social",
		"Finance & Insurance",
		"Miscellaneous",
	}
	return strings.Join(categories, ", ")
}
