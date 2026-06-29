package scraper

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// TableParser provides BeautifulSoup-style HTML table parsing.
type TableParser struct {
	client *http.Client
}

// NewTableParser creates a table parser with default HTTP client.
func NewTableParser() *TableParser {
	return &TableParser{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Row represents a single table row.
type Row struct {
	Cells []string
}

// Table represents a parsed HTML table.
type Table struct {
	Headers []string
	Rows    []Row
}

// ParseFromURL fetches and parses all tables from a URL.
func (p *TableParser) ParseFromURL(url string) ([]Table, error) {
	resp, err := p.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	return p.ParseDocument(doc)
}

// ParseFromString parses tables from an HTML string.
func (p *TableParser) ParseFromString(htmlStr string) ([]Table, error) {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}
	return p.ParseDocument(doc)
}

// ParseDocument parses all tables from an html.Node tree.
func (p *TableParser) ParseDocument(doc *html.Node) ([]Table, error) {
	var tables []Table

	var visitNode func(*html.Node)
	visitNode = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			table := p.parseTable(n)
			if table != nil && (len(table.Headers) > 0 || len(table.Rows) > 0) {
				tables = append(tables, *table)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visitNode(c)
		}
	}

	visitNode(doc)
	return tables, nil
}

// parseTable extracts headers and rows from a table node.
func (p *TableParser) parseTable(node *html.Node) *Table {
	table := &Table{}
	headersParsed := false

	var visitNode func(*html.Node)
	visitNode = func(n *html.Node) {
		if n.Type != html.ElementNode {
			return
		}

		switch n.Data {
		case "thead":
			p.parseHeaderRows(n, table)
			headersParsed = true
		case "tbody", "tfoot", "":
			if !headersParsed {
				p.parseHeaderRows(n, table)
				headersParsed = true
			}
			p.parseDataRows(n, table)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visitNode(c)
		}
	}

	visitNode(node)
	return table
}

// parseHeaderRows extracts header cells from a node.
func (p *TableParser) parseHeaderRows(node *html.Node, table *Table) {
	var visitNode func(*html.Node)
	visitNode = func(n *html.Node) {
		if n.Type != html.ElementNode {
			return
		}

		if n.Data == "tr" {
			row := p.parseRow(n)
			if len(row.Cells) > 0 {
				table.Headers = append(table.Headers, strings.Join(row.Cells, "|"))
			}
			return
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visitNode(c)
		}
	}

	visitNode(node)
}

// parseDataRows extracts data rows from a node.
func (p *TableParser) parseDataRows(node *html.Node, table *Table) {
	var visitNode func(*html.Node)
	visitNode = func(n *html.Node) {
		if n.Type != html.ElementNode {
			return
		}

		if n.Data == "tr" {
			row := p.parseRow(n)
			if len(row.Cells) > 0 {
				table.Rows = append(table.Rows, row)
			}
			return
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visitNode(c)
		}
	}

	visitNode(node)
}

// parseRow extracts cell text from a tr node.
func (p *TableParser) parseRow(node *html.Node) Row {
	var cells []string

	var visitNode func(*html.Node)
	visitNode = func(n *html.Node) {
		if n.Type != html.ElementNode {
			return
		}

		if n.Data == "th" || n.Data == "td" {
			cells = append(cells, p.extractText(n))
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visitNode(c)
		}
	}

	visitNode(node)
	return Row{Cells: cells}
}

// extractText gets all text content from a node.
func (p *TableParser) extractText(node *html.Node) string {
	var textParts []string

	var visitNode func(*html.Node)
	visitNode = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				textParts = append(textParts, text)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visitNode(c)
		}
	}

	visitNode(node)
	return strings.Join(textParts, " ")
}

// FindTableByKeyword finds the first table whose headers contain a keyword.
func (p *TableParser) FindTableByKeyword(tables []Table, keyword string) (*Table, error) {
	for i, t := range tables {
		for _, h := range t.Headers {
			if strings.Contains(strings.ToLower(h), strings.ToLower(keyword)) {
				return &tables[i], nil
			}
		}
	}
	return nil, fmt.Errorf("no table found with keyword: %s", keyword)
}

// ToCSV converts a table to CSV format.
func (t *Table) ToCSV() string {
	var lines []string
	if len(t.Headers) > 0 {
		lines = append(lines, strings.Join(t.Headers, ","))
	}
	for _, row := range t.Rows {
		lines = append(lines, strings.Join(row.Cells, ","))
	}
	return strings.Join(lines, "\n")
}

// ToJSON converts a table to a slice of maps (one per row).
func (t *Table) ToJSON() []map[string]string {
	if len(t.Headers) == 0 {
		return nil
	}

	headerCount := len(t.Headers)
	headers := strings.Split(t.Headers[0], "|")
	if len(headers) != headerCount {
		headers = make([]string, headerCount)
		for i := range headers {
			headers[i] = fmt.Sprintf("col%d", i)
		}
	}

	result := make([]map[string]string, len(t.Rows))
	for i, row := range t.Rows {
		entry := make(map[string]string)
		for j, cell := range row.Cells {
			key := "col"
			if j < len(headers) {
				key = headers[j]
			}
			if key == "" {
				key = fmt.Sprintf("col%d", j)
			}
			entry[key] = cell
		}
		result[i] = entry
	}
	return result
}
