package main

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"math"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

// Operation represents a text transformation operation
type Operation struct {
	Name string
	Func func(input, arg1, arg2 string) string
}

// PipelineNode represents a node in the tree-based operation pipeline
type PipelineNode struct {
	ID           string          `json:"id"`             // Unique identifier
	Type         string          `json:"type"`           // "operation", "if", "foreach", "group"
	Name         string          `json:"name"`           // Display name
	Operation    string          `json:"operation"`      // Operation name (for type="operation")
	Arg1         string          `json:"arg1"`           // First argument
	Arg2         string          `json:"arg2"`           // Second argument
	Condition    string          `json:"condition"`      // For if nodes: regex/pattern to test
	Children     []PipelineNode  `json:"children"`       // Child nodes
	ElseChildren []PipelineNode  `json:"else_children"`  // For if nodes: else branch
}

// GetOperations returns all available text operations
func GetOperations() []Operation {
	return []Operation{
		// Identity (no-op) operation
		{"Identity", identity},

		// Basic case operations
		{"Uppercase", uppercase},
		{"Lowercase", lowercase},
		{"Titlecase", titlecase},

		// Whitespace operations
		{"Trim", trim},
		{"Trim Left", trimLeft},
		{"Trim Right", trimRight},
		{"Normalize Whitespace", normalizeWhitespace},

		// Basic string operations
		{"Replace Text", replaceText},
		{"Add Prefix", addPrefix},
		{"Add Suffix", addSuffix},
		{"Remove Prefix", removePrefix},
		{"Remove Suffix", removeSuffix},
		{"Surround Text", surroundText},

		// Character extraction
		{"Left Characters", leftCharacters},
		{"Right Characters", rightCharacters},
		{"Mid Characters", midCharacters},

		// Text manipulation
		{"Split Format", splitFormat},

		// HTML operations
		{"HTML Decode", htmlDecode},
		{"HTML Encode", htmlEncode},
		{"Strip Tags", stripTags},
		{"Find HTML Links", findHtmlLinks},
		{"Select HTML", selectHtml},

		// JSON operations
		{"Select JSON", selectJson},

		// Regex operations
		{"Keep Match Lines", keepMatchLines},
		{"Remove Match Lines", removeMatchLines},
		{"Match Text", matchText},
		{"Replace Full", replaceFull},

		// Math operations
		{"Calculate", calculate},

		// Phase 1: Line Operations
		{"Sort Lines", sortLines},
		{"Number Lines", numberLines},
		{"Randomize Lines", randomizeLines},
		{"Invert Lines", invertLines},
		{"Deduplicate Lines", deduplicateLines},
		{"Filter Blank Lines", filterBlankLines},
		{"Filter Lines by Length", filterLinesByLength},

		// Phase 2: Text Formatting
		{"Wrap Text", wrapText},
		{"Rewrap Text", rewrapText},
		{"Make Paragraphs", makeParagraphs},
		{"Quote Text", quoteText},
		{"Indent Text", indentText},
		{"Unindent Text", unindentText},
		{"Center Text", centerText},

		// Phase 3: Case & Characters
		{"Capitalize Sentences", capitalizeSentences},
		{"Randomcase", randomcase},
		{"Strip Diacritics", stripDiacritics},
		{"Reverse Text", reverseText},
		{"Reverse Words", reverseWords},
		{"Reverse Lines", reverseLines},
		{"Slugify", slugify},
		{"Smart Quotes", smartQuotes},
		{"Straight Quotes", straightQuotes},

		// Phase 4: Encoding & Dates
		{"Base64 Encode", base64Encode},
		{"Base64 Decode", base64Decode},
		{"URL Encode", urlEncode},
		{"URL Decode", urlDecode},
		{"Hex Encode", hexEncode},
		{"Hex Decode", hexDecode},
		{"ROT13", rot13},
		{"Escape Quotes", escapeQuotes},
		{"Unescape Quotes", unescapeQuotes},
		{"Insert Date/Time", insertDateTime},

		// Phase 5: Markdown/HTML
		{"URLs to Hyperlinks", urlsToHyperlinks},
		{"Extract URLs", extractUrls},
		{"Extract Emails", extractEmails},
		{"Extract Numbers", extractNumbers},

		// Phase 6: Advanced Regex
		{"Extract with Groups", extractWithGroups},
		{"Replace with Groups", replaceWithGroups},
		{"Split by Regex", splitByRegex},
		{"Match Count", matchCount},

		// Phase 7: Math & Numbers
		{"Format Numbers", formatNumbersOperation},
		{"Round Numbers", roundNumbers},
		{"Sum Numbers", sumNumbers},

		// Phase 8: List & Extraction
		{"Join List", joinList},
		{"Remove Control Characters", removeControlCharacters},
		{"Count Occurrences", countOccurrences},
		{"Keep Lines Containing", keepLinesContaining},
		{"Remove Lines Containing", removeLinesContaining},
		{"Truncate Text", truncateText},

		// Phase 9: Conditional Operations
		{"Is Empty", isEmpty},
		{"Has Pattern", hasPattern},
		{"Starts With", startsWith},

		// Phase 10: List Processing
		{"Unique Values", uniqueValues},
		{"Most Common", mostCommon},
		{"Least Common", leastCommon},
		{"Reverse Lines", reverseLinesOrder},
		{"Group By Pattern", groupByPattern},

		// Phase 11: Advanced Text Operations
		{"Word Count", wordCount},
		{"Character Count", characterCount},
		{"Line Count", lineCount},
		{"Text Statistics", textStatistics},
		{"Min Word Length", minWordLength},
		{"Max Word Length", maxWordLength},
		{"Average Word Length", averageWordLength},

		// Phase 12: Advanced Pattern Operations
		{"Whole Word Match", wholeWordMatch},
		{"Case Sensitive Find", caseSensitiveFind},
		{"Multi-line Pattern", multilinePattern},
		{"Look-ahead Pattern", lookaheadPattern},
		{"Look-behind Pattern", lookbehindPattern},
		{"Conditional Replace", conditionalReplace},

		// Phase 13: Transformation Macros
		{"Chain Operations", chainOperations},
		{"Repeat Operation", repeatOperation},
		{"Swap Pairs", swapPairs},
		{"Reverse Order Items", reverseOrderItems},

		// Phase 14: HTML/Markdown Advanced
		{"HTML to Markdown", htmlToMarkdown},
		{"Markdown to HTML", markdownToHTML},
		{"Extract Text from HTML", extractTextFromHTML},
		{"Create Markdown Table", createMarkdownTable},
		{"Parse YAML Front Matter", parseYAMLFrontMatter},
		{"Markdown Link Format", markdownLinkFormat},

		// Phase 15: Unicode & Special Characters
		{"Unicode Names", unicodeNames},
		{"Convert Unicode Escapes", convertUnicodeEscapes},
		{"Escape Unicode", escapeUnicode},
		{"Show Invisible Characters", showInvisibleCharacters},
		{"Normalize Unicode", normalizeUnicode},
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

// ExecuteNode executes a pipeline node (tree-based execution)
func ExecuteNode(node *PipelineNode, input string) string {
	if node == nil {
		return input
	}

	switch node.Type {
	case "operation":
		return executeOperationNode(node, input)
	case "if":
		return executeIfNode(node, input)
	case "foreach":
		return executeForEachNode(node, input)
	case "group":
		return executeGroupNode(node, input)
	default:
		// Default to sequence behavior
		return executeSequenceNode(node, input)
	}
}

// executeOperationNode executes a single operation and then its children
func executeOperationNode(node *PipelineNode, input string) string {
	// Execute the operation
	result := ProcessText(input, node.Operation, node.Arg1, node.Arg2)

	// Execute children on the result
	return executeSequenceNode(&PipelineNode{Children: node.Children}, result)
}

// executeIfNode executes children based on condition match
func executeIfNode(node *PipelineNode, input string) string {
	if node.Condition == "" {
		return input
	}

	// Try to match the condition (treat as regex)
	re, err := regexp.Compile(node.Condition)
	var matches bool

	if err != nil {
		// If invalid regex, treat as literal string match
		matches = strings.Contains(input, node.Condition)
	} else {
		matches = re.MatchString(input)
	}

	// Execute appropriate branch
	if matches {
		return executeSequenceNode(&PipelineNode{Children: node.Children}, input)
	} else {
		return executeSequenceNode(&PipelineNode{Children: node.ElseChildren}, input)
	}
}

// executeForEachNode applies children to each line
func executeForEachNode(node *PipelineNode, input string) string {
	if input == "" {
		return input
	}

	lines := strings.Split(input, "\n")
	result := make([]string, len(lines))

	for i, line := range lines {
		// Execute all children on this line
		result[i] = executeSequenceNode(&PipelineNode{Children: node.Children}, line)
	}

	return strings.Join(result, "\n")
}

// executeGroupNode passes through to children (no operation logic)
func executeGroupNode(node *PipelineNode, input string) string {
	return executeSequenceNode(&PipelineNode{Children: node.Children}, input)
}

// executeSequenceNode executes children in sequence, piping output through each
func executeSequenceNode(node *PipelineNode, input string) string {
	result := input

	for i := range node.Children {
		result = ExecuteNode(&node.Children[i], result)
	}

	return result
}

// Operation implementations

// identity returns the input unchanged (no-op operation)
func identity(input, arg1, arg2 string) string {
	return input
}

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

// Phase 1: Line Operations

// sortLines sorts the lines of text
// arg1: sort options (n=numeric, r=reverse, i=case-insensitive)
func sortLines(input, arg1, arg2 string) string {
	if input == "" {
		return input
	}

	lines := strings.Split(input, "\n")

	// Parse options
	numeric := strings.Contains(arg1, "n")
	reverse := strings.Contains(arg1, "r")
	caseInsensitive := strings.Contains(arg1, "i")

	sort.Slice(lines, func(i, j int) bool {
		a, b := lines[i], lines[j]

		if caseInsensitive {
			a = strings.ToLower(a)
			b = strings.ToLower(b)
		}

		if numeric {
			// Extract leading numbers if present
			numA := extractLeadingNumber(a)
			numB := extractLeadingNumber(b)
			if numA != nil && numB != nil {
				less := *numA < *numB
				if reverse {
					return !less
				}
				return less
			}
		}

		less := a < b
		if reverse {
			return !less
		}
		return less
	})

	return strings.Join(lines, "\n")
}

// numberLines adds line numbers to each line
// arg1: starting number (default 1)
// arg2: format string (default "%d. ")
func numberLines(input, arg1, arg2 string) string {
	if input == "" {
		return input
	}

	startNum := 1
	if arg1 != "" {
		if n, err := strconv.Atoi(arg1); err == nil {
			startNum = n
		}
	}

	format := "%d. "
	if arg2 != "" {
		format = arg2
	}

	lines := strings.Split(input, "\n")
	result := make([]string, len(lines))

	for i, line := range lines {
		result[i] = fmt.Sprintf(format, startNum+i) + line
	}

	return strings.Join(result, "\n")
}

// randomizeLines shuffles the lines randomly
func randomizeLines(input, arg1, arg2 string) string {
	if input == "" {
		return input
	}

	lines := strings.Split(input, "\n")

	// Fisher-Yates shuffle using time-based seed
	for i := len(lines) - 1; i > 0; i-- {
		j := int(time.Now().UnixNano()) % (i + 1)
		lines[i], lines[j] = lines[j], lines[i]
	}

	return strings.Join(lines, "\n")
}

// invertLines reverses the order of lines
func invertLines(input, arg1, arg2 string) string {
	if input == "" {
		return input
	}

	lines := strings.Split(input, "\n")

	// Reverse the order
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}

	return strings.Join(lines, "\n")
}

// deduplicateLines removes duplicate lines, keeping the first occurrence
func deduplicateLines(input, arg1, arg2 string) string {
	if input == "" {
		return input
	}

	lines := strings.Split(input, "\n")
	seen := make(map[string]bool)
	result := []string{}

	for _, line := range lines {
		if !seen[line] {
			seen[line] = true
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// filterBlankLines removes empty or whitespace-only lines
func filterBlankLines(input, arg1, arg2 string) string {
	if input == "" {
		return input
	}

	lines := strings.Split(input, "\n")
	result := []string{}

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// filterLinesByLength filters lines by length
// arg1: minimum length (0 if not specified)
// arg2: maximum length (no limit if not specified)
func filterLinesByLength(input, arg1, arg2 string) string {
	if input == "" {
		return input
	}

	minLen := 0
	maxLen := -1

	if arg1 != "" {
		if n, err := strconv.Atoi(arg1); err == nil && n >= 0 {
			minLen = n
		}
	}

	if arg2 != "" {
		if n, err := strconv.Atoi(arg2); err == nil && n >= 0 {
			maxLen = n
		}
	}

	lines := strings.Split(input, "\n")
	result := []string{}

	for _, line := range lines {
		len := len([]rune(line))
		if len >= minLen && (maxLen < 0 || len <= maxLen) {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// Phase 2: Text Formatting

// wrapText wraps text at a specified column width
// arg1: column width (default 80)
func wrapText(input, arg1, arg2 string) string {
	width := 80
	if arg1 != "" {
		if w, err := strconv.Atoi(arg1); err == nil && w > 0 {
			width = w
		}
	}

	words := strings.Fields(input)
	if len(words) == 0 {
		return input
	}

	var result strings.Builder
	lineLen := 0

	for _, word := range words {
		wordLen := len([]rune(word))

		if lineLen == 0 {
			result.WriteString(word)
			lineLen = wordLen
		} else if lineLen+1+wordLen <= width {
			result.WriteString(" ")
			result.WriteString(word)
			lineLen += 1 + wordLen
		} else {
			result.WriteString("\n")
			result.WriteString(word)
			lineLen = wordLen
		}
	}

	return result.String()
}

// rewrapText unwraps text and then rewraps at specified width
// arg1: column width (default 80)
func rewrapText(input, arg1, arg2 string) string {
	// First unwrap by replacing newlines with spaces
	unwrapped := strings.ReplaceAll(input, "\n", " ")
	unwrapped = normalizeWhitespace(unwrapped, "", "")

	// Then wrap at the specified width
	return wrapText(unwrapped, arg1, arg2)
}

// makeParagraphs joins lines into paragraphs separated by blank lines
func makeParagraphs(input, arg1, arg2 string) string {
	if input == "" {
		return input
	}

	lines := strings.Split(input, "\n")
	var paragraphs []string
	var currentPara []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			if len(currentPara) > 0 {
				paragraphs = append(paragraphs, strings.Join(currentPara, " "))
				currentPara = []string{}
			}
		} else {
			currentPara = append(currentPara, strings.TrimSpace(line))
		}
	}

	if len(currentPara) > 0 {
		paragraphs = append(paragraphs, strings.Join(currentPara, " "))
	}

	return strings.Join(paragraphs, "\n\n")
}

// quoteText adds a prefix to each line (like "> " for blockquote)
// arg1: prefix string (default "> ")
func quoteText(input, arg1, arg2 string) string {
	if input == "" {
		return input
	}

	prefix := "> "
	if arg1 != "" {
		prefix = arg1
	}

	lines := strings.Split(input, "\n")
	result := make([]string, len(lines))

	for i, line := range lines {
		result[i] = prefix + line
	}

	return strings.Join(result, "\n")
}

// indentText adds indentation to each line
// arg1: indentation string (default "    " - 4 spaces)
func indentText(input, arg1, arg2 string) string {
	return quoteText(input, arg1, arg2)
}

// unindentText removes common leading whitespace
func unindentText(input, arg1, arg2 string) string {
	if input == "" {
		return input
	}

	lines := strings.Split(input, "\n")

	// Find minimum indentation
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		indent := 0
		for _, ch := range line {
			if ch == ' ' || ch == '\t' {
				indent++
			} else {
				break
			}
		}

		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return input
	}

	// Remove the minimum indentation
	result := make([]string, len(lines))
	for i, line := range lines {
		if len(line) >= minIndent {
			result[i] = line[minIndent:]
		} else {
			result[i] = line
		}
	}

	return strings.Join(result, "\n")
}

// centerText centers each line within a specified width
// arg1: width (default 80)
func centerText(input, arg1, arg2 string) string {
	width := 80
	if arg1 != "" {
		if w, err := strconv.Atoi(arg1); err == nil && w > 0 {
			width = w
		}
	}

	lines := strings.Split(input, "\n")
	result := make([]string, len(lines))

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		lineLen := len([]rune(trimmed))

		if lineLen >= width {
			result[i] = trimmed
		} else {
			padding := (width - lineLen) / 2
			result[i] = strings.Repeat(" ", padding) + trimmed
		}
	}

	return strings.Join(result, "\n")
}

// Phase 3: Case & Character Operations

// capitalizeSentences capitalizes the first letter of each sentence
func capitalizeSentences(input, arg1, arg2 string) string {
	if input == "" {
		return input
	}

	result := []rune{}
	capitalizeNext := true

	for _, r := range input {
		if capitalizeNext && unicode.IsLetter(r) {
			result = append(result, unicode.ToUpper(r))
			capitalizeNext = false
		} else if r == '.' || r == '!' || r == '?' {
			result = append(result, r)
			capitalizeNext = true
		} else {
			result = append(result, r)
		}
	}

	return string(result)
}

// randomcase randomly capitalizes or lowercases each letter
func randomcase(input, arg1, arg2 string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) {
			if int(time.Now().UnixNano())%2 == 0 {
				return unicode.ToUpper(r)
			}
			return unicode.ToLower(r)
		}
		return r
	}, input)
}

// stripDiacritics removes diacritical marks from characters
// This is a simple version that removes common diacritics
func stripDiacritics(input, arg1, arg2 string) string {
	// Map of accented characters to their base forms
	replacements := map[rune]rune{
		'á': 'a', 'à': 'a', 'ä': 'a', 'â': 'a', 'ã': 'a', 'å': 'a',
		'é': 'e', 'è': 'e', 'ë': 'e', 'ê': 'e',
		'í': 'i', 'ì': 'i', 'ï': 'i', 'î': 'i',
		'ó': 'o', 'ò': 'o', 'ö': 'o', 'ô': 'o', 'õ': 'o', 'ø': 'o',
		'ú': 'u', 'ù': 'u', 'ü': 'u', 'û': 'u',
		'ý': 'y', 'ÿ': 'y',
		'ñ': 'n', 'ç': 'c',
		'Á': 'A', 'À': 'A', 'Ä': 'A', 'Â': 'A', 'Ã': 'A', 'Å': 'A',
		'É': 'E', 'È': 'E', 'Ë': 'E', 'Ê': 'E',
		'Í': 'I', 'Ì': 'I', 'Ï': 'I', 'Î': 'I',
		'Ó': 'O', 'Ò': 'O', 'Ö': 'O', 'Ô': 'O', 'Õ': 'O', 'Ø': 'O',
		'Ú': 'U', 'Ù': 'U', 'Ü': 'U', 'Û': 'U',
		'Ý': 'Y',
		'Ñ': 'N', 'Ç': 'C',
	}

	return strings.Map(func(r rune) rune {
		if replacement, ok := replacements[r]; ok {
			return replacement
		}
		return r
	}, input)
}

// reverseText reverses the entire text
func reverseText(input, arg1, arg2 string) string {
	runes := []rune(input)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// reverseWords reverses the characters in each word
func reverseWords(input, arg1, arg2 string) string {
	words := strings.Fields(input)
	result := make([]string, len(words))

	for i, word := range words {
		runes := []rune(word)
		for j, k := 0, len(runes)-1; j < k; j, k = j+1, k-1 {
			runes[j], runes[k] = runes[k], runes[j]
		}
		result[i] = string(runes)
	}

	return strings.Join(result, " ")
}

// reverseLines reverses characters in each line
func reverseLines(input, arg1, arg2 string) string {
	lines := strings.Split(input, "\n")
	result := make([]string, len(lines))

	for i, line := range lines {
		runes := []rune(line)
		for j, k := 0, len(runes)-1; j < k; j, k = j+1, k-1 {
			runes[j], runes[k] = runes[k], runes[j]
		}
		result[i] = string(runes)
	}

	return strings.Join(result, "\n")
}

// slugify creates a URL-safe slug from text
func slugify(input, arg1, arg2 string) string {
	// Convert to lowercase
	slug := strings.ToLower(input)

	// Remove diacritics
	slug = stripDiacritics(slug, "", "")

	// Replace spaces and underscores with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// Remove non-alphanumeric characters except hyphens
	slug = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(slug, "")

	// Remove consecutive hyphens
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	return slug
}

// smartQuotes converts straight quotes to curly/smart quotes
func smartQuotes(input, arg1, arg2 string) string {
	result := input

	// Replace double quotes with smart quotes (alternating left and right)
	// This is a simple implementation that replaces pairs of straight quotes
	quoteRunes := []rune(result)
	inQuote := false
	processedRunes := make([]rune, 0, len(quoteRunes))

	for _, r := range quoteRunes {
		if r == '"' {
			if inQuote {
				processedRunes = append(processedRunes, '\u201d') // right double quotation mark
			} else {
				processedRunes = append(processedRunes, '\u201c') // left double quotation mark
			}
			inQuote = !inQuote
		} else if r == '\'' {
			if inQuote {
				processedRunes = append(processedRunes, '\u2019') // right single quotation mark
			} else {
				processedRunes = append(processedRunes, '\u2018') // left single quotation mark
			}
		} else {
			processedRunes = append(processedRunes, r)
		}
	}

	return string(processedRunes)
}

// straightQuotes converts curly/smart quotes to straight quotes
func straightQuotes(input, arg1, arg2 string) string {
	result := input

	// Replace curly double quotes with straight quotes
	result = strings.ReplaceAll(result, "\u201c", `"`) // Replace left double quotation mark with "
	result = strings.ReplaceAll(result, "\u201d", `"`) // Replace right double quotation mark with "

	// Replace curly single quotes with straight quotes
	result = strings.ReplaceAll(result, "\u2018", "'") // Replace left single quotation mark with '
	result = strings.ReplaceAll(result, "\u2019", "'") // Replace right single quotation mark with '

	return result
}

// Phase 4: Encoding & Dates

// base64Encode encodes text as base64
func base64Encode(input, arg1, arg2 string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// base64Decode decodes base64-encoded text
func base64Decode(input, arg1, arg2 string) string {
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return input
	}
	return string(decoded)
}

// urlEncode percent-encodes text for URLs
func urlEncode(input, arg1, arg2 string) string {
	return url.QueryEscape(input)
}

// urlDecode percent-decodes URL-encoded text
func urlDecode(input, arg1, arg2 string) string {
	decoded, err := url.QueryUnescape(input)
	if err != nil {
		return input
	}
	return decoded
}

// hexEncode converts text to hexadecimal
func hexEncode(input, arg1, arg2 string) string {
	return hex.EncodeToString([]byte(input))
}

// hexDecode converts hexadecimal to text
func hexDecode(input, arg1, arg2 string) string {
	decoded, err := hex.DecodeString(input)
	if err != nil {
		return input
	}
	return string(decoded)
}

// rot13 applies ROT13 cipher
func rot13(input, arg1, arg2 string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case 'a' <= r && r <= 'z':
			return 'a' + (r-'a'+13)%26
		case 'A' <= r && r <= 'Z':
			return 'A' + (r-'A'+13)%26
		default:
			return r
		}
	}, input)
}

// escapeQuotes escapes quote characters for use in strings
func escapeQuotes(input, arg1, arg2 string) string {
	result := strings.ReplaceAll(input, `"`, `\"`)
	result = strings.ReplaceAll(result, "'", `\'`)
	return result
}

// unescapeQuotes unescapes escaped quote characters
func unescapeQuotes(input, arg1, arg2 string) string {
	result := strings.ReplaceAll(input, `\"`, `"`)
	result = strings.ReplaceAll(result, `\'`, "'")
	return result
}

// insertDateTime inserts current date/time
// arg1: format string (e.g., "2006-01-02" for date, default: RFC3339)
func insertDateTime(input, arg1, arg2 string) string {
	format := time.RFC3339
	if arg1 != "" {
		format = arg1
	}

	timestamp := time.Now().Format(format)

	// Replace ${DATE} or ${TIME} if present, otherwise just return timestamp
	if strings.Contains(input, "${DATE}") || strings.Contains(input, "${TIME}") {
		result := strings.ReplaceAll(input, "${DATE}", timestamp)
		result = strings.ReplaceAll(result, "${TIME}", timestamp)
		return result
	}

	return timestamp
}

// Phase 5: Markdown/HTML

// urlsToHyperlinks converts plain URLs to HTML hyperlinks
func urlsToHyperlinks(input, arg1, arg2 string) string {
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)

	format := `<a href="$0">$0</a>`
	if arg1 != "" {
		format = arg1
	}

	result := urlRegex.ReplaceAllStringFunc(input, func(match string) string {
		if strings.Contains(format, "$0") {
			return strings.ReplaceAll(format, "$0", match)
		}
		return match
	})

	return result
}

// extractUrls finds all URLs in text
func extractUrls(input, arg1, arg2 string) string {
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)
	matches := urlRegex.FindAllString(input, -1)

	if len(matches) == 0 {
		return ""
	}

	return strings.Join(matches, "\n")
}

// extractEmails finds all email addresses
func extractEmails(input, arg1, arg2 string) string {
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	matches := emailRegex.FindAllString(input, -1)

	if len(matches) == 0 {
		return ""
	}

	return strings.Join(matches, "\n")
}

// extractNumbers finds all numbers in text
func extractNumbers(input, arg1, arg2 string) string {
	numberRegex := regexp.MustCompile(`-?\d+(?:\.\d+)?`)
	matches := numberRegex.FindAllString(input, -1)

	if len(matches) == 0 {
		return ""
	}

	return strings.Join(matches, "\n")
}

// Phase 6: Advanced Regex

// extractWithGroups extracts regex matches with capture groups
// arg1: regex pattern
// arg2: output template (e.g., "$1 - $2")
func extractWithGroups(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}

	re, err := regexp.Compile(arg1)
	if err != nil {
		return input
	}

	template := "$0"
	if arg2 != "" {
		template = arg2
	}

	matches := re.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 {
		return ""
	}

	var result []string
	for _, match := range matches {
		line := template
		for i, group := range match {
			line = strings.ReplaceAll(line, fmt.Sprintf("$%d", i), group)
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// replaceWithGroups replaces text using regex with capture group references
// arg1: regex pattern
// arg2: replacement template (e.g., "$1-$2")
func replaceWithGroups(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}

	re, err := regexp.Compile(arg1)
	if err != nil {
		return input
	}

	replacement := arg2
	if replacement == "" {
		replacement = "$0"
	}

	return re.ReplaceAllString(input, replacement)
}

// splitByRegex splits text by regex pattern
// arg1: regex pattern
// arg2: delimiter to rejoin (if empty, just split)
func splitByRegex(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}

	re, err := regexp.Compile(arg1)
	if err != nil {
		return input
	}

	parts := re.Split(input, -1)

	delimiter := "\n"
	if arg2 != "" {
		delimiter = arg2
	}

	return strings.Join(parts, delimiter)
}

// matchCount counts the number of regex matches
// arg1: regex pattern
func matchCount(input, arg1, arg2 string) string {
	if arg1 == "" {
		return "0"
	}

	re, err := regexp.Compile(arg1)
	if err != nil {
		return "0"
	}

	matches := re.FindAllString(input, -1)
	return fmt.Sprintf("%d", len(matches))
}

// Phase 7: Math & Numbers

// formatNumbersOperation formats all numbers in text
// arg1: decimal places
// arg2: thousands separator (comma by default)
func formatNumbersOperation(input, arg1, arg2 string) string {
	decimals := 2
	if arg1 != "" {
		if d, err := strconv.Atoi(arg1); err == nil && d >= 0 {
			decimals = d
		}
	}

	separator := ","
	if arg2 != "" {
		separator = arg2
	}

	numberRegex := regexp.MustCompile(`-?\d+(?:\.\d+)?`)

	result := numberRegex.ReplaceAllStringFunc(input, func(match string) string {
		num, err := strconv.ParseFloat(match, 64)
		if err != nil {
			return match
		}

		// Format with specified decimal places
		formatted := fmt.Sprintf("%.*f", decimals, num)

		// Add thousands separator
		parts := strings.Split(formatted, ".")
		intPart := parts[0]

		// Handle negative sign
		negative := false
		if strings.HasPrefix(intPart, "-") {
			negative = true
			intPart = intPart[1:]
		}

		// Add separators
		runes := []rune(intPart)
		if len(runes) > 3 {
			var withSeparators []rune
			for i, r := range runes {
				if i > 0 && (len(runes)-i)%3 == 0 {
					withSeparators = append(withSeparators, []rune(separator)...)
				}
				withSeparators = append(withSeparators, r)
			}
			intPart = string(withSeparators)
		}

		if negative {
			intPart = "-" + intPart
		}

		if len(parts) > 1 {
			return intPart + "." + parts[1]
		}
		return intPart
	})

	return result
}

// roundNumbers rounds all numbers in text to specified decimal places
// arg1: decimal places
func roundNumbers(input, arg1, arg2 string) string {
	decimals := 0
	if arg1 != "" {
		if d, err := strconv.Atoi(arg1); err == nil && d >= 0 {
			decimals = d
		}
	}

	numberRegex := regexp.MustCompile(`-?\d+(?:\.\d+)?`)

	result := numberRegex.ReplaceAllStringFunc(input, func(match string) string {
		num, err := strconv.ParseFloat(match, 64)
		if err != nil {
			return match
		}

		rounded := math.Round(num*math.Pow10(decimals)) / math.Pow10(decimals)
		return formatNumber(rounded)
	})

	return result
}

// sumNumbers extracts all numbers and returns their sum
func sumNumbers(input, arg1, arg2 string) string {
	numberRegex := regexp.MustCompile(`-?\d+(?:\.\d+)?`)
	matches := numberRegex.FindAllString(input, -1)

	sum := 0.0
	for _, match := range matches {
		if num, err := strconv.ParseFloat(match, 64); err == nil {
			sum += num
		}
	}

	return formatNumber(sum)
}

// Phase 8: List & Extraction

// joinList joins lines with a delimiter
// arg1: delimiter
func joinList(input, arg1, arg2 string) string {
	if input == "" {
		return input
	}

	lines := strings.Split(input, "\n")
	delimiter := ", "
	if arg1 != "" {
		delimiter = arg1
	}

	return strings.Join(lines, delimiter)
}

// removeControlCharacters removes non-printable control characters
func removeControlCharacters(input, arg1, arg2 string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		return r
	}, input)
}

// countOccurrences counts how many times a string appears
// arg1: search string
func countOccurrences(input, arg1, arg2 string) string {
	if arg1 == "" {
		return "0"
	}

	count := strings.Count(input, arg1)
	return fmt.Sprintf("%d", count)
}

// keepLinesContaining keeps only lines containing the search string
// arg1: search string
// arg2: case-insensitive flag ("i")
func keepLinesContaining(input, arg1, arg2 string) string {
	if input == "" || arg1 == "" {
		return input
	}

	lines := strings.Split(input, "\n")
	var result []string

	caseInsensitive := strings.Contains(arg2, "i")

	for _, line := range lines {
		lineToCheck := line
		searchStr := arg1

		if caseInsensitive {
			lineToCheck = strings.ToLower(line)
			searchStr = strings.ToLower(arg1)
		}

		if strings.Contains(lineToCheck, searchStr) {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// removeLinesContaining removes lines containing the search string
// arg1: search string
// arg2: case-insensitive flag ("i")
func removeLinesContaining(input, arg1, arg2 string) string {
	if input == "" || arg1 == "" {
		return input
	}

	lines := strings.Split(input, "\n")
	var result []string

	caseInsensitive := strings.Contains(arg2, "i")

	for _, line := range lines {
		lineToCheck := line
		searchStr := arg1

		if caseInsensitive {
			lineToCheck = strings.ToLower(line)
			searchStr = strings.ToLower(arg1)
		}

		if !strings.Contains(lineToCheck, searchStr) {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// truncateText truncates text to maximum length
// arg1: maximum length
// arg2: ellipsis string (default "...")
func truncateText(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}

	maxLen, err := strconv.Atoi(arg1)
	if err != nil || maxLen < 0 {
		return input
	}

	ellipsis := "..."
	if arg2 != "" {
		ellipsis = arg2
	}

	runes := []rune(input)
	if len(runes) <= maxLen {
		return input
	}

	return string(runes[:maxLen]) + ellipsis
}

// Phase 9: Conditional Operations

// isEmpty returns "true" if input is empty or whitespace, "false" otherwise
func isEmpty(input, arg1, arg2 string) string {
	if strings.TrimSpace(input) == "" {
		return "true"
	}
	return "false"
}

// hasPattern returns "true" if input matches pattern, "false" otherwise
// arg1: regex pattern
func hasPattern(input, arg1, arg2 string) string {
	if arg1 == "" {
		return "false"
	}

	re, err := regexp.Compile(arg1)
	if err != nil {
		return "false"
	}

	if re.MatchString(input) {
		return "true"
	}
	return "false"
}

// startsWith returns "true" if input starts with arg1
func startsWith(input, arg1, arg2 string) string {
	if strings.HasPrefix(input, arg1) {
		return "true"
	}
	return "false"
}

// Phase 10: List Processing

// uniqueValues keeps only unique values (removes duplicates)
// arg1: delimiter (default: newline)
func uniqueValues(input, arg1, arg2 string) string {
	delimiter := "\n"
	if arg1 != "" {
		delimiter = arg1
	}

	items := strings.Split(input, delimiter)
	seen := make(map[string]bool)
	var result []string

	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return strings.Join(result, delimiter)
}

// mostCommon returns the most frequently occurring item
// arg1: delimiter
func mostCommon(input, arg1, arg2 string) string {
	delimiter := "\n"
	if arg1 != "" {
		delimiter = arg1
	}

	items := strings.Split(input, delimiter)
	counts := make(map[string]int)
	maxCount := 0
	var mostCommonItem string

	for _, item := range items {
		counts[item]++
		if counts[item] > maxCount {
			maxCount = counts[item]
			mostCommonItem = item
		}
	}

	return mostCommonItem
}

// leastCommon returns the least frequently occurring item
// arg1: delimiter
func leastCommon(input, arg1, arg2 string) string {
	delimiter := "\n"
	if arg1 != "" {
		delimiter = arg1
	}

	items := strings.Split(input, delimiter)
	counts := make(map[string]int)
	minCount := len(items)
	var leastCommonItem string

	for _, item := range items {
		counts[item]++
	}

	for item, count := range counts {
		if count < minCount {
			minCount = count
			leastCommonItem = item
		}
	}

	return leastCommonItem
}

// reverseLinesOrder reverses the order of lines
func reverseLinesOrder(input, arg1, arg2 string) string {
	return invertLines(input, arg1, arg2)
}

// groupByPattern groups lines by pattern match
// arg1: regex pattern to match
func groupByPattern(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}

	re, err := regexp.Compile(arg1)
	if err != nil {
		return input
	}

	lines := strings.Split(input, "\n")
	var result strings.Builder

	for _, line := range lines {
		if re.MatchString(line) {
			result.WriteString("MATCH: ")
		} else {
			result.WriteString("NO MATCH: ")
		}
		result.WriteString(line)
		result.WriteString("\n")
	}

	return strings.TrimSuffix(result.String(), "\n")
}

// Phase 11: Advanced Text Operations

// wordCount returns word/char/line statistics
func wordCount(input, arg1, arg2 string) string {
	words := strings.Fields(input)
	chars := len(input)
	lines := len(strings.Split(input, "\n"))

	return fmt.Sprintf("Words: %d\nCharacters: %d\nLines: %d", len(words), chars, lines)
}

// characterCount counts occurrences of a specific character
// arg1: character to count
func characterCount(input, arg1, arg2 string) string {
	if arg1 == "" {
		return "0"
	}

	count := strings.Count(input, arg1)
	return fmt.Sprintf("%d", count)
}

// lineCount returns the number of lines
func lineCount(input, arg1, arg2 string) string {
	if input == "" {
		return "0"
	}

	count := len(strings.Split(input, "\n"))
	return fmt.Sprintf("%d", count)
}

// textStatistics returns detailed text statistics
func textStatistics(input, arg1, arg2 string) string {
	if input == "" {
		return ""
	}

	words := strings.Fields(input)
	lines := strings.Split(input, "\n")

	totalChars := len(input)
	totalWords := len(words)
	totalLines := len(lines)

	var minLen, maxLen, totalLen int
	if len(words) > 0 {
		minLen = len(words[0])
		maxLen = len(words[0])

		for _, w := range words {
			l := len(w)
			totalLen += l
			if l < minLen {
				minLen = l
			}
			if l > maxLen {
				maxLen = l
			}
		}
	}

	avgLen := 0.0
	if totalWords > 0 {
		avgLen = float64(totalLen) / float64(totalWords)
	}

	return fmt.Sprintf("Lines: %d\nWords: %d\nCharacters: %d\nMin Word: %d\nMax Word: %d\nAvg Word: %.2f",
		totalLines, totalWords, totalChars, minLen, maxLen, avgLen)
}

// minWordLength returns the minimum word length
func minWordLength(input, arg1, arg2 string) string {
	words := strings.Fields(input)
	if len(words) == 0 {
		return "0"
	}

	minLen := len(words[0])
	for _, w := range words {
		if len(w) < minLen {
			minLen = len(w)
		}
	}

	return fmt.Sprintf("%d", minLen)
}

// maxWordLength returns the maximum word length
func maxWordLength(input, arg1, arg2 string) string {
	words := strings.Fields(input)
	if len(words) == 0 {
		return "0"
	}

	maxLen := len(words[0])
	for _, w := range words {
		if len(w) > maxLen {
			maxLen = len(w)
		}
	}

	return fmt.Sprintf("%d", maxLen)
}

// averageWordLength returns the average word length
func averageWordLength(input, arg1, arg2 string) string {
	words := strings.Fields(input)
	if len(words) == 0 {
		return "0"
	}

	totalLen := 0
	for _, w := range words {
		totalLen += len(w)
	}

	avg := float64(totalLen) / float64(len(words))
	return fmt.Sprintf("%.2f", avg)
}

// Phase 12: Advanced Pattern Operations

// wholeWordMatch finds whole word matches only
// arg1: word to match
func wholeWordMatch(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}

	// Use word boundary regex
	pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(arg1))
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(input, -1)

	if len(matches) == 0 {
		return ""
	}

	return strings.Join(matches, "\n")
}

// caseSensitiveFind finds case-sensitive matches
// arg1: search string
func caseSensitiveFind(input, arg1, arg2 string) string {
	if arg1 == "" {
		return "0"
	}

	count := strings.Count(input, arg1)
	return fmt.Sprintf("%d", count)
}

// multilinePattern applies multiline regex matching
// arg1: regex pattern
func multilinePattern(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}

	pattern := "(?m)" + arg1
	re, err := regexp.Compile(pattern)
	if err != nil {
		return input
	}

	matches := re.FindAllString(input, -1)
	if len(matches) == 0 {
		return ""
	}

	return strings.Join(matches, "\n")
}

// lookaheadPattern matches text followed by pattern
// arg1: pattern to match, arg2: lookahead pattern
func lookaheadPattern(input, arg1, arg2 string) string {
	if arg1 == "" || arg2 == "" {
		return input
	}

	pattern := fmt.Sprintf(`%s(?=%s)`, arg1, arg2)
	re, err := regexp.Compile(pattern)
	if err != nil {
		return input
	}

	matches := re.FindAllString(input, -1)
	if len(matches) == 0 {
		return ""
	}

	return strings.Join(matches, "\n")
}

// lookbehindPattern matches text preceded by pattern
// Note: Go regexp doesn't support lookbehind, so we implement a simpler version
// arg1: pattern to match, arg2: lookbehind pattern (as literal string)
func lookbehindPattern(input, arg1, arg2 string) string {
	if arg1 == "" || arg2 == "" {
		return input
	}

	// Find all matches of pattern that are preceded by arg2
	re, err := regexp.Compile(arg1)
	if err != nil {
		return input
	}

	lines := strings.Split(input, "\n")
	var result []string

	for _, line := range lines {
		if strings.Contains(line, arg2) {
			idx := strings.Index(line, arg2)
			remaining := line[idx+len(arg2):]
			matches := re.FindAllString(remaining, -1)
			result = append(result, matches...)
		}
	}

	return strings.Join(result, "\n")
}

// conditionalReplace replaces based on conditions
// arg1: condition pattern, arg2: replacement
func conditionalReplace(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}

	re, err := regexp.Compile(arg1)
	if err != nil {
		return input
	}

	replacement := arg2
	if replacement == "" {
		replacement = "[MATCH]"
	}

	return re.ReplaceAllString(input, replacement)
}

// Phase 13: Transformation Macros

// chainOperations chains multiple operations (simplified version)
// arg1: operation1|operation2|operation3 format
func chainOperations(input, arg1, arg2 string) string {
	if arg1 == "" {
		return input
	}

	// Simple implementation: split by |, treat each as operation name
	ops := strings.Split(arg1, "|")
	result := input

	for _, op := range ops {
		opName := strings.TrimSpace(op)
		if opName != "" {
			result = ProcessText(result, opName, "", "")
		}
	}

	return result
}

// repeatOperation repeats an operation multiple times
// arg1: operation name, arg2: count
func repeatOperation(input, arg1, arg2 string) string {
	if arg1 == "" || arg2 == "" {
		return input
	}

	count, err := strconv.Atoi(arg2)
	if err != nil || count <= 0 {
		return input
	}

	result := input
	for i := 0; i < count; i++ {
		result = ProcessText(result, arg1, "", "")
	}

	return result
}

// swapPairs swaps pairs of items separated by delimiter
// arg1: delimiter
func swapPairs(input, arg1, arg2 string) string {
	delimiter := "|"
	if arg1 != "" {
		delimiter = arg1
	}

	lines := strings.Split(input, "\n")
	var result []string

	for _, line := range lines {
		parts := strings.Split(line, delimiter)
		if len(parts) == 2 {
			result = append(result, parts[1]+delimiter+parts[0])
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// reverseOrderItems reverses the order of items
// arg1: delimiter
func reverseOrderItems(input, arg1, arg2 string) string {
	delimiter := "\n"
	if arg1 != "" {
		delimiter = arg1
	}

	items := strings.Split(input, delimiter)
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}

	return strings.Join(items, delimiter)
}

// Phase 14: HTML/Markdown Advanced

// htmlToMarkdown converts HTML to Markdown (simplified)
func htmlToMarkdown(input, arg1, arg2 string) string {
	result := input

	// Simple conversions
	result = regexp.MustCompile(`<h[1-6][^>]*>([^<]+)</h[1-6]>`).ReplaceAllString(result, "# $1")
	result = regexp.MustCompile(`<p[^>]*>([^<]+)</p>`).ReplaceAllString(result, "$1\n\n")
	result = regexp.MustCompile(`<strong[^>]*>([^<]+)</strong>`).ReplaceAllString(result, "**$1**")
	result = regexp.MustCompile(`<b[^>]*>([^<]+)</b>`).ReplaceAllString(result, "**$1**")
	result = regexp.MustCompile(`<em[^>]*>([^<]+)</em>`).ReplaceAllString(result, "*$1*")
	result = regexp.MustCompile(`<i[^>]*>([^<]+)</i>`).ReplaceAllString(result, "*$1*")
	result = regexp.MustCompile(`<a[^>]*href="([^"]*)"[^>]*>([^<]+)</a>`).ReplaceAllString(result, "[$2]($1)")
	result = regexp.MustCompile(`<li[^>]*>([^<]+)</li>`).ReplaceAllString(result, "- $1")
	result = regexp.MustCompile(`<br[^>]*/?>`).ReplaceAllString(result, "  \n")

	return result
}

// markdownToHTML converts Markdown to HTML (simplified)
func markdownToHTML(input, arg1, arg2 string) string {
	result := input

	// Simple conversions
	result = regexp.MustCompile(`^# (.+)$`).ReplaceAllString(result, "<h1>$1</h1>")
	result = regexp.MustCompile(`^## (.+)$`).ReplaceAllString(result, "<h2>$1</h2>")
	result = regexp.MustCompile(`\*\*(.+?)\*\*`).ReplaceAllString(result, "<strong>$1</strong>")
	result = regexp.MustCompile(`\*(.+?)\*`).ReplaceAllString(result, "<em>$1</em>")
	result = regexp.MustCompile(`\[(.+?)\]\((.+?)\)`).ReplaceAllString(result, "<a href=\"$2\">$1</a>")
	result = regexp.MustCompile(`^- (.+)$`).ReplaceAllString(result, "<li>$1</li>")

	return result
}

// extractTextFromHTML extracts all text content from HTML
func extractTextFromHTML(input, arg1, arg2 string) string {
	return stripTags(input, arg1, arg2)
}

// createMarkdownTable creates a Markdown table from delimited data
// arg1: delimiter for columns
// arg2: rows delimiter
func createMarkdownTable(input, arg1, arg2 string) string {
	colDelim := "|"
	if arg1 != "" {
		colDelim = arg1
	}

	rowDelim := "\n"
	if arg2 != "" {
		rowDelim = arg2
	}

	rows := strings.Split(input, rowDelim)
	if len(rows) == 0 {
		return input
	}

	var result strings.Builder
	result.WriteString("|")

	// Header row
	firstRow := strings.Split(rows[0], colDelim)
	for _, cell := range firstRow {
		result.WriteString(" ")
		result.WriteString(strings.TrimSpace(cell))
		result.WriteString(" |")
	}

	result.WriteString("\n|")

	// Separator
	for range firstRow {
		result.WriteString(" --- |")
	}

	result.WriteString("\n")

	// Data rows
	for i := 1; i < len(rows); i++ {
		cells := strings.Split(rows[i], colDelim)
		result.WriteString("|")
		for _, cell := range cells {
			result.WriteString(" ")
			result.WriteString(strings.TrimSpace(cell))
			result.WriteString(" |")
		}
		result.WriteString("\n")
	}

	return result.String()
}

// parseYAMLFrontMatter extracts YAML front matter
func parseYAMLFrontMatter(input, arg1, arg2 string) string {
	// Look for --- delimiters
	if !strings.HasPrefix(input, "---") {
		return ""
	}

	// Find the closing ---
	remaining := strings.TrimPrefix(input, "---\n")
	idx := strings.Index(remaining, "---")

	if idx == -1 {
		return ""
	}

	return remaining[:idx]
}

// markdownLinkFormat converts markdown links to custom format
// arg1: output format (default "[text](url)")
func markdownLinkFormat(input, arg1, arg2 string) string {
	re := regexp.MustCompile(`\[(.+?)\]\((.+?)\)`)

	format := "[text](url)"
	if arg1 != "" {
		format = arg1
	}

	result := re.ReplaceAllStringFunc(input, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 3 {
			text := parts[1]
			url := parts[2]

			output := strings.ReplaceAll(format, "text", text)
			output = strings.ReplaceAll(output, "url", url)
			return output
		}
		return match
	})

	return result
}

// Phase 15: Unicode & Special Characters

// unicodeNames converts characters to their Unicode names
func unicodeNames(input, arg1, arg2 string) string {
	var result strings.Builder

	for _, r := range input {
		if r < 128 {
			result.WriteRune(r)
		} else {
			result.WriteString(fmt.Sprintf("U+%04X ", r))
		}
	}

	return result.String()
}

// convertUnicodeEscapes converts \uXXXX escapes to characters
func convertUnicodeEscapes(input, arg1, arg2 string) string {
	return processEscapeSequences(input)
}

// escapeUnicode converts characters to \uXXXX format
func escapeUnicode(input, arg1, arg2 string) string {
	var result strings.Builder

	for _, r := range input {
		if r < 128 && r >= 32 {
			result.WriteRune(r)
		} else {
			result.WriteString(fmt.Sprintf("\\u%04X", r))
		}
	}

	return result.String()
}

// showInvisibleCharacters displays invisible characters visibly
func showInvisibleCharacters(input, arg1, arg2 string) string {
	result := input

	result = strings.ReplaceAll(result, "\n", "↓\n")
	result = strings.ReplaceAll(result, "\t", "→")
	result = strings.ReplaceAll(result, " ", "·")
	result = strings.ReplaceAll(result, "\r", "↵")

	return result
}

// normalizeUnicode applies Unicode normalization (simplified - no decomposition)
// arg1: normalization form (NFC, NFD, NFKC, NFKD - not fully implemented)
func normalizeUnicode(input, arg1, arg2 string) string {
	// Go doesn't have built-in Unicode normalization in standard lib
	// This is a simplified version that removes common combining marks
	return stripDiacritics(input, arg1, arg2)
}

// normalizeWhitespace collapses multiple whitespace characters to single spaces
func normalizeWhitespace(input, arg1, arg2 string) string {
	// Replace multiple spaces/tabs/etc with single space
	result := regexp.MustCompile(`\s+`).ReplaceAllString(input, " ")
	return strings.TrimSpace(result)
}

// Helper functions

// extractLeadingNumber extracts the leading number from a string
func extractLeadingNumber(s string) *float64 {
	re := regexp.MustCompile(`^-?\d+(?:\.\d+)?`)
	match := re.FindString(s)
	if match == "" {
		return nil
	}

	num, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return nil
	}

	return &num
}

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
