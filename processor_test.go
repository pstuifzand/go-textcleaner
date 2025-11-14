package main

import (
	"fmt"
	"testing"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		// Simple operations
		{"6 + 5", "11", "Simple addition"},
		{"10 - 3", "7", "Simple subtraction"},
		{"4 * 5", "20", "Simple multiplication"},
		{"20 / 4", "5", "Simple division"},

		// Operator precedence
		{"2 + 3 * 4", "14", "Multiplication before addition"},
		{"10 - 2 * 3", "4", "Multiplication before subtraction"},
		{"20 / 4 + 3", "8", "Division before addition"},
		{"3 + 4 * 2 - 1", "10", "Complex expression"},

		// Floats
		{"5.5 + 2.5", "8", "Float addition"},
		{"10.5 - 3.2", "7.3", "Float subtraction"},
		{"2.5 * 4", "10", "Float multiplication"},
		{"7.5 / 2.5", "3", "Float division"},

		// Negative numbers
		{"-5 + 10", "5", "Negative number addition"},
		{"-10 * 2", "-20", "Negative number multiplication"},
		{"5 - 10", "-5", "Subtraction resulting in negative"},
		{"2 - 3", "-1", "Subtraction with spaces"},
		{"2-3", "-1", "Subtraction without spaces"},

		// Text with calculations
		{"Item 1 = 6 + 5", "Item 1 = 11", "Calculation in text"},
		{"Price: 5 * 55 + 3", "Price: 278", "Embedded calculation"},
		{"Total = 100 / 4 - 5", "Total = 20", "Calculation with division"},

		// Multiple calculations in one text
		{"5 + 3 and 4 * 2", "8 and 8", "Multiple expressions"},

		// Whitespace handling
		{"5+3", "8", "No spaces"},
		{"5 + 3", "8", "With spaces"},

		// Just numbers
		{"42", "42", "Single number"},
		{"3.14", "3.14", "Single float"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			result := calculate(test.input, "", "")
			if result != test.expected {
				t.Errorf("Input: %q", test.input)
				t.Errorf("Expected: %q", test.expected)
				t.Errorf("Got: %q", result)
			}
		})
	}
}

func TestEvaluateExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		shouldErr bool
	}{
		{"5", 5, false},
		{"-5", -5, false},
		{"3.14", 3.14, false},
		{"5 + 3", 8, false},
		{"10 - 3", 7, false},
		{"2 - 3", -1, false},
		{"2-3", -1, false},
		{"4 * 5", 20, false},
		{"20 / 4", 5, false},
		{"2 + 3 * 4", 14, false},
		{"10 - 2 * 3", 4, false},
		{"5.5 + 2.5", 8, false},
		{"-10 * 2", -20, false},
		{"20 / 0", 0, true},
		{"", 0, true},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%q", test.input), func(t *testing.T) {
			result, err := evaluateExpression(test.input)
			if test.shouldErr && err == nil {
				t.Errorf("Expected error for %q", test.input)
			}
			if !test.shouldErr && err != nil {
				t.Errorf("Unexpected error for %q: %v", test.input, err)
			}
			if !test.shouldErr && result != test.expected {
				t.Errorf("Input: %q, Expected: %f, Got: %f", test.input, test.expected, result)
			}
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{5, "5"},
		{5.0, "5"},
		{5.5, "5.5"},
		{3.14159, "3.14159"},
		{-5, "-5"},
		{-5.5, "-5.5"},
		{0, "0"},
		{0.1, "0.1"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.input), func(t *testing.T) {
			result := formatNumber(test.input)
			if result != test.expected {
				t.Errorf("Input: %v, Expected: %q, Got: %q", test.input, test.expected, result)
			}
		})
	}
}
