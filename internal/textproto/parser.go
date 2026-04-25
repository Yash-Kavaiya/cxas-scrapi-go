// Package textproto parses a subset of Protocol Buffer text format (textproto).
// This is used to parse evaluation export files from the CES API.
package textproto

import (
	"strings"
	"unicode"
)

type tokenKind int

const (
	tokIdent  tokenKind = iota // field name or unquoted value
	tokColon                   // :
	tokLBrace                  // {
	tokRBrace                  // }
	tokString                  // "quoted string"
	tokNumber                  // numeric literal
)

type token struct {
	kind  tokenKind
	value string
}

// Parse parses a textproto string into a nested map[string]interface{}.
// Repeated fields become []interface{}. Nested messages become map[string]interface{}.
func Parse(text string) map[string]interface{} {
	tokens := tokenize(text)
	result, _ := parseMessage(tokens, 0)
	return result
}

func tokenize(text string) []token {
	var tokens []token
	i := 0
	for i < len(text) {
		ch := text[i]

		// Skip whitespace and comments
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			i++
			continue
		}
		if ch == '#' {
			for i < len(text) && text[i] != '\n' {
				i++
			}
			continue
		}

		switch ch {
		case ':':
			tokens = append(tokens, token{tokColon, ":"})
			i++
		case '{':
			tokens = append(tokens, token{tokLBrace, "{"})
			i++
		case '}':
			tokens = append(tokens, token{tokRBrace, "}"})
			i++
		case '"', '\'':
			quote := ch
			i++
			var sb strings.Builder
			for i < len(text) && text[i] != quote {
				if text[i] == '\\' && i+1 < len(text) {
					i++
					switch text[i] {
					case 'n':
						sb.WriteByte('\n')
					case 't':
						sb.WriteByte('\t')
					case '\\':
						sb.WriteByte('\\')
					case '"':
						sb.WriteByte('"')
					case '\'':
						sb.WriteByte('\'')
					default:
						sb.WriteByte('\\')
						sb.WriteByte(text[i])
					}
				} else {
					sb.WriteByte(text[i])
				}
				i++
			}
			if i < len(text) {
				i++ // consume closing quote
			}
			tokens = append(tokens, token{tokString, sb.String()})
		default:
			if unicode.IsLetter(rune(ch)) || ch == '_' {
				start := i
				for i < len(text) && (unicode.IsLetter(rune(text[i])) || unicode.IsDigit(rune(text[i])) || text[i] == '_' || text[i] == '-' || text[i] == '.') {
					i++
				}
				tokens = append(tokens, token{tokIdent, text[start:i]})
			} else if unicode.IsDigit(rune(ch)) || ch == '-' {
				start := i
				if ch == '-' {
					i++
				}
				for i < len(text) && (unicode.IsDigit(rune(text[i])) || text[i] == '.') {
					i++
				}
				tokens = append(tokens, token{tokNumber, text[start:i]})
			} else {
				i++ // skip unknown characters
			}
		}
	}
	return tokens
}

// parseMessage reads key-value pairs until a closing '}' or end of tokens.
// Returns the parsed map and the index of the next unconsumed token.
func parseMessage(tokens []token, pos int) (map[string]interface{}, int) {
	result := make(map[string]interface{})

	for pos < len(tokens) {
		tok := tokens[pos]

		// End of nested message
		if tok.kind == tokRBrace {
			pos++ // consume '}'
			return result, pos
		}

		// Expect a field name (ident)
		if tok.kind != tokIdent {
			pos++
			continue
		}
		fieldName := tok.value
		pos++

		if pos >= len(tokens) {
			break
		}

		next := tokens[pos]

		// Field with nested message: field_name { ... }
		if next.kind == tokLBrace {
			pos++ // consume '{'
			nested, newPos := parseMessage(tokens, pos)
			pos = newPos
			setOrAppend(result, fieldName, nested)
			continue
		}

		// Field with colon: field_name: value
		if next.kind == tokColon {
			pos++ // consume ':'
			if pos >= len(tokens) {
				break
			}
			valTok := tokens[pos]
			pos++

			switch valTok.kind {
			case tokLBrace:
				nested, newPos := parseMessage(tokens, pos)
				pos = newPos
				setOrAppend(result, fieldName, nested)
			case tokString:
				setOrAppend(result, fieldName, valTok.value)
			case tokIdent:
				// boolean or enum
				switch valTok.value {
				case "true":
					setOrAppend(result, fieldName, true)
				case "false":
					setOrAppend(result, fieldName, false)
				default:
					setOrAppend(result, fieldName, valTok.value)
				}
			case tokNumber:
				setOrAppend(result, fieldName, valTok.value)
			}
			continue
		}

		// Skip unexpected token
		pos++
	}

	return result, pos
}

// setOrAppend sets result[key] = value, but if the key already exists it
// converts the value to a []interface{} (repeated field semantics).
func setOrAppend(m map[string]interface{}, key string, value interface{}) {
	existing, exists := m[key]
	if !exists {
		m[key] = value
		return
	}
	switch v := existing.(type) {
	case []interface{}:
		m[key] = append(v, value)
	default:
		m[key] = []interface{}{existing, value}
	}
}
