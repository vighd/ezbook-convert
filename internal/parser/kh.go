package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"
)

// KHTransaction represents a single transaction from K&H Bank export
type KHTransaction struct {
	Date           string
	TransactionID  string
	Type           string
	AccountNumber  string
	AccountName    string
	PartnerAccount string
	PartnerName    string
	Amount         string
	Currency       string
	Description    string
}

// ParseKHExport reads and parses K&H TSV export file
func ParseKHExport(reader io.Reader) ([]*KHTransaction, error) {
	csvReader := csv.NewReader(reader)
	csvReader.Comma = '\t'
	csvReader.LazyQuotes = true
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading TSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("file must contain at least header and one transaction")
	}

	var transactions []*KHTransaction
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 9 {
			continue // Skip malformed rows
		}

		transactions = append(transactions, &KHTransaction{
			Date:           strings.TrimSpace(record[0]),
			TransactionID:  strings.TrimSpace(record[1]),
			Type:           strings.TrimSpace(record[2]),
			AccountNumber:  strings.TrimSpace(record[3]),
			AccountName:    strings.TrimSpace(record[4]),
			PartnerAccount: strings.TrimSpace(record[5]),
			PartnerName:    strings.TrimSpace(record[6]),
			Amount:         strings.TrimSpace(record[7]),
			Currency:       strings.TrimSpace(record[8]),
			Description:    getField(record, 9),
		})
	}

	return transactions, nil
}

// ParseDate parses K&H date format (YYYY.MM.DD) with optional time
func ParseDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	
	// Try with time first (future-proof)
	if t, err := time.Parse("2006.01.02 15:04:05", dateStr); err == nil {
		return t, nil
	}
	
	// Fall back to date only
	t, err := time.Parse("2006.01.02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
	}
	
	return t, nil
}

func getField(record []string, index int) string {
	if index < len(record) {
		return strings.TrimSpace(record[index])
	}
	return ""
}
