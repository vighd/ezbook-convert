package anonymizer

import (
	"regexp"
	"strings"
)

// AnonymizationResult contains the anonymized merchant name and detection info
type AnonymizationResult struct {
	Original      string
	Anonymized    string
	IsPersonal    bool
	DetectionType string // "owner_name", "transfer_partner", "account_number", "none"
}

var (
	// Account number patterns
	accountNumberPattern = regexp.MustCompile(`^[A-Z]{2}\d{10,34}$`) // IBAN
)

// Config for anonymization
type Config struct {
	OwnerName string // Account owner name to anonymize
}

// Anonymize determines if a merchant name should be anonymized
func Anonymize(merchantName, transactionType string, cfg *Config) AnonymizationResult {
	trimmed := strings.TrimSpace(merchantName)
	
	if trimmed == "" {
		return AnonymizationResult{
			Original:      merchantName,
			Anonymized:    merchantName,
			IsPersonal:    false,
			DetectionType: "none",
		}
	}

	// Check for account number (IBAN format)
	if accountNumberPattern.MatchString(trimmed) {
		return AnonymizationResult{
			Original:      merchantName,
			Anonymized:    "[ACCOUNT_NUMBER]",
			IsPersonal:    true,
			DetectionType: "account_number",
		}
	}

	// Check if it's the account owner's name
	if cfg != nil && cfg.OwnerName != "" {
		if strings.EqualFold(trimmed, cfg.OwnerName) {
			return AnonymizationResult{
				Original:      merchantName,
				Anonymized:    "[OWNER_NAME]",
				IsPersonal:    true,
				DetectionType: "owner_name",
			}
		}
	}

	// Check if it's a transfer (átutalás) - partner names in transfers are personal
	typeLower := strings.ToLower(transactionType)
	isTransfer := strings.Contains(typeLower, "átutalás") || 
	              strings.Contains(typeLower, "átvezetés") ||
	              strings.Contains(typeLower, "jóváírás")
	
	if isTransfer {
		// In transfers, the partner name is likely a person
		return AnonymizationResult{
			Original:      merchantName,
			Anonymized:    "[TRANSFER_PARTNER]",
			IsPersonal:    true,
			DetectionType: "transfer_partner",
		}
	}

	// Default: not personal data (business/merchant names)
	return AnonymizationResult{
		Original:      merchantName,
		Anonymized:    merchantName,
		IsPersonal:    false,
		DetectionType: "none",
	}
}

// AnonymizeMerchantList anonymizes a list of merchant names with their transaction types
func AnonymizeMerchantList(merchants []string, transactionTypes []string, cfg *Config) []AnonymizationResult {
	results := make([]AnonymizationResult, len(merchants))
	for i, merchant := range merchants {
		transType := ""
		if i < len(transactionTypes) {
			transType = transactionTypes[i]
		}
		results[i] = Anonymize(merchant, transType, cfg)
	}
	return results
}
