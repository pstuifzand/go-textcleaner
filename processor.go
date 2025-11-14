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
	operations := GetOperations()

	for _, op := range operations {
		if op.Name == operationName {
			return op.Func(input, arg1, arg2)
		}
	}

	return input
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
	return arg1 + input
}

func addSuffix(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}
	return input + arg1
}

func removePrefix(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}
	return strings.TrimPrefix(input, arg1)
}

func removeSuffix(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}
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
	return arg1 + input + arg2
}

// splitFormat splits text and reformats it
func splitFormat(input, arg1, arg2 string) string {
	if arg1 == "" || arg2 == "" {
		return input
	}

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

	// Normalize line breaks in replacement string
	replacement := strings.ReplaceAll(arg2, "\\n", "\n")
	replacement = strings.ReplaceAll(replacement, "\\r\\n", "\n")

	return re.ReplaceAllString(input, replacement)
}

// findHtmlLinks extracts HTML links
func findHtmlLinks(input, arg1, arg2 string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(input))
	if err != nil {
		return input
	}

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

// calculate evaluates mathematical expressions
func calculate(input, arg1, arg2 string) string {
	// Simple calculator supporting basic operations
	// This is a simplified version - the original uses a full expression parser

	// Try to evaluate simple expressions
	input = strings.TrimSpace(input)

	// Support basic operations: +, -, *, /
	result := evaluateExpression(input)
	if result != "" {
		return result
	}

	return input
}

// Helper functions

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

// evaluateExpression evaluates a simple mathematical expression
func evaluateExpression(expr string) string {
	expr = strings.TrimSpace(expr)

	// Handle simple number
	if num, err := strconv.ParseFloat(expr, 64); err == nil {
		return formatNumber(num)
	}

	// Try to parse and evaluate basic expressions
	// This is a very simplified calculator
	for _, op := range []string{"+", "-", "*", "/"} {
		if strings.Contains(expr, op) {
			parts := strings.Split(expr, op)
			if len(parts) == 2 {
				left, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
				right, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)

				if err1 == nil && err2 == nil {
					var result float64
					switch op {
					case "+":
						result = left + right
					case "-":
						result = left - right
					case "*":
						result = left * right
					case "/":
						if right != 0 {
							result = left / right
						} else {
							return expr
						}
					}
					return formatNumber(result)
				}
			}
		}
	}

	return ""
}

// formatNumber formats a number for display
func formatNumber(num float64) string {
	if math.Floor(num) == num {
		return fmt.Sprintf("%.0f", num)
	}
	return fmt.Sprintf("%g", num)
}
