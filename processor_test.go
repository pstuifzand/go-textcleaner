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

func TestPHPDeserialize(t *testing.T) {
	tests := []struct {
		input    string
		arg1     string // outputValidJSON: "true" or "false"
		expected string
		desc     string
	}{
		// Basic types
		{`n;`, "false", `null`, "Null value"},
		{`i:42;`, "false", `42`, "Integer"},
		{`d:3.14;`, "false", `3.14`, "Double"},
		{`b:1;`, "false", `true`, "Boolean true"},
		{`b:0;`, "false", `false`, "Boolean false"},
		{`s:5:"hello";`, "false", `"hello"`, "String"},

		// Arrays with string keys
		{`a:1:{s:1:"a";i:10;}`, "false", `{"a": 10}`, "Array with string key"},
		{`a:2:{s:1:"a";i:10;s:1:"b";i:20;}`, "false", `{"a": 10,"b": 20}`, "Array with multiple string keys"},

		// Sequential arrays (numeric keys 0,1,2,...)
		{`a:1:{i:0;s:5:"hello";}`, "false", `["hello"]`, "Sequential array with single element"},
		{`a:3:{i:0;s:5:"hello";i:1;s:5:"world";i:2;i:42;}`, "false", `["hello","world",42]`, "Sequential array with multiple types"},
		{`a:2:{s:1:"a";i:10;i:0;a:1:{i:0;s:2:"ab";}}`, "false", `{"a": 10,"0": ["ab"]}`, "Mixed associative and sequential"},

		// Non-sequential arrays (should output as objects)
		{`a:2:{i:0;s:1:"a";i:2;s:1:"b";}`, "false", `{"0": "a","2": "b"}`, "Non-sequential array (gaps in indices)"},
		{`a:2:{i:1;s:1:"a";i:0;s:1:"b";}`, "false", `{"1": "a","0": "b"}`, "Non-sequential array (wrong order)"},

		// Objects (PHP class instances)
		{`O:8:"stdClass":1:{s:4:"data";s:10:"Some data!";}`, "false", `{"__class__": "stdClass", "properties": {"data": "Some data!"}}`, "stdClass object"},
		{`O:7:"MyClass":2:{s:5:"field";i:42;s:3:"foo";s:3:"bar";}`, "false", `{"__class__": "MyClass", "properties": {"field": 42,"foo": "bar"}}`, "Custom class object"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			result := phpDeserialize(test.input, test.arg1, "")
			if result != test.expected {
				t.Errorf("Input: %q, Expected: %q, Got: %q", test.input, test.expected, result)
			}
		})
	}
}

func TestJSONFormat(t *testing.T) {
	tests := []struct {
		input    string
		arg1     string // indent size (default 2)
		arg2     string // prefix (default "")
		expected string
		desc     string
	}{
		// Basic formatting with default indent (2 spaces)
		{`{"name":"John","age":30}`, "", "", `{
  "age": 30,
  "name": "John"
}`, "Compact JSON to formatted with default indent"},

		// Custom indent size
		{`{"a":1,"b":2}`, "4", "", `{
    "a": 1,
    "b": 2
}`, "Format with 4-space indent"},

		// No indentation (compact)
		{`{"x":1,"y":2}`, "0", "", `{"x":1,"y":2}`, "Compact format with 0 indent"},

		// Nested objects
		{`{"user":{"name":"John","age":30},"active":true}`, "2", "", `{
  "active": true,
  "user": {
    "age": 30,
    "name": "John"
  }
}`, "Nested object formatting"},

		// Array formatting
		{`[1,2,3]`, "2", "", `[
  1,
  2,
  3
]`, "Array formatting"},

		// With prefix (quoting style)
		{`{"a":1}`, "2", "> ", `> {
>   "a": 1
> }`, "Format with quoting prefix"},

		// With custom prefix
		{`{"x":10}`, "2", "| ", `| {
|   "x": 10
| }`, "Format with custom prefix"},

		// Already formatted JSON (should reformat)
		{`{
  "test": "value"
}`, "4", "", `{
    "test": "value"
}`, "Reformat with different indent"},

		// Empty object
		{`{}`, "", "", `{}`, "Empty object"},

		// Empty array
		{`[]`, "", "", `[]`, "Empty array"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			result := jsonFormat(test.input, test.arg1, test.arg2)
			if result != test.expected {
				t.Errorf("Expected:\n%q\n\nGot:\n%q", test.expected, result)
			}
		})
	}
}

func TestRemoveComments(t *testing.T) {
	tests := []struct {
		input    string
		arg1     string // comment type filter ("", "#", "//", "/*", or "all")
		expected string
		desc     string
	}{
		// Test removing # comments (Python/Shell style)
		{
			input:    "name = 'John' # This is a comment",
			arg1:     "#",
			expected: "name = 'John'",
			desc:     "Remove Python comment",
		},
		{
			input:    "# This is a full line comment\nprint('hello')",
			arg1:     "#",
			expected: "\nprint('hello')",
			desc:     "Remove full-line Python comment",
		},
		{
			input:    "var x = 5; # inline comment\nvar y = 10",
			arg1:     "#",
			expected: "var x = 5;\nvar y = 10",
			desc:     "Remove inline Python comment",
		},

		// Test removing // comments (C/C++/JavaScript style)
		{
			input:    "int x = 5; // This is a comment",
			arg1:     "//",
			expected: "int x = 5;",
			desc:     "Remove C++ comment",
		},
		{
			input:    "// This is a full line comment\nint main() {}",
			arg1:     "//",
			expected: "\nint main() {}",
			desc:     "Remove full-line C++ comment",
		},
		{
			input:    "const x = 5; // variable\nconst y = 10;",
			arg1:     "//",
			expected: "const x = 5;\nconst y = 10;",
			desc:     "Remove inline JavaScript comment",
		},

		// Test removing /* */ comments (block comments)
		{
			input:    "int x = 5; /* block comment */ int y = 10;",
			arg1:     "/*",
			expected: "int x = 5;  int y = 10;",
			desc:     "Remove inline block comment",
		},
		{
			input:    "/* Full block\ncomment\nacross lines */\nint z = 15;",
			arg1:     "/*",
			expected: "\nint z = 15;",
			desc:     "Remove multi-line block comment",
		},
		{
			input:    "code /* comment1 */ more /* comment2 */ end",
			arg1:     "/*",
			expected: "code  more  end",
			desc:     "Remove multiple block comments",
		},

		// Test removing all types (default)
		{
			input:    "x = 5 # python\ny = 10 // c++\nz = 15 /* block */",
			arg1:     "",
			expected: "x = 5\ny = 10\nz = 15 ",
			desc:     "Remove all comment types",
		},
		{
			input:    "a = 1 # comment1\nb = 2 // comment2\nc = 3 /* block */",
			arg1:     "all",
			expected: "a = 1\nb = 2\nc = 3 ",
			desc:     "Remove all comment types with 'all' filter",
		},

		// Test empty input
		{
			input:    "",
			arg1:     "",
			expected: "",
			desc:     "Empty input",
		},

		// Test no comments
		{
			input:    "just plain text\nwith no comments",
			arg1:     "",
			expected: "just plain text\nwith no comments",
			desc:     "Text with no comments",
		},

		// Test string literals preservation (simple cases)
		{
			input:    `msg = "hello # world"`,
			arg1:     "#",
			expected: `msg = "hello # world"`,
			desc:     "Preserve comment marker in double-quoted string",
		},
		{
			input:    `msg = 'hello # world'`,
			arg1:     "#",
			expected: `msg = 'hello # world'`,
			desc:     "Preserve comment marker in single-quoted string",
		},

		// Test escaped quotes
		{
			input:    `x = 5 // comment`,
			arg1:     "//",
			expected: `x = 5`,
			desc:     "Remove comment after code",
		},

		// Test multiple lines
		{
			input:    "line1 # comment1\nline2 // comment2\nline3",
			arg1:     "",
			expected: "line1\nline2\nline3",
			desc:     "Multiple lines with mixed comments",
		},

		// Test block comment edge cases
		{
			input:    "start /* comment without end",
			arg1:     "/*",
			expected: "start ",
			desc:     "Unclosed block comment",
		},
		{
			input:    "/* first */ middle /* second */ text",
			arg1:     "/*",
			expected: " middle  text",
			desc:     "Multiple separate block comments",
		},

		// Test trailing whitespace handling
		{
			input:    "code   // comment",
			arg1:     "//",
			expected: "code",
			desc:     "Remove trailing whitespace after comment",
		},
		{
			input:    "code    # comment",
			arg1:     "#",
			expected: "code",
			desc:     "Remove trailing spaces before comment",
		},

		// Test alias arguments for comment types
		{
			input:    "# comment here",
			arg1:     "python",
			expected: "",
			desc:     "Use 'python' alias for # comments",
		},
		{
			input:    "// comment here",
			arg1:     "c",
			expected: "",
			desc:     "Use 'c' alias for // comments",
		},
		{
			input:    "/* comment */ text",
			arg1:     "block",
			expected: " text",
			desc:     "Use 'block' alias for /* */ comments",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			result := removeComments(test.input, test.arg1, "")
			if result != test.expected {
				t.Errorf("Input: %q", test.input)
				t.Errorf("Arg1: %q", test.arg1)
				t.Errorf("Expected: %q", test.expected)
				t.Errorf("Got: %q", result)
			}
		})
	}
}


