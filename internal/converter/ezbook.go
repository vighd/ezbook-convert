package converter

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"ezbook-convert/internal/categorizer"
	"ezbook-convert/internal/parser"
)

// EzBookTransaction represents a transaction in ezBookkeeping format
type EzBookTransaction struct {
	Type        string
	Category    string
	SubCategory string
	Account     string
	Amount      string
	DateTime    string
	Description string
	Tags        string
}

// Converter handles conversion from K&H to ezBookkeeping format
type Converter struct {
	categorizer *categorizer.Categorizer
	accountName string
}

// New creates a new Converter
func New(cat *categorizer.Categorizer, accountName string) *Converter {
	return &Converter{
		categorizer: cat,
		accountName: accountName,
	}
}

// Convert transforms K&H transactions to ezBookkeeping format
func (c *Converter) Convert(khTransactions []*parser.KHTransaction) ([]*EzBookTransaction, []error) {
	var ezTransactions []*EzBookTransaction
	var errors []error

	for _, kh := range khTransactions {
		ez, err := c.convertSingle(kh)
		if err != nil {
			errors = append(errors, fmt.Errorf("transaction %s: %w", kh.TransactionID, err))
			continue
		}
		ezTransactions = append(ezTransactions, ez)
	}

	return ezTransactions, errors
}

func (c *Converter) convertSingle(kh *parser.KHTransaction) (*EzBookTransaction, error) {
	// Parse date
	date, err := parser.ParseDate(kh.Date)
	if err != nil {
		return nil, err
	}

	// Parse amount
	amount, err := parseAmount(kh.Amount)
	if err != nil {
		return nil, err
	}

	// Determine transaction type
	transactionType := "Expense"
	if amount > 0 {
		transactionType = "Income"
	}
	amount = math.Abs(amount)

	// Categorize
	category, subCategory := c.categorizer.Categorize(kh.PartnerName, kh.Type)

	// If no subcategory was assigned, use default based on transaction type
	if subCategory == "" {
		if transactionType == "Expense" {
			subCategory = "Other Expense"
		} else if transactionType == "Income" {
			subCategory = "Other Income"
		}
	}

	// Build description
	description := buildDescription(kh)

	return &EzBookTransaction{
		Type:        transactionType,
		Category:    category,
		SubCategory: subCategory,
		Account:     c.accountName,
		Amount:      formatAmount(amount),
		DateTime:    formatDateTime(date),
		Description: description,
		Tags:        "",
	}, nil
}

// WriteCSV writes ezBookkeeping transactions to CSV
func WriteCSV(writer io.Writer, transactions []*EzBookTransaction) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header with ezBookkeeping complete export format
	// All 14 columns are required for ezBookkeeping Data Export File format
	header := []string{
		"Time",
		"Timezone",
		"Type",
		"Category",
		"Sub Category",
		"Account",
		"Account Currency",
		"Amount",
		"Account2",
		"Account2 Currency",
		"Account2 Amount",
		"Geographic Location",
		"Tags",
		"Description",
	}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	// Write transactions
	for _, t := range transactions {
		record := []string{
			t.DateTime,
			"+01:00",        // Timezone (Central European Time - Hungary)
			t.Type,
			t.Category,
			t.SubCategory,
			t.Account,
			"HUF",           // Account Currency
			t.Amount,
			"",              // Account2 (for transfers)
			"",              // Account2 Currency
			"",              // Account2 Amount
			"",              // Geographic Location
			t.Tags,
			t.Description,
		}
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func parseAmount(amountStr string) (float64, error) {
	amountStr = strings.ReplaceAll(amountStr, " ", "")
	amountStr = strings.ReplaceAll(amountStr, ",", ".")
	
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount: %s", amountStr)
	}
	
	return amount, nil
}

func formatAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

func formatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func buildDescription(kh *parser.KHTransaction) string {
	var parts []string

	if kh.PartnerName != "" {
		parts = append(parts, kh.PartnerName)
	}

	if kh.Type != "" {
		parts = append(parts, fmt.Sprintf("(%s)", kh.Type))
	}

	// Add description if different from partner name
	if kh.Description != "" && kh.Description != kh.PartnerName {
		parts = append(parts, kh.Description)
	}

	return strings.Join(parts, " - ")
}
