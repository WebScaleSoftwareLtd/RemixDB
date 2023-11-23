// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package ast

import (
	"math/big"
	"strconv"
	"strings"
)

// Parses a single quoted string. The first single quote has already been read.
func parseSingleQuotedString(r *strings.Reader) (StringLiteralToken, *ParserError) {
	pos := getReaderPos(r) - 1
	content := ""
	esc := false
	for {
		// Read the next Unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			// End of file.
			return StringLiteralToken{}, &ParserError{
				Message:  "unexpected end of file after single quoted string",
				Position: pos,
			}
		}

		// Switch on the character.
		switch c {
		case '\'':
			// If we are escaped, add the character.
			if esc {
				content += string(c)
				esc = false
				continue
			}

			// Return the string literal.
			return StringLiteralToken{
				Value:    content,
				Position: pos,
			}, nil
		case '\\':
			// If we are escaped, add the character.
			if esc {
				content += string(c)
				esc = false
				continue
			}

			// We are now escaped.
			esc = true
		default:
			// Add the character to the content.
			content += string(c)

			// We are no longer escaped if we were.
			esc = false
		}
	}
}

// Parses a double quoted string. The first double quote has already been read.
func parseDoubleQuotedString(r *strings.Reader) (StringLiteralToken, *ParserError) {
	pos := getReaderPos(r) - 1
	content := ""
	esc := false
	for {
		// Read the next Unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			// End of file.
			return StringLiteralToken{}, &ParserError{
				Message:  "unexpected end of file after double quoted string",
				Position: pos,
			}
		}

		// Switch on the character.
		switch c {
		case '"':
			// If we are escaped, add the character.
			if esc {
				content += string(c)
				esc = false
				continue
			}

			// Return the string literal.
			return StringLiteralToken{
				Value:    content,
				Position: pos,
			}, nil
		case '\\':
			// If we are escaped, add the character.
			if esc {
				content += string(c)
				esc = false
				continue
			}

			// We are now escaped.
			esc = true
		default:
			// Add the character to the content.
			content += string(c)

			// We are no longer escaped if we were.
			esc = false
		}
	}
}

// Parses a number literal.
func parseNumberLiteral(r *strings.Reader, eot rune) (any, *ParserError) {
	// Defines the start position.
	pos := getReaderPos(r)

	// Defines the content.
	content := ""

	// Defines the state.
	state := 0

	// Get the number content.
numberGetLoop:
	for {
		// Read the next Unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			// Before end of token so this is a error.
			return NumberLiteralToken{}, &ParserError{
				Message:  "unexpected end of file after number literal",
				Position: pos,
			}
		}

		// Switch on the character.
		switch c {
		case eot, ',', ' ', '\t', '\n', '\r':
			// Rewind the rune.
			_ = r.UnreadRune()

			// Break the loop.
			break numberGetLoop
		case '-':
			// Make sure we are in state 0.
			if state != 0 {
				return NumberLiteralToken{}, &ParserError{
					Message:  "unexpected '-' after start of number",
					Position: pos,
				}
			}

			// We are now in state 2.
			state = 2
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', '+', 'a', 'A',
			'b', 'B', 'c', 'C', 'd', 'D', 'e', 'E', 'f', 'F', 'x', 'o':
			// If we are in state 0, set it to 1.
			if state == 0 {
				state = 1
			}

			// Add the character to the content.
			content += string(c)
		}
	}

	// Check if the content is a float.
	if strings.Contains(content, ".") {
		// If it starts with a dot, add a zero.
		if strings.HasPrefix(content, ".") {
			content = "0" + content
		}

		// Parse the float as a float64.
		float, err := strconv.ParseFloat(content, 64)
		if err != nil {
			return nil, &ParserError{
				Message:  "error while parsing float: " + err.Error(),
				Position: pos,
			}
		}

		// Return the float literal.
		return FloatLiteralToken{
			Value:    float,
			Position: pos,
		}, nil
	}

	// If content is 0, return a zero.
	if content == "0" {
		return NumberLiteralToken{
			Value:    0,
			Position: pos,
		}, nil
	}

	// Check if this is another base.
	base := 10
	switch {
	case strings.HasPrefix(content, "0x"):
		base = 16
		content = content[2:]
	case strings.HasPrefix(content, "0b"):
		base = 2
		content = content[2:]
	case strings.HasPrefix(content, "0"):
		base = 8
		content = content[1:]
		if content[0] == 'o' {
			content = content[1:]
		}
	}

	// Check if the content fits in a 64 bit integer.
	v, err := strconv.ParseInt(content, base, 64)
	if err == nil {
		// Return the number literal.
		return NumberLiteralToken{
			Value:    int(v),
			Position: pos,
		}, nil
	}

	// Check if this is a valid big integer.
	x, _ := big.NewInt(0).SetString(content, base)
	if x != nil {
		// Return the big integer literal.
		return BigIntLiteralToken{
			Value:    content,
			Position: pos,
		}, nil
	}

	// Return a error.
	return nil, &ParserError{
		Message:  "invalid number literal",
		Position: pos,
	}
}

// Parses the inners of an object literal. The first brace has already been read.
func parseObjectLiteral(r *strings.Reader) (ObjectLiteralToken, *ParserError) {
	m := map[string]any{}
	comments := []CommentToken{}
	pos := getReaderPos(r) - 1

	for {
		// Loop to get the information we need.
		name := ""
	nameLoop:
		for {
			// Get the next Unicode character.
			c, _, err := r.ReadRune()
			if err != nil {
				// End of file.
				return ObjectLiteralToken{}, &ParserError{
					Message:  "unexpected end of file after object literal",
					Position: pos,
				}
			}

			// Switch on the character.
			switch c {
			case '}':
				// Handle if a name is present.
				if name != "" {
					return ObjectLiteralToken{}, &ParserError{
						Message:  "unexpected '}' after object literal name",
						Position: pos,
					}
				}

				// Return the object literal.
				return ObjectLiteralToken{
					Values:   m,
					Comments: comments,
					Position: pos,
				}, nil
			case '/':
				// Parse the comment.
				a := []any{}
				e := parseComment(r, &a)
				if e != nil {
					return ObjectLiteralToken{}, e
				}

				// Add the comment.
				comments = append(comments, a[0].(CommentToken))

				// Break the name loop if there's a name.
				if name != "" {
					break nameLoop
				}
			case '=':
				// Rewind the rune.
				_ = r.UnreadRune()

				// If the name is blank, this is a syntax error.
				if name == "" {
					return ObjectLiteralToken{}, &ParserError{
						Message:  "unexpected '=' without object literal name",
						Position: pos,
					}
				}

				// Break the name loop.
				break nameLoop
			case ' ', '\t', '\n', '\r':
				// Handle if a name is present.
				if name != "" {
					break nameLoop
				}
			case '"':
				// Handle if a name is present.
				if name != "" {
					return ObjectLiteralToken{}, &ParserError{
						Message:  "unexpected '\"' after object literal name",
						Position: pos,
					}
				}

				// Parse the double quoted string.
				str, err := parseDoubleQuotedString(r)
				if err != nil {
					return ObjectLiteralToken{}, err
				}

				// Set the name to the string.
				name = str.Value

				// Break the name loop.
				break nameLoop
			case '\'':
				// Handle if a name is present.
				if name != "" {
					return ObjectLiteralToken{}, &ParserError{
						Message:  "unexpected ''' after object literal name",
						Position: pos,
					}
				}

				// Parse the single quoted string.
				str, err := parseSingleQuotedString(r)
				if err != nil {
					return ObjectLiteralToken{}, err
				}

				// Set the name to the string.
				name = str.Value

				// Break the name loop.
				break nameLoop
			default:
				// Add the character to the name.
				name += string(c)
			}
		}

		// Consume all of the whitespace.
		gulpWhitespace(r)

		// Get the next character.
		c, _, err := r.ReadRune()
		if err != nil {
			// End of file.
			return ObjectLiteralToken{}, &ParserError{
				Message:  "unexpected end of file after object literal name",
				Position: pos,
			}
		}

		// Make sure this is a equals sign.
		if c != '=' {
			return ObjectLiteralToken{}, &ParserError{
				Message:  "unexpected '" + string(c) + "' after object literal name",
				Position: pos,
			}
		}

		// Consume all of the whitespace.
		gulpWhitespace(r)

		// Read the argument.
		arg, perr := parseInnerContractTokenWithOpGrouping(r, ')')
		if perr != nil {
			// The argument was unable to be read.
			return ObjectLiteralToken{}, perr
		}

		// Add the argument.
		m[name] = arg
	}
}

// Parses the inners of an array literal. The first bracket has already been read.
func parseArrayLiteral(r *strings.Reader) (ArrayLiteralToken, *ParserError) {
	items := []any{}
	pos := getReaderPos(r) - 1
	for {
		// Get the next token.
		token, err := parseInnerContractTokenWithOpGrouping(r, ']')
		if err != nil {
			return ArrayLiteralToken{}, err
		}
		if token == nil {
			// Return the array literal.
			return ArrayLiteralToken{
				Values:   items,
				Position: pos,
			}, nil
		}

		// Add the token.
		items = append(items, token)

		// Consume the next spaces until either a comma or closing bracket.
	commaLoop:
		for {
			// Read the next Unicode character.
			c, _, err := r.ReadRune()
			if err != nil {
				// End of file.
				return ArrayLiteralToken{}, &ParserError{
					Message:  "unexpected end of file after array literal",
					Position: pos,
				}
			}

			// Switch on the character.
			switch c {
			case ' ', '\t', '\n', '\r':
				// Ignore whitespace.
			case ',':
				// Break the comma loop.
				break commaLoop
			case ']':
				// Return the array literal.
				return ArrayLiteralToken{
					Values:   items,
					Position: pos,
				}, nil
			}
		}
	}
}
