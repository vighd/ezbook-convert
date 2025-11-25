# ezbook-convert

Convert K&H Bank (Hungary) transaction exports to ezBookkeeping-compatible CSV format.

## Quick Start

### 1. Build

```bash
go build -o ezbook-convert
```

### 2. First-time Setup: Generate Categories

```bash
# Generate LLM prompt to categorize merchants
./ezbook-convert update-config --input your_kh_export.csv --config categories.yaml

# Copy the generated prompt to ChatGPT or Gemini
# Save the LLM's YAML response to categories.yaml
```

### 3. Convert Transactions

```bash
./ezbook-convert convert \
  --input your_kh_export.csv \
  --output ezbook_output.csv \
  --account-name "K&H Checking Account" \
  --config categories.yaml
```

### 4. Import to ezBookkeeping

Upload `ezbook_output.csv` to your ezBookkeeping instance.

## Commands

### `convert`

Converts K&H TSV export to ezBookkeeping CSV format.

**Flags:**
- `--input` - Input K&H TSV file path (required)
- `--output` - Output ezBookkeeping CSV file path (required)
- `--account-name` - Account name for transactions (required)
- `--config` - YAML config file path (optional)

**Example:**
```bash
./ezbook-convert convert \
  --input kh_november.csv \
  --output ezbook_november.csv \
  --account-name "K&H Account" \
  --config categories.yaml
```

### `update-config`

Detects new merchants and generates an LLM prompt to update categorization.

**Flags:**
- `--input` - Input K&H TSV file path (required)
- `--config` - YAML config file path (default: categories.yaml)

**Example:**
```bash
./ezbook-convert update-config \
  --input kh_november.csv \
  --config categories.yaml
```

**Workflow:**
1. Run the command
2. Copy the generated prompt
3. Paste into ChatGPT (free tier) or Gemini
4. Save the LLM's YAML response to `categories.yaml`
5. Run `convert` command with updated config

## Configuration File

See `examples/categories.yaml` for a complete example.

**Structure:**
```yaml
# Track known merchants to detect new ones
known_partners:
  - "ALDI 241.SZ."
  - "AUCHAN SZEGED"

# Categorization rules
categories:
  Food & Drink:
    subcategory: "Food"  # Required: subcategory name
    keywords:
      - aldi
      - auchan
      - foodora
    exact_matches:
      - "ALDI 241.SZ."
      
  Transportation:
    subcategory: "Public Transit"
    keywords:
      - easypark
      - mol
      - benzinkút
```

**Matching Priority:**
1. Exact match (full partner name)
2. Keyword match (case-insensitive, partial)
3. Transaction type fallback
4. "Uncategorized" if no match

## K&H Export Format

K&H Bank exports transaction history in **TSV** (tab-separated) format.

**How to export from K&H:**
1. Log into K&H netbank
2. Go to Account History
3. Select date range
4. Download as CSV (it's actually TSV format)

**Expected format:**
- Tab-separated values
- Date format: `YYYY.MM.DD`
- 21 columns (only first 10 are used)
- Hungarian field names

## ezBookkeeping CSV Format

Output format compatible with ezBookkeeping import:

**Columns:**
- `Type` - Income / Expense / Transfer
- `Category` - Main category name (e.g., "Food & Drink")
- `SubCategory` - Subcategory name (e.g., "Food", "Drink")
- `Account` - Account name from `--account-name` flag
- `Amount` - Absolute value in decimal format
- `DateTime` - Format: `YYYY-MM-DD HH:MM:SS`
- `Description` - Partner name + transaction type + notes
- `Tags` - (currently empty, reserved for future use)

## Available Categories

Based on ezBookkeeping defaults:

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

You can create custom categories in your YAML config.

## Troubleshooting

### "All merchants are already categorized"

Good! No new merchants found. You can run `convert` directly.

### Missing categories in ezBookkeeping after import

Make sure the category names in your `categories.yaml` match the ones in your ezBookkeeping instance. You may need to create custom categories in ezBookkeeping first.

### Date parsing errors

K&H export should be in `YYYY.MM.DD` format. If you see errors, check the date column format in your export file.

### Encoding issues

K&H exports use UTF-8 encoding. If you see garbled Hungarian characters, ensure your terminal/editor supports UTF-8.

## Development

See [COPILOT.md](COPILOT.md) for detailed project documentation.

**Project structure:**
```
ezbook-convert/
├── main.go              # CLI entry point
├── cmd/                 # Command implementations
├── internal/
│   ├── parser/          # K&H TSV parser
│   ├── converter/       # ezBookkeeping converter
│   ├── config/          # YAML config handling
│   └── categorizer/     # Categorization logic
└── examples/            # Example configs
```

## License

MIT

## Contributing

PRs welcome! Please ensure:
- Code comments in English
- Follow existing code style
- Test with real K&H exports
