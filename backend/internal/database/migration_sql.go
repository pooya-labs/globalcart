package database

import (
	"fmt"
	"strings"
)

// splitSQLStatements separates PostgreSQL migration statements without splitting
// quoted text, identifiers, comments or dollar-quoted function bodies.
func splitSQLStatements(source string) ([]string, error) {
	const (
		normal = iota
		singleQuoted
		doubleQuoted
		lineComment
		blockComment
		dollarQuoted
	)
	state, blockDepth := normal, 0
	dollarTag, escapeString := "", false
	start := 0
	statements := make([]string, 0)
	for i := 0; i < len(source); {
		switch state {
		case singleQuoted:
			if escapeString && source[i] == '\\' {
				i += 2
				continue
			}
			if source[i] == '\'' {
				if i+1 < len(source) && source[i+1] == '\'' {
					i += 2
					continue
				}
				state = normal
			}
			i++
		case doubleQuoted:
			if source[i] == '"' {
				if i+1 < len(source) && source[i+1] == '"' {
					i += 2
					continue
				}
				state = normal
			}
			i++
		case lineComment:
			if source[i] == '\n' {
				state = normal
			}
			i++
		case blockComment:
			if i+1 < len(source) && source[i:i+2] == "/*" {
				blockDepth++
				i += 2
				continue
			}
			if i+1 < len(source) && source[i:i+2] == "*/" {
				blockDepth--
				i += 2
				if blockDepth == 0 {
					state = normal
				}
				continue
			}
			i++
		case dollarQuoted:
			if strings.HasPrefix(source[i:], dollarTag) {
				i += len(dollarTag)
				dollarTag, state = "", normal
				continue
			}
			i++
		default:
			switch {
			case i+1 < len(source) && source[i:i+2] == "--":
				state, i = lineComment, i+2
			case i+1 < len(source) && source[i:i+2] == "/*":
				state, blockDepth, i = blockComment, 1, i+2
			case source[i] == '\'':
				state = singleQuoted
				escapeString = i > 0 && (source[i-1] == 'e' || source[i-1] == 'E') &&
					(i == 1 || !isSQLIdentifierByte(source[i-2]))
				i++
			case source[i] == '"':
				state = doubleQuoted
				i++
			case source[i] == '$':
				if delimiter, ok := sqlDollarQuoteDelimiter(source, i); ok {
					dollarTag, state = delimiter, dollarQuoted
					i += len(delimiter)
				} else {
					i++
				}
			case source[i] == ';':
				if statement := strings.TrimSpace(source[start:i]); statement != "" {
					statements = append(statements, statement)
				}
				start, i = i+1, i+1
			default:
				i++
			}
		}
	}
	switch state {
	case singleQuoted:
		return nil, fmt.Errorf("unterminated SQL string literal")
	case doubleQuoted:
		return nil, fmt.Errorf("unterminated SQL quoted identifier")
	case blockComment:
		return nil, fmt.Errorf("unterminated SQL block comment")
	case dollarQuoted:
		return nil, fmt.Errorf("unterminated SQL dollar-quoted body %s", dollarTag)
	}
	if tail := strings.TrimSpace(source[start:]); tail != "" {
		statements = append(statements, tail)
	}
	return statements, nil
}

func sqlDollarQuoteDelimiter(source string, start int) (string, bool) {
	end := strings.IndexByte(source[start+1:], '$')
	if end < 0 {
		return "", false
	}
	tag := source[start+1 : start+1+end]
	if tag != "" {
		if !isSQLIdentifierStart(tag[0]) {
			return "", false
		}
		for i := 1; i < len(tag); i++ {
			if !isSQLIdentifierByte(tag[i]) {
				return "", false
			}
		}
	}
	return source[start : start+end+2], true
}

func isSQLIdentifierStart(value byte) bool {
	return value == '_' || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z'
}

func isSQLIdentifierByte(value byte) bool {
	return isSQLIdentifierStart(value) || value >= '0' && value <= '9' || value == '$'
}
