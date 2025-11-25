# ezbook-convert

## Project Overview

A CLI tool to convert K&H Bank (Hungary) transaction exports to ezBookkeeping-compatible CSV format.

### What is ezBookkeeping?

ezBookkeeping is an open-source, self-hosted personal finance application that supports importing transactions in CSV format. It uses a hierarchical category system with main categories and subcategories for both Income, Expense, and Transfer transaction types.

### Problem Statement

K&H Bank exports transaction history in TSV (tab-separated) format with Hungarian field names and date formats. This needs to be converted to ezBookkeeping's CSV import format with proper categorization.

## Features

### 1. Convert Command
Converts K&H TSV export to ezBookkeeping CSV format.

**Usage:**
```bash
./ezbook-convert convert \
  --input kh_export.csv \
  --output ezbook.csv \
  --account-name "K&H Account" \
  --config categories.yaml
```

**Functionality:**
- Parses K&H TSV format (21 columns, tab-separated)
- Maps fields: date, amount, partner name, transaction type, description
- Categorizes transactions based on YAML config rules
- Outputs ezBookkeeping-compatible CSV

**Field Mapping:**
- `könyvelés dátuma` (booking date) → `DateTime` (format: 2006-01-02 15:04:05)
- `összeg` (amount) → `Amount` (absolute value)
- Amount sign → `Type` (negative = Expense, positive = Income)
- `partner elnevezése` (partner name) → Used for categorization
- `típus` (type) → Used for categorization
- `közlemény` (description) → `Description`
- Account name from CLI parameter → `Account`

### 2. Update-Config Command
Generates an LLM prompt to help update the category configuration.

**Usage:**
```bash
./ezbook-convert update-config \
  --input kh_export.csv \
  --config categories.yaml
```

**Functionality:**
- Reads existing K&H export
- Identifies new/uncategorized merchant names
- Compares against `known_partners` list in config
- Generates English prompt for ChatGPT/Gemini
- User copies prompt → LLM → saves returned YAML

**Why this approach?**
- Free LLM usage (ChatGPT free tier, Gemini)
- Minimal user interaction
- Reliable categorization by AI
- Config remains version-controlled and editable

## Configuration File Format

**File:** `categories.yaml`

```yaml
# List of known partners (for detecting new merchants)
known_partners:
  - "ALDI 241.SZ."
  - "AUCHAN SZEGED"
  - "foodora"
  - "MOL 63150 sz. toltoall"

# Categorization rules
categories:
  Food & Drink:
    subcategory: "Food"  # Sub-category name (required)
    keywords:
      - aldi
      - auchan
      - lidl
      - cba
      - dm
      - rossmann
      - foodora
      - wolt
      - burger king
    exact_matches:
      - "ALDI 241.SZ."
      
  Transportation:
    subcategory: "Public Transit"
    keywords:
      - easypark
      - szkt
      - benzinkút
      - benzinku
      - mol
      - shell
    exact_matches:
      - "SZKT FEDE LZETI JEGY"
      
  Entertainment:
    subcategory: "Subscriptions"
    keywords:
      - youtube
      - sweet.tv
      - netflix
      - spotify
      - vidampark
      
  Medical & Healthcare:
    subcategory: "Medical Expense"
    keywords:
      - patika
      - gyógyszert
      
  Housing & Houseware:
    subcategory: "Electronics"
    keywords:
      - praktiker
      - electrolux
      - euronics
      
  Finance & Insurance:
    subcategory: "Service Charge"
    keywords:
      - hitel törlesztés
      - biztosítási díj
      - csomagdíj
      - tranzakciós költség
```

### Matching Logic Priority

1. **Exact match** - Full partner name matches `exact_matches` list
2. **Keyword match** - Partner name contains any keyword (case-insensitive)
3. **Transaction type fallback** - Based on K&H transaction type field
4. **Default** - "Uncategorized" if no match found

## LLM Prompt Format

When `update-config` detects new merchants, it generates this prompt:

```
=== PROMPT FOR LLM ===
Copy the text below and paste it into ChatGPT/Gemini:

---
I have a transaction categorization config in YAML format for my personal finance app.
I received new transactions that contain unknown merchants/partners.

CURRENT CONFIG:
---
[current categories.yaml content]
---

NEW UNCATEGORIZED MERCHANTS:
1. "EURONICS MU SZAKI A RU"
2. "SIMPLEP baloghekszer"
3. "VIDAMPARK ORIASKEREK"
4. "ELECTROLUX LEHEL HUTOG"
5. "GITHUB INC."

TASK:
1. Research each merchant if needed (use internet search)
2. Assign appropriate categories based on merchant business type
3. Add keywords to existing categories or suggest new category if needed
4. Return the COMPLETE updated YAML config with:
   - New merchants added to known_partners list
   - New keywords added to appropriate categories
   - Each category MUST have a 'subcategory' field with appropriate subcategory name
   
AVAILABLE CATEGORY NAMES (from ezBookkeeping defaults):
- Food & Drink
- Clothing & Appearance
- Housing & Houseware
- Transportation
- Communication
- Entertainment
- Education & Studying
- Medical & Healthcare
- Gift & Social
- Finance & Insurance
- Miscellaneous

IMPORTANT: Return ONLY the valid YAML config, no explanations or markdown code blocks.
---

=== END OF PROMPT ===

After you get the response from the LLM, save it to categories.yaml
```

## K&H Export Format

**File format:** TSV (Tab-Separated Values)
**Encoding:** UTF-8
**Date format:** YYYY.MM.DD
**Amount format:** Integer, negative for expenses, positive for income

**Columns (21 total):**
1. könyvelés dátuma - Booking date
2. tranzakció azonosító - Transaction ID
3. típus - Transaction type (e.g., "Vásárlás belföldi kereskedőnél", "Forint átutalás jóváírás")
4. könyvelési számla - Account number
5. könyvelési számla elnevezése - Account name
6. partner számla - Partner account
7. partner elnevezése - Partner name (merchant)
8. összeg - Amount
9. összeg devizaneme - Currency
10. közlemény - Description/note
11-21. Additional fields (mostly empty)

## ezBookkeeping CSV Format

**Required columns:**
- Type - "Income" / "Expense" / "Transfer"
- Category - Main category name (e.g., "Food & Drink")
- SubCategory - Subcategory name (e.g., "Food", "Drink")
- Account - Account name from `--account-name` flag
- Amount - Absolute value, decimal format
- DateTime - Format: "2006-01-02 15:04:05"
- Description - Transaction description
- Tags - Optional tags (can be empty)

## Development Notes

### Date Handling
- K&H provides only date, no time → append "00:00:00"
- Future-proof: If K&H adds time in format "YYYY.MM.DD HH:MM:SS", parse it

### Category Matching
- Case-insensitive keyword matching
- Hungarian characters supported (ő, ű, etc.)
- Partial string matching (e.g., "benzinkú" matches "SZEGED BENZINKÚ T AUT.")

### Error Handling
- Invalid dates → skip transaction with warning
- Invalid amounts → skip transaction with warning
- Missing config file → use empty categories (all "Uncategorized")
- Invalid YAML → clear error message

## Tech Stack

- **Language:** Go 1.25+
- **Dependencies:**
  - `gopkg.in/yaml.v3` - YAML parsing
  - Standard library for CSV/TSV handling
  
## Project Structure

```
ezbook-convert/
├── COPILOT.md              # This file
├── README.md               # User documentation
├── go.mod
├── main.go                 # CLI entry point
├── cmd/
│   ├── convert.go          # Convert command
│   └── update_config.go    # Update-config command
├── internal/
│   ├── parser/
│   │   └── kh.go          # K&H TSV parser
│   ├── converter/
│   │   └── ezbook.go      # ezBookkeeping converter
│   ├── config/
│   │   └── config.go      # YAML config handling
│   └── categorizer/
│       └── categorizer.go # Categorization logic
└── examples/
    └── categories.yaml     # Example config
```

## Example Workflow

1. User receives monthly K&H export: `kh_2025_11.csv`
2. First time setup:
   ```bash
   # Generate prompt for initial categorization
   ./ezbook-convert update-config --input kh_2025_11.csv --config categories.yaml
   
   # Copy prompt to ChatGPT, get YAML response, save to categories.yaml
   ```
3. Convert transactions:
   ```bash
   ./ezbook-convert convert \
     --input kh_2025_11.csv \
     --output ezbook_2025_11.csv \
     --account-name "K&H Checking" \
     --config categories.yaml
   ```
4. Import `ezbook_2025_11.csv` into ezBookkeeping
5. Next month: New merchants appear → run `update-config` again → update categories.yaml

## Future Enhancements (Out of Scope for v1)

- Support for other Hungarian banks (OTP, Erste, etc.)
- GUI for category mapping
- Direct integration with ezBookkeeping API
- Statistics/reports on categorized transactions
- Multi-account support in single export
