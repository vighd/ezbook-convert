package categorizer

import (
	"strings"

	"ezbook-convert/internal/config"
)

// Categorizer handles transaction categorization
type Categorizer struct {
	config *config.Config
}

// New creates a new Categorizer
func New(cfg *config.Config) *Categorizer {
	return &Categorizer{config: cfg}
}

// Categorize determines the category for a transaction
// Returns main category, subcategory, or ("Uncategorized", "") if no match found
func (c *Categorizer) Categorize(partnerName, transactionType string) (string, string) {
	partnerLower := strings.ToLower(partnerName)
	typeLower := strings.ToLower(transactionType)

	// Priority 1: Exact match
	for categoryName, category := range c.config.Categories {
		for _, exactMatch := range category.ExactMatches {
			if partnerName == exactMatch {
				return categoryName, category.SubCategory
			}
		}
	}

	// Priority 2: Keyword match in partner name
	for categoryName, category := range c.config.Categories {
		for _, keyword := range category.Keywords {
			keywordLower := strings.ToLower(keyword)
			if strings.Contains(partnerLower, keywordLower) {
				return categoryName, category.SubCategory
			}
		}
	}

	// Priority 3: Transaction type fallback
	if strings.Contains(typeLower, "jóváírás") || strings.Contains(typeLower, "fizetés") {
		return "Miscellaneous", "Other Income"
	}

	if strings.Contains(typeLower, "hitel törlesztés") {
		return "Finance & Insurance", "Interest Expense"
	}

	if strings.Contains(typeLower, "készpénz") {
		return "General Transfer", "Deposits & Withdrawals"
	}

	if strings.Contains(typeLower, "díj") || strings.Contains(typeLower, "költség") {
		return "Finance & Insurance", "Service Charge"
	}

	// Default: Miscellaneous with type-specific subcategory
	// This will be determined based on transaction amount sign in converter
	return "Miscellaneous", ""
}

// GetUncategorizedPartners finds partners not in known_partners list
func (c *Categorizer) GetUncategorizedPartners(partners []string) []string {
	var uncategorized []string
	seen := make(map[string]bool)

	for _, partner := range partners {
		partner = strings.TrimSpace(partner)
		if partner == "" {
			continue
		}

		// Skip if already seen
		if seen[partner] {
			continue
		}
		seen[partner] = true

		// Skip if in known partners
		if c.config.IsKnownPartner(partner) {
			continue
		}

		uncategorized = append(uncategorized, partner)
	}

	return uncategorized
}
