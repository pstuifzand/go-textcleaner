package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

// Operation represents a text transformation operation
type Operation struct {
	Name string
	Func func(input, arg1, arg2 string) string
}

// GetOperations returns all available text operations
func GetOperations() []Operation {
	return []Operation{
		{"Uppercase", uppercase},
		{"Lowercase", lowercase},
		{"Titlecase", titlecase},
		{"Trim", trim},
		{"Trim Left", trimLeft},
		{"Trim Right", trimRight},
		{"Replace Text", replaceText},
		{"HTML Decode", htmlDecode},
		{"HTML Encode", htmlEncode},
		{"Add Prefix", addPrefix},
		{"Add Suffix", addSuffix},
		{"Remove Prefix", removePrefix},
		{"Remove Suffix", removeSuffix},
		{"Left Characters", leftCharacters},
		{"Right Characters", rightCharacters},
		{"Mid Characters", midCharacters},
		{"Surround Text", surroundText},
		{"Split Format", splitFormat},
		{"Strip Tags", stripTags},
		{"Keep Match Lines", keepMatchLines},
		{"Remove Match Lines", removeMatchLines},
		{"Match Text", matchText},
		{"Replace Full", replaceFull},
		{"Find HTML Links", findHtmlLinks},
		{"Select HTML", selectHtml},
		{"Select JSON", selectJson},
		{"Calculate", calculate},
	}
}

// ProcessText processes the input text with the given operation and arguments
func ProcessText(input string, operationName string, arg1, arg2 string) string {
	return ProcessTextWithMode(input, operationName, arg1, arg2, false)
}

// ProcessTextWithMode processes the input text with the given operation, arguments, and mode
// If lineBased is true, the operation is applied to each line individually
func ProcessTextWithMode(input string, operationName string, arg1, arg2 string, lineBased bool) string {
	operations := GetOperations()

	for _, op := range operations {
		if op.Name == operationName {
			if lineBased {
				return applyLineBased(op.Func, input, arg1, arg2)
			}
			return op.Func(input, arg1, arg2)
		}
	}

	return input
}

// applyLineBased applies an operation to each line of the input text individually
func applyLineBased(opFunc func(input, arg1, arg2 string) string, input string, arg1, arg2 string) string {
	if input == "" {
		return input
	}

	lines := strings.Split(input, "\n")
	result := make([]string, len(lines))

	for i, line := range lines {
		result[i] = opFunc(line, arg1, arg2)
	}

	return strings.Join(result, "\n")
}

// Operation implementations

func uppercase(input, arg1, arg2 string) string {
	return strings.ToUpper(input)
}

func lowercase(input, arg1, arg2 string) string {
	return strings.ToLower(input)
}

func titlecase(input, arg1, arg2 string) string {
	return strings.Title(strings.ToLower(input))
}

func trim(input, arg1, arg2 string) string {
	return strings.TrimSpace(input)
}

func trimLeft(input, arg1, arg2 string) string {
	return strings.TrimLeftFunc(input, unicode.IsSpace)
}

func trimRight(input, arg1, arg2 string) string {
	return strings.TrimRightFunc(input, unicode.IsSpace)
}

func replaceText(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}
	arg1 = processEscapeSequences(arg1)
	arg2 = processEscapeSequences(arg2)
	return strings.ReplaceAll(input, arg1, arg2)
}

func htmlDecode(input, arg1, arg2 string) string {
	return html.UnescapeString(input)
}

func htmlEncode(input, arg1, arg2 string) string {
	return html.EscapeString(input)
}

func addPrefix(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}
	arg1 = processEscapeSequences(arg1)
	return arg1 + input
}

func addSuffix(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}
	arg1 = processEscapeSequences(arg1)
	return input + arg1
}

func removePrefix(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}
	arg1 = processEscapeSequences(arg1)
	return strings.TrimPrefix(input, arg1)
}

func removeSuffix(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}
	arg1 = processEscapeSequences(arg1)
	return strings.TrimSuffix(input, arg1)
}

// leftCharacters extracts a specified number of characters from the left
func leftCharacters(input, arg1, arg2 string) string {
	count, err := strconv.Atoi(arg1)
	if err != nil || count < 0 {
		return input
	}

	runes := []rune(input)
	if count > len(runes) {
		return input
	}
	return string(runes[:count])
}

// rightCharacters extracts a specified number of characters from the right
func rightCharacters(input, arg1, arg2 string) string {
	count, err := strconv.Atoi(arg1)
	if err != nil || count < 0 {
		return input
	}

	runes := []rune(input)
	if count > len(runes) {
		return input
	}
	return string(runes[len(runes)-count:])
}

// midCharacters extracts characters from the middle
func midCharacters(input, arg1, arg2 string) string {
	position, err1 := strconv.Atoi(arg1)
	count, err2 := strconv.Atoi(arg2)

	if err1 != nil || err2 != nil || position < 0 || count < 0 {
		return input
	}

	runes := []rune(input)
	if position >= len(runes) {
		return ""
	}

	endPos := position + count
	if endPos > len(runes) {
		endPos = len(runes)
	}

	return string(runes[position:endPos])
}

// surroundText wraps text with prefix and suffix
func surroundText(input, arg1, arg2 string) string {
	arg1 = processEscapeSequences(arg1)
	arg2 = processEscapeSequences(arg2)
	return arg1 + input + arg2
}

// splitFormat splits text and reformats it
func splitFormat(input, arg1, arg2 string) string {
	if arg1 == "" || arg2 == "" {
		return input
	}

	arg1 = processEscapeSequences(arg1)
	arg2 = processEscapeSequences(arg2)
	parts := strings.Split(input, arg1)

	// Convert parts to interface{} slice for fmt.Sprintf
	args := make([]interface{}, len(parts))
	for i, part := range parts {
		args[i] = part
	}

	// Try to format - if it fails, return original input
	defer func() {
		if r := recover(); r != nil {
			// Format string was invalid
		}
	}()

	result := fmt.Sprintf(arg2, args...)
	return result
}

// stripTags removes HTML/XML tags
func stripTags(input, arg1, arg2 string) string {
	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(input))
	if err != nil {
		// Fallback to simple regex-based tag stripping
		re := regexp.MustCompile(`<[^>]*>`)
		return html.UnescapeString(re.ReplaceAllString(input, ""))
	}

	// Remove script and style tags
	doc.Find("script, style").Remove()

	// Get text content
	return doc.Text()
}

// keepMatchLines keeps only lines matching the regex pattern
func keepMatchLines(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}

	// Parse regex options from arg2
	flags := parseRegexFlags(arg2)
	re, err := regexp.Compile(addRegexFlags(arg1, flags))
	if err != nil {
		return input
	}

	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(input))

	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return strings.TrimSuffix(result.String(), "\n")
}

// removeMatchLines removes lines matching the regex pattern
func removeMatchLines(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}

	// Parse regex options from arg2
	flags := parseRegexFlags(arg2)
	re, err := regexp.Compile(addRegexFlags(arg1, flags))
	if err != nil {
		return input
	}

	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(input))

	for scanner.Scan() {
		line := scanner.Text()
		if !re.MatchString(line) {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return strings.TrimSuffix(result.String(), "\n")
}

// matchText finds all matches of a regex pattern
func matchText(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}

	// Parse regex options from arg2
	flags := parseRegexFlags(arg2)
	re, err := regexp.Compile(addRegexFlags(arg1, flags))
	if err != nil {
		return input
	}

	matches := re.FindAllString(input, -1)
	if len(matches) == 0 {
		return ""
	}

	return strings.Join(matches, "\n")
}

// replaceFull performs regex replacement
func replaceFull(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}

	re, err := regexp.Compile(arg1)
	if err != nil {
		return input
	}

	// Process escape sequences in replacement string
	replacement := processEscapeSequences(arg2)

	return re.ReplaceAllString(input, replacement)
}

// findHtmlLinks extracts HTML links
func findHtmlLinks(input, arg1, arg2 string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(input))
	if err != nil {
		return input
	}

	// Process escape sequences in format string
	arg1 = processEscapeSequences(arg1)

	var result strings.Builder
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		text := strings.TrimSpace(s.Text())

		// Normalize whitespace
		text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

		// If format string provided, use it
		if arg1 != "" {
			formatted := strings.ReplaceAll(arg1, "{text}", text)
			formatted = strings.ReplaceAll(formatted, "{href}", href)
			result.WriteString(formatted)
			result.WriteString("\n")
		} else {
			// Default format
			result.WriteString(text)
			result.WriteString("\n")
			result.WriteString(href)
			result.WriteString("\n")
		}
	})

	return strings.TrimSuffix(result.String(), "\n")
}

// selectHtml selects HTML elements using CSS selectors
func selectHtml(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(input))
	if err != nil {
		return input
	}

	var result strings.Builder
	selection := doc.Find(arg1)

	if selection.Length() == 0 {
		return input
	}

	// Parse output format from arg2
	commands := strings.Split(arg2, "|")
	if len(commands) == 0 || arg2 == "" {
		commands = []string{"text"}
	}

	selection.Each(func(i int, s *goquery.Selection) {
		for _, cmd := range commands {
			cmd = strings.TrimSpace(cmd)
			switch {
			case cmd == "outer":
				html, _ := s.Html()
				result.WriteString("<" + goquery.NodeName(s) + ">" + html + "</" + goquery.NodeName(s) + ">")
				result.WriteString("\n")
			case cmd == "inner":
				html, _ := s.Html()
				result.WriteString(html)
				result.WriteString("\n")
			case cmd == "text":
				result.WriteString(s.Text())
				result.WriteString("\n")
			case strings.HasPrefix(cmd, "attr:"):
				attrName := strings.TrimPrefix(cmd, "attr:")
				if attr, exists := s.Attr(attrName); exists {
					result.WriteString(attr)
					result.WriteString("\n")
				}
			}
		}
	})

	return strings.TrimSuffix(result.String(), "\n")
}

// selectJson extracts JSON data using a simple path notation
func selectJson(input, arg1, arg2 string) string {
	var data interface{}
	err := json.Unmarshal([]byte(input), &data)
	if err != nil {
		return input
	}

	// If no path specified, return formatted JSON
	if arg1 == "" {
		output, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return input
		}
		return string(output)
	}

	// Simple path navigation (supports dot notation)
	parts := strings.Split(arg1, ".")
	current := data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		case []interface{}:
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(v) {
				return input
			}
			current = v[idx]
		default:
			return input
		}

		if current == nil {
			return ""
		}
	}

	// Convert result to string
	output, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", current)
	}
	return string(output)
}

// calculate evaluates mathematical expressions found in text
func calculate(input, arg1, arg2 string) string {
	if input == "" {
		return input
	}

	// Find all mathematical expressions in the text and evaluate them
	// Expressions are patterns like: number op number op number ...
	// Examples: "6 + 5", "5 * 55 + 3", "-10 + 5.5", "Item 1 = 6 + 5"

	result := input

	// Regular expression to match mathematical expressions
	// Matches: optional spaces, optional minus, number, then (operator number) repeated
	// This pattern captures complete mathematical expressions embedded in text
	re := regexp.MustCompile(`(-?\d+(?:\.\d+)?(?:\s*[+\-*/]\s*-?\d+(?:\.\d+)?)*)`)

	result = re.ReplaceAllStringFunc(result, func(match string) string {
		// Evaluate each matched expression
		if value, err := evaluateExpression(match); err == nil {
			return formatNumber(value)
		}
		return match
	})

	return result
}

// Helper functions

// processEscapeSequences converts escape sequences in a string
// Handles: \n, \r, \t, \f, \v, \b, \a, \\, and \xHH hex escapes
func processEscapeSequences(s string) string {
	var result strings.Builder
	runes := []rune(s)

	for i := 0; i < len(runes); i++ {
		if runes[i] == '\\' && i+1 < len(runes) {
			nextChar := runes[i+1]
			switch nextChar {
			case 'n':
				result.WriteRune('\n')
				i++
			case 'r':
				result.WriteRune('\r')
				i++
			case 't':
				result.WriteRune('\t')
				i++
			case 'f':
				result.WriteRune('\f')
				i++
			case 'v':
				result.WriteRune('\v')
				i++
			case 'b':
				result.WriteRune('\b')
				i++
			case 'a':
				result.WriteRune('\a')
				i++
			case '\\':
				result.WriteRune('\\')
				i++
			case 'x':
				// Handle \xHH hex escape
				if i+3 < len(runes) {
					hexStr := string(runes[i+2 : i+4])
					if val, err := strconv.ParseInt(hexStr, 16, 8); err == nil {
						result.WriteRune(rune(val))
						i += 3
					} else {
						result.WriteRune(runes[i])
					}
				} else {
					result.WriteRune(runes[i])
				}
			case 'u':
				// Handle \uHHHH unicode escape
				if i+5 < len(runes) {
					hexStr := string(runes[i+2 : i+6])
					if val, err := strconv.ParseInt(hexStr, 16, 32); err == nil {
						result.WriteRune(rune(val))
						i += 5
					} else {
						result.WriteRune(runes[i])
					}
				} else {
					result.WriteRune(runes[i])
				}
			case 'U':
				// Handle \UHHHHHHHH unicode escape
				if i+9 < len(runes) {
					hexStr := string(runes[i+2 : i+10])
					if val, err := strconv.ParseInt(hexStr, 16, 32); err == nil {
						result.WriteRune(rune(val))
						i += 9
					} else {
						result.WriteRune(runes[i])
					}
				} else {
					result.WriteRune(runes[i])
				}
			default:
				// Not a recognized escape sequence, keep the backslash
				result.WriteRune(runes[i])
			}
		} else {
			result.WriteRune(runes[i])
		}
	}

	return result.String()
}

// parseRegexFlags parses regex flag string (e.g., "ims")
func parseRegexFlags(flags string) map[rune]bool {
	result := make(map[rune]bool)
	for _, ch := range flags {
		result[ch] = true
	}
	return result
}

// addRegexFlags adds regex flags to pattern
func addRegexFlags(pattern string, flags map[rune]bool) string {
	prefix := "(?m)" // Always use multiline mode for line operations
	if flags['i'] {
		prefix += "(?i)"
	}
	if flags['s'] {
		prefix = "(?s)" + prefix[4:] // Replace multiline with singleline
	}
	return prefix + pattern
}

// Token represents a token in the expression
type Token struct {
	Type  string  // "number", "operator", "eof"
	Value float64 // for numbers
	Op    rune    // for operators: +, -, *, /
}

// Tokenizer tokenizes a mathematical expression
type Tokenizer struct {
	expr      string
	pos       int
	ch        rune
	lastToken Token // Track last token to distinguish "-" operator from negative sign
}

// NewTokenizer creates a new tokenizer
func NewTokenizer(expr string) *Tokenizer {
	expr = strings.TrimSpace(expr)
	t := &Tokenizer{expr: expr, pos: 0, lastToken: Token{Type: "eof"}}
	if len(expr) > 0 {
		t.ch = rune(expr[0])
	}
	return t
}

// NextToken returns the next token from the expression
func (t *Tokenizer) NextToken() Token {
	// Skip whitespace
	for t.pos < len(t.expr) && unicode.IsSpace(t.ch) {
		t.advance()
	}

	if t.pos >= len(t.expr) {
		return Token{Type: "eof"}
	}

	// Determine if "-" should be treated as negative sign or operator
	// "-" is a negative sign only if the last token was an operator or EOF (start of expression)
	isNegativeSign := t.ch == '-' && (t.lastToken.Type == "operator" || t.lastToken.Type == "eof")

	// Handle numbers and negative numbers
	if unicode.IsDigit(t.ch) || (isNegativeSign && t.pos+1 < len(t.expr) && (unicode.IsDigit(rune(t.expr[t.pos+1])) || rune(t.expr[t.pos+1]) == '.')) ||
		(t.ch == '.' && t.pos+1 < len(t.expr) && unicode.IsDigit(rune(t.expr[t.pos+1]))) {
		return t.readNumber()
	}

	// Handle operators
	if t.ch == '+' || t.ch == '-' || t.ch == '*' || t.ch == '/' {
		op := t.ch
		t.advance()
		token := Token{Type: "operator", Op: op}
		t.lastToken = token
		return token
	}

	// Unknown character - skip it
	t.advance()
	return t.NextToken()
}

// readNumber reads a number from the expression
func (t *Tokenizer) readNumber() Token {
	start := t.pos

	// Handle negative sign
	if t.ch == '-' {
		t.advance()
	}

	// Read integer part
	for t.pos < len(t.expr) && unicode.IsDigit(t.ch) {
		t.advance()
	}

	// Read decimal part
	if t.pos < len(t.expr) && t.ch == '.' {
		t.advance()
		for t.pos < len(t.expr) && unicode.IsDigit(t.ch) {
			t.advance()
		}
	}

	numStr := t.expr[start:t.pos]
	num, _ := strconv.ParseFloat(numStr, 64)
	token := Token{Type: "number", Value: num}
	t.lastToken = token
	return token
}

// advance moves to the next character
func (t *Tokenizer) advance() {
	t.pos++
	if t.pos < len(t.expr) {
		t.ch = rune(t.expr[t.pos])
	}
}

// Parser parses and evaluates mathematical expressions
type Parser struct {
	tokenizer *Tokenizer
	current   Token
}

// NewParser creates a new parser
func NewParser(expr string) *Parser {
	t := NewTokenizer(expr)
	p := &Parser{tokenizer: t}
	p.current = p.tokenizer.NextToken()
	return p
}

// Parse parses and evaluates the expression
func (p *Parser) Parse() (float64, error) {
	if p.current.Type == "eof" {
		return 0, fmt.Errorf("empty expression")
	}
	result, err := p.parseAddSub()
	if p.current.Type != "eof" {
		return 0, fmt.Errorf("unexpected token after expression")
	}
	return result, err
}

// parseAddSub parses addition and subtraction (lowest precedence)
func (p *Parser) parseAddSub() (float64, error) {
	left, err := p.parseMulDiv()
	if err != nil {
		return 0, err
	}

	for p.current.Type == "operator" && (p.current.Op == '+' || p.current.Op == '-') {
		op := p.current.Op
		p.current = p.tokenizer.NextToken()
		right, err := p.parseMulDiv()
		if err != nil {
			return 0, err
		}
		if op == '+' {
			left = left + right
		} else {
			left = left - right
		}
	}

	return left, nil
}

// parseMulDiv parses multiplication and division (highest precedence)
func (p *Parser) parseMulDiv() (float64, error) {
	left, err := p.parseNumber()
	if err != nil {
		return 0, err
	}

	for p.current.Type == "operator" && (p.current.Op == '*' || p.current.Op == '/') {
		op := p.current.Op
		p.current = p.tokenizer.NextToken()
		right, err := p.parseNumber()
		if err != nil {
			return 0, err
		}
		if op == '*' {
			left = left * right
		} else {
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			left = left / right
		}
	}

	return left, nil
}

// parseNumber parses a number or a negative number
func (p *Parser) parseNumber() (float64, error) {
	if p.current.Type == "number" {
		num := p.current.Value
		p.current = p.tokenizer.NextToken()
		return num, nil
	}

	if p.current.Type == "operator" && p.current.Op == '-' {
		p.current = p.tokenizer.NextToken()
		num, err := p.parseNumber()
		if err != nil {
			return 0, err
		}
		return -num, nil
	}

	return 0, fmt.Errorf("expected number")
}

// evaluateExpression evaluates a mathematical expression and returns the result
func evaluateExpression(expr string) (float64, error) {
	parser := NewParser(expr)
	return parser.Parse()
}

// formatNumber formats a number for display
func formatNumber(num float64) string {
	// Handle special values
	if math.IsNaN(num) || math.IsInf(num, 0) {
		return fmt.Sprintf("%v", num)
	}

	// Remove trailing zeros and decimal point if not needed
	if math.Floor(num) == num {
		return fmt.Sprintf("%.0f", num)
	}

	// Format with up to 10 decimal places, removing trailing zeros
	formatted := fmt.Sprintf("%.10f", num)
	formatted = strings.TrimRight(formatted, "0")
	formatted = strings.TrimRight(formatted, ".")

	return formatted
}
