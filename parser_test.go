package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

func TestIsValidVariableName(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		expected bool
	}{
		{"valid uppercase", "DATABASE_URL", true},
		{"valid lowercase", "foobar", true},
		{"valid with underscore", "FOO_BAR", true},
		{"valid starting with underscore", "_PRIVATE", true},
		{"invalid with dash", "NO-WORK", false},
		{"invalid with special chars", "ÜBER", false},
		{"invalid starting with digit", "2MUCH", false},
		{"empty string", "", false},
		{"only digits", "123", false},
		{"mixed valid", "test_123", true},
		{"single letter", "A", true},
		{"single underscore", "_", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidVariableName(tt.varName)
			if result != tt.expected {
				t.Errorf("isValidVariableName(%q) = %v, want %v", tt.varName, result, tt.expected)
			}
		})
	}
}

func TestParseSimpleVariables(t *testing.T) {
	parser := NewParser()
	parser.lines = []string{
		"SIMPLE=value",
		"DATABASE_URL=postgres://localhost/db",
		"PORT=3000",
	}

	variables, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expected := []Variable{
		{Name: "SIMPLE", Value: "value"},
		{Name: "DATABASE_URL", Value: "postgres://localhost/db"},
		{Name: "PORT", Value: "3000"},
	}

	if !reflect.DeepEqual(variables, expected) {
		t.Errorf("Parse() = %v, want %v", variables, expected)
	}
}

func TestParseWithExportPrefix(t *testing.T) {
	parser := NewParser()
	parser.lines = []string{
		"export DATABASE_URL=postgres://localhost/db",
		"export SECRET_KEY=mysecret",
		"NORMAL=value",
	}

	variables, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expected := []Variable{
		{Name: "DATABASE_URL", Value: "postgres://localhost/db"},
		{Name: "SECRET_KEY", Value: "mysecret"},
		{Name: "NORMAL", Value: "value"},
	}

	if !reflect.DeepEqual(variables, expected) {
		t.Errorf("Parse() = %v, want %v", variables, expected)
	}
}

func TestParseComments(t *testing.T) {
	parser := NewParser()
	parser.lines = []string{
		"# This is a comment",
		"DATABASE_URL=postgres://localhost/db",
		"# Another comment",
		"SECRET_KEY=mysecret # inline comment",
		"",
		"PORT=3000",
	}

	variables, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expected := []Variable{
		{Name: "DATABASE_URL", Value: "postgres://localhost/db"},
		{Name: "SECRET_KEY", Value: "mysecret"},
		{Name: "PORT", Value: "3000"},
	}

	if !reflect.DeepEqual(variables, expected) {
		t.Errorf("Parse() = %v, want %v", variables, expected)
	}
}

func TestParseQuotedValues(t *testing.T) {
	parser := NewParser()
	parser.lines = []string{
		`SINGLE_QUOTED='single value'`,
		`DOUBLE_QUOTED="double value"`,
		`WITH_SPACES="  spaced value  "`,
		`EMPTY_SINGLE=''`,
		`EMPTY_DOUBLE=""`,
	}

	variables, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expected := []Variable{
		{Name: "SINGLE_QUOTED", Value: "single value"},
		{Name: "DOUBLE_QUOTED", Value: "double value"},
		{Name: "WITH_SPACES", Value: "  spaced value  "},
		{Name: "EMPTY_SINGLE", Value: ""},
		{Name: "EMPTY_DOUBLE", Value: ""},
	}

	if !reflect.DeepEqual(variables, expected) {
		t.Errorf("Parse() = %v, want %v", variables, expected)
	}
}

func TestParseEscapeSequences(t *testing.T) {
	parser := NewParser()
	parser.lines = []string{
		`NEWLINE="line1\nline2"`,
		`TAB="col1\tcol2"`,
		`QUOTE="He said \"hello\""`,
		`BACKSLASH="path\\to\\file"`,
		`UNICODE="Unicode: \u0041\u0042\u0043"`,
		`CARRIAGE_RETURN="line1\rline2"`,
		`FORM_FEED="page1\fpage2"`,
		`BACKSPACE="text\bspace"`,
	}

	variables, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expected := []Variable{
		{Name: "NEWLINE", Value: "line1\nline2"},
		{Name: "TAB", Value: "col1\tcol2"},
		{Name: "QUOTE", Value: `He said "hello"`},
		{Name: "BACKSLASH", Value: `path\to\file`},
		{Name: "UNICODE", Value: "Unicode: ABC"},
		{Name: "CARRIAGE_RETURN", Value: "line1\rline2"},
		{Name: "FORM_FEED", Value: "page1\fpage2"},
		{Name: "BACKSPACE", Value: "text\bspace"},
	}

	if !reflect.DeepEqual(variables, expected) {
		t.Errorf("Parse() = %v, want %v", strconv.Quote(fmt.Sprintf("%v", variables)), strconv.Quote(fmt.Sprintf("%v", expected)))
	}
}

func TestParseSingleQuoteNoEscape(t *testing.T) {
	parser := NewParser()
	parser.lines = []string{
		`NO_ESCAPE='raw\ntext\twith\backslashes'`,
		`QUOTE_ESCAPE='can\'t escape'`,
		`BACKSLASH_ESCAPE='path\\to\\file'`,
	}

	variables, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expected := []Variable{
		{Name: "NO_ESCAPE", Value: `raw\ntext\twith\backslashes`},
		{Name: "QUOTE_ESCAPE", Value: `can't escape`},
		{Name: "BACKSLASH_ESCAPE", Value: `path\to\file`},
	}

	if !reflect.DeepEqual(variables, expected) {
		t.Errorf("Parse() = %v, want %v", variables, expected)
	}
}

func TestParseVariableInterpolation(t *testing.T) {
	// Set up environment variable for testing
	os.Setenv("TEST_ENV_VAR", "from_env")
	defer os.Unsetenv("TEST_ENV_VAR")

	parser := NewParser()
	parser.lines = []string{
		"USER=admin",
		"EMAIL=${USER}@example.org",
		`FULL_PATH="/home/${USER}/documents"`,
		"ENV_VAR=${TEST_ENV_VAR}",
		"UNDEFINED=${UNDEFINED_VAR}",
		`NO_INTERPOLATION='${USER} not interpolated'`,
	}

	variables, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expected := []Variable{
		{Name: "USER", Value: "admin"},
		{Name: "EMAIL", Value: "admin@example.org"},
		{Name: "FULL_PATH", Value: "/home/admin/documents"},
		{Name: "ENV_VAR", Value: "from_env"},
		{Name: "UNDEFINED", Value: ""},
		{Name: "NO_INTERPOLATION", Value: "${USER} not interpolated"},
	}

	if !reflect.DeepEqual(variables, expected) {
		t.Errorf("Parse() = %v, want %v", variables, expected)
	}
}

func TestParseMultilineValues(t *testing.T) {
	parser := NewParser()
	parser.lines = []string{
		`SINGLE_LINE_TRIPLE="""single line"""`,
		`MULTILINE_DOUBLE="""`,
		`line 1`,
		`line 2`,
		`"""`,
		`MULTILINE_SINGLE='''`,
		`raw line 1`,
		`raw line 2`,
		`'''`,
		`WITH_INTERPOLATION="""`,
		`Hello ${USER}`,
		`"""`,
	}

	// Add a variable for interpolation test
	parser.variables = append(parser.variables, Variable{Name: "USER", Value: "admin"})

	variables, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expected := []Variable{
		{Name: "USER", Value: "admin"},
		{Name: "SINGLE_LINE_TRIPLE", Value: "single line"},
		{Name: "MULTILINE_DOUBLE", Value: "line 1\nline 2\n"},
		{Name: "MULTILINE_SINGLE", Value: "raw line 1\nraw line 2\n"},
		{Name: "WITH_INTERPOLATION", Value: "Hello admin\n"},
	}

	if !reflect.DeepEqual(variables, expected) {
		t.Errorf("Parse() = %v, want %v", variables, expected)
	}
}

func TestParseWhitespace(t *testing.T) {
	parser := NewParser()
	parser.lines = []string{
		"  LEADING_SPACE=value",
		"TRAILING_SPACE=value  ",
		"  BOTH_SPACES  =  value  ",
		"TABS\t=\tvalue\t",
		"MIXED   \t  = \t  value  \t ",
	}

	variables, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expected := []Variable{
		{Name: "LEADING_SPACE", Value: "value"},
		{Name: "TRAILING_SPACE", Value: "value"},
		{Name: "BOTH_SPACES", Value: "value"},
		{Name: "TABS", Value: "value"},
		{Name: "MIXED", Value: "value"},
	}

	if !reflect.DeepEqual(variables, expected) {
		t.Errorf("Parse() = %v, want %v", variables, expected)
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
	}{
		{
			name:  "invalid variable name",
			lines: []string{"123INVALID=value"},
		},
		{
			name:  "unterminated quote",
			lines: []string{`UNTERMINATED="missing quote`},
		},
		{
			name:  "unterminated multiline",
			lines: []string{`UNTERMINATED="""`, `some content`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()
			parser.lines = tt.lines

			_, err := parser.Parse()
			if err == nil {
				t.Errorf("Parse() should have returned an error for %s", tt.name)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.env")

	content := `# Test file
DATABASE_URL=postgres://localhost/db
export SECRET_KEY=mysecret
QUOTED="quoted value"
MULTILINE="""
line 1
line 2
"""
`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewParser()
	variables, err := parser.ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	expected := []Variable{
		{Name: "DATABASE_URL", Value: "postgres://localhost/db"},
		{Name: "SECRET_KEY", Value: "mysecret"},
		{Name: "QUOTED", Value: "quoted value"},
		{Name: "MULTILINE", Value: "line 1\nline 2\n"},
	}

	if !reflect.DeepEqual(variables, expected) {
		t.Errorf("ParseFile() = %v, want %v", variables, expected)
	}
}

func TestParseTestEnvFile(t *testing.T) {
	// Set up environment variable for testing PWD interpolation
	originalPwd := os.Getenv("PWD")
	os.Setenv("PWD", "/current/working/dir")
	defer func() {
		if originalPwd != "" {
			os.Setenv("PWD", originalPwd)
		} else {
			os.Unsetenv("PWD")
		}
	}()

	parser := NewParser()
	parser.lines = []string{
		"# Test dotenv file",
		"export DATABASE_URL=postgres://user:pass@localhost/db",
		"SIMPLE=xyz123",
		"",
		"# Variable interpolation",
		"USER=admin",
		"EMAIL=${USER}@example.org",
		"CACHE_DIR=${PWD}/cache",
		"",
		"# Quoted values",
		`INTERPOLATED="Multiple\nLines and variable substitution: ${SIMPLE}"`,
		"NON_INTERPOLATED='raw text without variable interpolation'",
		"",
		"# Multiline",
		`PRIVATE_KEY="""`,
		"-----BEGIN RSA PRIVATE KEY-----",
		"HkVN9...",
		"-----END RSA PRIVATE KEY-----",
		`"""`,
		"",
		`SINGLE_QUOTE_MULTILINE='''`,
		"Hello ${PERSON},",
		"Nice to meet you!",
		`'''`,
		"",
		"# Comments and special characters",
		"SECRET_KEY=YOURSECRETKEYGOESHERE # also a comment",
		`SECRET_HASH="something-with-a-hash-#-this-is-not-a-comment"`,
		"PASSWORD='!@G0${k}k'",
		"",
		"# Unicode escape",
		`UNICODE_TEST="Unicode: \\u0041\\u0042\\u0043"`,
		"",
		"# Edge cases",
		"EMPTY_VALUE=",
		"WHITESPACE_VALUE=   some value   ",
	}

	variables, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Check specific variables
	variableMap := make(map[string]string)
	for _, v := range variables {
		variableMap[v.Name] = v.Value
	}

	tests := []struct {
		name     string
		expected string
	}{
		{"DATABASE_URL", "postgres://user:pass@localhost/db"},
		{"SIMPLE", "xyz123"},
		{"USER", "admin"},
		{"EMAIL", "admin@example.org"},
		{"CACHE_DIR", "/current/working/dir/cache"},
		{"INTERPOLATED", "Multiple\nLines and variable substitution: xyz123"},
		{"NON_INTERPOLATED", "raw text without variable interpolation"},
		{"PRIVATE_KEY", "-----BEGIN RSA PRIVATE KEY-----\nHkVN9...\n-----END RSA PRIVATE KEY-----\n"},
		{"SINGLE_QUOTE_MULTILINE", "Hello ${PERSON},\nNice to meet you!\n"},
		{"SECRET_KEY", "YOURSECRETKEYGOESHERE"},
		{"SECRET_HASH", "something-with-a-hash-#-this-is-not-a-comment"},
		{"PASSWORD", "!@G0${k}k"},
		{"UNICODE_TEST", "Unicode: ABC"},
		{"EMPTY_VALUE", ""},
		{"WHITESPACE_VALUE", "some value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if value, exists := variableMap[tt.name]; !exists {
				t.Errorf("Variable %s not found", tt.name)
			} else if value != tt.expected {
				t.Errorf("Variable %s = %q, want %q", tt.name, value, tt.expected)
			}
		})
	}

	// Check total number of variables
	expectedCount := 15
	if len(variables) != expectedCount {
		t.Errorf("Expected %d variables, got %d", expectedCount, len(variables))
	}
}

// Benchmark tests
func BenchmarkParseSimple(b *testing.B) {
	lines := []string{
		"DATABASE_URL=postgres://localhost/db",
		"SECRET_KEY=mysecret",
		"PORT=3000",
	}

	for i := 0; i < b.N; i++ {
		parser := NewParser()
		parser.lines = lines
		_, err := parser.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseComplex(b *testing.B) {
	lines := []string{
		"# Complex configuration",
		"export DB_HOST=localhost",
		`DB_URL="postgres://${DB_HOST}:5432/db"`,
		`MULTILINE="""`,
		`line 1`,
		`line 2`,
		`"""`,
		`ESCAPED="value\nwith\tescapes"`,
		"SIMPLE=value",
	}

	for i := 0; i < b.N; i++ {
		parser := NewParser()
		parser.lines = lines
		_, err := parser.Parse()
		if err != nil {
			b.Fatal(err)
		}
	}
}
