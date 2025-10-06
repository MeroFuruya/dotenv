package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type Variable struct {
	Name  string
	Value string
}

type Parser struct {
	variables []Variable
	lines     []string
	position  int
}

func NewParser() *Parser {
	return &Parser{
		variables: make([]Variable, 0),
		lines:     make([]string, 0),
		position:  0,
	}
}

var validVarNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func isValidVariableName(name string) bool {
	return validVarNameRegex.MatchString(name)
}

func (p *Parser) ParseFile(filename string) ([]Variable, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		p.lines = append(p.lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return p.Parse()
}

func (p *Parser) Parse() ([]Variable, error) {
	for p.position < len(p.lines) {
		if err := p.parseLine(); err != nil {
			return nil, fmt.Errorf("error parsing line %d: %w", p.position+1, err)
		}
		p.position++
	}
	return p.variables, nil
}

func (p *Parser) parseLine() error {
	line := p.lines[p.position]

	if strings.TrimSpace(line) == "" {
		return nil
	}

	if strings.TrimSpace(line)[0] == '#' {
		return nil
	}

	line = strings.TrimPrefix(strings.TrimSpace(line), "export ")
	line = strings.TrimSpace(line)

	eqIndex := strings.IndexByte(line, '=')
	if eqIndex == -1 {
		return nil
	}

	keyPart := strings.TrimSpace(line[:eqIndex])
	valuePart := line[eqIndex+1:]

	if !isValidVariableName(keyPart) {
		return fmt.Errorf("invalid variable name: %s", keyPart)
	}

	value, err := p.parseValue(valuePart)
	if err != nil {
		return err
	}

	p.variables = append(p.variables, Variable{Name: keyPart, Value: value})
	return nil
}

func (p *Parser) parseValue(valuePart string) (string, error) {
	valuePart = strings.TrimLeftFunc(valuePart, unicode.IsSpace)

	if len(valuePart) == 0 {
		return "", nil
	}

	if strings.HasPrefix(valuePart, `"""`) || strings.HasPrefix(valuePart, `'''`) {
		return p.parseMultilineValue(valuePart)
	}

	if len(valuePart) > 0 && (valuePart[0] == '"' || valuePart[0] == '\'') {
		return p.parseQuotedValue(valuePart)
	}

	return p.parseUnquotedValue(valuePart)
}

func (p *Parser) parseMultilineValue(valuePart string) (string, error) {
	var delimiter string
	var interpolate bool

	if strings.HasPrefix(valuePart, `"""`) {
		delimiter = `"""`
		interpolate = true
		valuePart = valuePart[3:]
	} else if strings.HasPrefix(valuePart, `'''`) {
		delimiter = `'''`
		interpolate = false
		valuePart = valuePart[3:]
	} else {
		return "", fmt.Errorf("invalid multiline delimiter")
	}

	var result strings.Builder

	// If the rest of the line after the opening delimiter is not empty, include it
	if strings.TrimSpace(valuePart) != "" {
		if strings.HasSuffix(valuePart, delimiter) {
			// Single line triple-quoted string
			content := valuePart[:len(valuePart)-3]
			if interpolate {
				return p.processEscapesAndInterpolation(content)
			}
			return content, nil
		}
		result.WriteString(valuePart)
		result.WriteString("\n")
	}

	// Continue reading lines until we find the closing delimiter
	p.position++
	for p.position < len(p.lines) {
		line := p.lines[p.position]
		if strings.HasSuffix(line, delimiter) {
			// Found closing delimiter
			content := line[:len(line)-3]
			result.WriteString(content)
			break
		}
		result.WriteString(line)
		result.WriteString("\n")
		p.position++
	}

	if p.position >= len(p.lines) {
		return "", fmt.Errorf("unterminated multiline string")
	}

	finalValue := result.String()
	if interpolate {
		return p.processEscapesAndInterpolation(finalValue)
	}
	return finalValue, nil
}

// parseQuotedValue parses single or double quoted values
func (p *Parser) parseQuotedValue(valuePart string) (string, error) {
	quote := valuePart[0]
	interpolate := quote == '"'

	var result strings.Builder
	i := 1 // Skip opening quote

	for i < len(valuePart) {
		if valuePart[i] == quote {
			// Found closing quote
			finalValue := result.String()
			if interpolate {
				return p.processEscapesAndInterpolation(finalValue)
			}
			return finalValue, nil
		}

		if valuePart[i] == '\\' && i+1 < len(valuePart) {
			// Handle escape sequences
			if interpolate {
				nextChar := valuePart[i+1]
				switch nextChar {
				case 'n':
					result.WriteByte('\n')
				case 'r':
					result.WriteByte('\r')
				case 't':
					result.WriteByte('\t')
				case 'f':
					result.WriteByte('\f')
				case 'b':
					result.WriteByte('\b')
				case '"':
					result.WriteByte('"')
				case '\'':
					result.WriteByte('\'')
				case '\\':
					result.WriteByte('\\')
				case 'u':
					if i+5 <= len(valuePart) {
						hexCode := valuePart[i+2 : i+6]
						if codepoint, err := strconv.ParseInt(hexCode, 16, 32); err == nil {
							result.WriteRune(rune(codepoint))
							i += 6 // Skip '\', 'u' and 4 hex digits
							continue
						}
					}
					// Invalid unicode escape, treat literally
					result.WriteByte(nextChar)
				default:
					// Any other character after backslash is treated literally
					result.WriteByte(nextChar)
				}
				i += 2
			} else {
				// In single quotes, only escape single quotes and backslashes
				nextChar := valuePart[i+1]
				if nextChar == '\'' || nextChar == '\\' {
					result.WriteByte(nextChar)
					i += 2
				} else {
					result.WriteByte('\\')
					i++
				}
			}
		} else {
			result.WriteByte(valuePart[i])
			i++
		}
	}

	return "", fmt.Errorf("unterminated quoted string")
}

// parseUnquotedValue parses unquoted values
func (p *Parser) parseUnquotedValue(valuePart string) (string, error) {
	// Find comment start (not within quotes)
	commentIndex := -1
	for i, char := range valuePart {
		if char == '#' {
			commentIndex = i
			break
		}
	}

	if commentIndex != -1 {
		valuePart = valuePart[:commentIndex]
	}

	value := strings.TrimRightFunc(valuePart, unicode.IsSpace)
	return p.processEscapesAndInterpolation(value)
}

// processEscapesAndInterpolation processes escape sequences and variable interpolation
func (p *Parser) processEscapesAndInterpolation(value string) (string, error) {
	// First process escape sequences
	value = p.processEscapeSequences(value)

	// Then process variable interpolation
	return p.interpolateVariables(value), nil
}

// processEscapeSequences processes escape sequences in the value
func (p *Parser) processEscapeSequences(value string) string {
	var result strings.Builder
	i := 0

	for i < len(value) {
		if value[i] == '\\' && i+1 < len(value) {
			nextChar := value[i+1]
			switch nextChar {
			case 'n':
				result.WriteByte('\n')
			case 'r':
				result.WriteByte('\r')
			case 't':
				result.WriteByte('\t')
			case 'f':
				result.WriteByte('\f')
			case 'b':
				result.WriteByte('\b')
			case '"':
				result.WriteByte('"')
			case '\'':
				result.WriteByte('\'')
			case '\\':
				result.WriteByte('\\')
			case 'u':
				if i+5 <= len(value) {
					hexCode := value[i+2 : i+6]
					if codepoint, err := strconv.ParseInt(hexCode, 16, 32); err == nil {
						result.WriteRune(rune(codepoint))
						i += 6 // Skip '\', 'u' and 4 hex digits
						continue
					}
				}
				// Invalid unicode escape, treat literally
				result.WriteByte(nextChar)
			default:
				// Any other character after backslash is treated literally
				result.WriteByte(nextChar)
			}
			i += 2
		} else {
			result.WriteByte(value[i])
			i++
		}
	}

	return result.String()
}

// interpolateVariables performs variable substitution using ${VAR} syntax
func (p *Parser) interpolateVariables(value string) string {
	var result strings.Builder
	i := 0

	for i < len(value) {
		if i+1 < len(value) && value[i] == '$' && value[i+1] == '{' {
			// Find the closing brace
			closeIndex := strings.IndexByte(value[i+2:], '}')
			if closeIndex != -1 {
				varName := value[i+2 : i+2+closeIndex]

				// Look up the variable value - first check parsed variables
				found := false
				for _, variable := range p.variables {
					if variable.Name == varName {
						result.WriteString(variable.Value)
						found = true
						break
					}
				}

				// If not found in parsed variables, check environment
				if !found {
					if envValue := os.Getenv(varName); envValue != "" {
						result.WriteString(envValue)
					}
					// If variable doesn't exist anywhere, substitute with empty string
				}

				i = i + 2 + closeIndex + 1 // Skip past the }
				continue
			}
		}
		result.WriteByte(value[i])
		i++
	}

	return result.String()
}
