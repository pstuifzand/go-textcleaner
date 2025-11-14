package main

import (
	"html"
	"strings"
	"unicode"
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
