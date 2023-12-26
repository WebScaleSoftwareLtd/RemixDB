// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package ast

import (
	"io"
	"strings"
)

// Parses the using types of a mapping.
func parseMappingUsingTypes(r *strings.Reader, comments *[]CommentToken) ([]string, *ParserError) {
	// Defines the type names.
	typeNames := []string{}
	currentTypeName := ""

	// Get the position.
	pos := getReaderPos(r)

	// Defines the state.
	state := 0

	// Defines the loop.
	for {
		// Read the next Unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			return nil, &ParserError{
				Message:  "expected using type, got EOF",
				Position: pos,
			}
		}

		// Switch on the character.
		switch c {
		case ' ', '\t', '\n', '\r':
			// If it is whitespace, check the state.
			if state == 1 {
				// If it is 1, add the type name to the list.
				typeNames = append(typeNames, currentTypeName)
				currentTypeName = ""
				state = 2
			}
		case ',':
			// If it is a comma, check the state.
			switch state {
			case 0:
				// Comma before anything else is an error.
				return nil, &ParserError{
					Message:  "expected using type, got ','",
					Position: pos,
				}
			case 1:
				// This is just the type name, meaning we can add it to the list.
				typeNames = append(typeNames, currentTypeName)
				currentTypeName = ""
				state = 0
			case 2:
				// Reset the state.
				state = 0
			}
		case '{':
			// If the state is 0, there is no type name.
			if state == 0 && len(typeNames) == 0 {
				return nil, &ParserError{
					Message:  "expected using type, got '{'",
					Position: pos,
				}
			}

			// Rewind a rune and return.
			_ = r.UnreadRune()
			if currentTypeName != "" {
				typeNames = append(typeNames, currentTypeName)
			}
			return typeNames, nil
		case '/':
			// If it is a slash, parse a comment.
			a := []any{}
			err := parseComment(r, &a)
			if err != nil {
				return []string{}, err
			}
			*comments = append(*comments, a[0].(CommentToken))

			// Gulps the whitespace.
			gulpWhitespace(r)
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
			'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't',
			'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D',
			'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N',
			'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X',
			'Y', 'Z', '0', '1', '2', '3', '4', '5', '6', '7',
			'8', '9':
			// Handle the state.
			switch state {
			case 0, 1:
				// Add the character to the type name.
				state = 1
				currentTypeName += string(c)
			case 2:
				// Expected a comma.
				return nil, &ParserError{
					Message:  "expected ',' after using type, got '" + string(c) + "'",
					Position: pos,
				}
			}
		}
	}
}

// Parses the inner of a mapping.
func parseMappingInner(r *strings.Reader, comments *[]CommentToken) (string, any, *ParserError) {
	// Gulps the whitespace.
	_, err := gulpWhitespace(r)
	if err != nil {
		return "", nil, err
	}

	// Get the start of the mapping insides index.
	pos := getReaderPos(r)

	// Get the key.
	key := ""
keyLoop:
	for {
		// Read the next Unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			return "", nil, &ParserError{
				Message:  "expected mapping key, got EOF",
				Position: pos,
			}
		}

		// Switch on the character.
		switch c {
		case '-':
			// Rewind a rune and break.
			_ = r.UnreadRune()
			break keyLoop
		case ' ', '\t', '\n', '\r':
			// If it is whitespace, break.
			break keyLoop
		case '/':
			// If it is a slash, parse a comment.
			a := []any{}
			err := parseComment(r, &a)
			if err != nil {
				return "", nil, err
			}
			*comments = append(*comments, a[0].(CommentToken))

			// Gulps the whitespace.
			gulpWhitespace(r)

			// Set the position.
			pos = getReaderPos(r)
		default:
			// If it is anything else, add it to the key.
			key += string(c)
		}
	}

	// Make sure the key isn't empty.
	if len(key) == 0 {
		return "", nil, &ParserError{
			Message:  "mapping key cannot be empty",
			Position: pos,
		}
	}

	for {
		// Gulps the whitespace.
		_, err = gulpWhitespace(r)
		if err != nil {
			return "", nil, err
		}

		// Get the next 2 bytes.
		b := make([]byte, 2)
		_, _ = r.Read(b)

		// If they are '//', rewind one and parse a comment.
		if string(b) == "//" {
			_, _ = r.Seek(-1, io.SeekCurrent)
			a := []any{}
			err := parseComment(r, &a)
			if err != nil {
				return "", nil, err
			}
			*comments = append(*comments, a[0].(CommentToken))
			continue
		}

		// Make sure the next 2 characters are '->'.
		if string(b) != "->" {
			return "", nil, &ParserError{
				Message:  "expected '->' after mapping key, got '" + string(b) + "'",
				Position: pos,
			}
		}
		break
	}

	for {
		// Gulps the whitespace.
		_, err = gulpWhitespace(r)
		if err != nil {
			return "", nil, err
		}

		// Get the next character.
		c, _, errIface := r.ReadRune()
		if errIface != nil {
			return "", nil, &ParserError{
				Message:  "expected mapping value, got EOF",
				Position: pos,
			}
		}

		// Switch on the character.
		switch c {
		case '/':
			// If it is a slash, parse a comment.
			a := []any{}
			err := parseComment(r, &a)
			if err != nil {
				return "", nil, err
			}

			// Add the comment.
			*comments = append(*comments, a[0].(CommentToken))
		case '{':
			// Get the position of the mapping.
			pos := getReaderPos(r) - 1

			// Get the mapping we will embed.
			commentsInner := []CommentToken{}
			key, value, err := parseMappingInner(r, &commentsInner)
			if err != nil {
				return "", nil, err
			}
			innerMapping := MappingPartialToken{
				Value:    value,
				Position: pos,
				Key:      key,
				Comments: commentsInner,
			}

			// Gulp the whitespace.
			_, err = gulpWhitespace(r)
			if err != nil {
				return "", nil, err
			}

			// Make sure the next character is '}'.
			c, _, errIface = r.ReadRune()
			if errIface != nil {
				return "", nil, &ParserError{
					Message:  "expected '}' after mapping value, got EOF",
					Position: pos,
				}
			}
			if c != '}' {
				return "", nil, &ParserError{
					Message:  "expected '}' after mapping value, got '" + string(c) + "'",
					Position: pos,
				}
			}

			// Return the mapping.
			return key, innerMapping, nil
		case '}':
			// This is a syntax error because it is pointing to the end of the mapping.
			return "", nil, &ParserError{
				Message:  "expected mapping value, got '}'",
				Position: pos,
			}
		default:
			// Start consuming the value.
			value := string(c)
		valueLoop:
			for {
				// Read the next Unicode character.
				c, _, err := r.ReadRune()
				if err != nil {
					return "", nil, &ParserError{
						Message:  "expected mapping value, got EOF",
						Position: pos,
					}
				}

				// Switch on the character.
				switch c {
				case '/':
					// If it is a slash, parse a comment.
					a := []any{}
					err := parseComment(r, &a)
					if err != nil {
						return "", nil, err
					}

					// Add the comment.
					*comments = append(*comments, a[0].(CommentToken))
				case ' ', '\t', '\n', '\r':
					// If it is whitespace, break unless value is empty.
					if len(value) != 0 {
						break valueLoop
					}
				case '}':
					// If it is the end bracket, return.
					if len(value) == 0 {
						return "", nil, &ParserError{
							Message:  "mapping value cannot be empty",
							Position: pos,
						}
					}
					return key, value, nil
				default:
					// If it is anything else, add it to the value.
					value += string(c)
				}
			}

			// Make sure the value isn't empty.
			if len(value) == 0 {
				return "", nil, &ParserError{
					Message:  "mapping value cannot be empty",
					Position: pos,
				}
			}

			// Gulps the whitespace.
			_, err = gulpWhitespace(r)
			if err != nil {
				return "", nil, err
			}

			// Make sure the next character is '}'.
			c, _, errIface = r.ReadRune()
			if errIface != nil {
				return "", nil, &ParserError{
					Message:  "expected '}' after mapping value, got EOF",
					Position: pos,
				}
			}
			if c != '}' {
				return "", nil, &ParserError{
					Message:  "expected '}' after mapping value, got '" + string(c) + "'",
					Position: pos,
				}
			}

			// Return the value.
			return key, value, nil
		}
	}
}

// Parses the mapping keyword. The first 'm' should have been read already.
func parseMapping(r *strings.Reader, decorators []DecoratorToken) (MappingToken, *ParserError) {
	// Get the position with one subtracted because it was already read.
	pos := getReaderPos(r) - 1

	// Make sure the next bit is 'apping'.
	b := make([]byte, 6)
	r.Read(b)
	if string(b) != "apping" {
		return MappingToken{}, &ParserError{
			Message:  "expected 'mapping', got '" + string(b) + "'",
			Position: pos,
		}
	}

	// Make sure there was whitespace.
	wasWhitespace, err := gulpWhitespace(r)
	if err != nil {
		return MappingToken{}, err
	}
	if !wasWhitespace {
		return MappingToken{}, &ParserError{
			Message:  "expected whitespace after 'mapping'",
			Position: pos,
		}
	}

	// Read the name of the mapping.
	comments := []CommentToken{}
	name := ""
nameLoop:
	for {
		// Read the next Unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			return MappingToken{}, &ParserError{
				Message:  "expected mapping name, got EOF",
				Position: pos,
			}
		}

		// Switch on the character.
		switch c {
		case '{':
			// If it's the start bracket, rewind a rune and fall through.
			_ = r.UnreadRune()
			fallthrough
		case ' ', '\t', '\n', '\r':
			// If it is whitespace, break.
			break nameLoop
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
			'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't',
			'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D',
			'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N',
			'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X',
			'Y', 'Z':
			// If it is a letter, add it to the name.
			name += string(c)
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// If it is a number, add it to the name if it isn't the first character.
			if len(name) == 0 {
				return MappingToken{}, &ParserError{
					Message:  "mapping name cannot start with a number",
					Position: pos,
				}
			}
			name += string(c)
		case '/':
			// If it is a slash, parse a comment.
			a := []any{}
			err := parseComment(r, &a)
			if err != nil {
				return MappingToken{}, err
			}
			comments = append(comments, a[0].(CommentToken))

			// Gulps the whitespace.
			gulpWhitespace(r)
		default:
			// If it is anything else, error.
			return MappingToken{}, &ParserError{
				Message:  "unexpected character '" + string(c) + "' in mapping name",
				Position: pos,
			}
		}
	}

	// Make sure the name isn't empty.
	if len(name) == 0 {
		return MappingToken{}, &ParserError{
			Message:  "mapping name cannot be empty",
			Position: pos,
		}
	}

	// Make sure that we have a 'using' after the name.
	b = make([]byte, 5)
	_, _ = r.Read(b)
	if string(b) != "using" {
		return MappingToken{}, &ParserError{
			Message:  "expected 'using' after mapping keyword, got '" + string(b) + "'",
			Position: pos,
		}
	}

	// Make sure there was whitespace.
	wasWhitespace, err = gulpWhitespace(r)
	if err != nil {
		return MappingToken{}, err
	}
	if !wasWhitespace {
		return MappingToken{}, &ParserError{
			Message:  "expected whitespace after 'using'",
			Position: pos,
		}
	}

	// Read the using types of the mapping.
	usingTypes, err := parseMappingUsingTypes(r, &comments)
	if err != nil {
		return MappingToken{}, err
	}

	// Find the start bracket.
startBracketLoop:
	for {
		// Read the next Unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			return MappingToken{}, &ParserError{
				Message:  "expected '{' after mapping name, got EOF",
				Position: pos,
			}
		}

		// Switch on the character.
		switch c {
		case '{':
			// If it's the start bracket, break.
			break startBracketLoop
		case ' ', '\t', '\n', '\r':
			// Ignore whitespace.
		case '/':
			// If it is a slash, parse a comment.
			a := []any{}
			err := parseComment(r, &a)
			if err != nil {
				return MappingToken{}, err
			}
			comments = append(comments, a[0].(CommentToken))
		default:
			// If it is anything else, error.
			return MappingToken{}, &ParserError{
				Message:  "expected '{' after mapping name, got '" + string(c) + "'",
				Position: pos,
			}
		}
	}

	// Parse the inners of the mapping.
	key, value, err := parseMappingInner(r, &comments)
	if err != nil {
		return MappingToken{}, err
	}

	// Return the mapping.
	return MappingToken{
		MappingPartialToken: MappingPartialToken{
			Position: pos,
			Key:      key,
			Value:    value,
			Comments: comments,
		},
		Name:       name,
		Using:      usingTypes,
		Decorators: decorators,
	}, nil
}
