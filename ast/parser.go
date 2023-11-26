// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package ast

import (
	"regexp"
	"strings"
)

// ParserError is used to define an error that occurred while parsing.
type ParserError struct {
	// Message is the error message.
	Message string

	// Position is the position of the error.
	Position int
}

func getReaderPos(r *strings.Reader) int {
	return int(r.Size()) - r.Len()
}

// Gulps the whitespace.
func gulpWhitespace(r *strings.Reader) (bool, *ParserError) {
	pos := getReaderPos(r)
	wasWhitespace := false
	for {
		// Read the next Unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			return false, &ParserError{
				Message:  "expected whitespace, got EOF",
				Position: pos,
			}
		}

		// Check if it is whitespace.
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			// If it is \r, check if the next character is \n.
			if c == '\r' {
				c, _, err = r.ReadRune()
				if err != nil {
					return false, &ParserError{
						Message:  "expected whitespace, got EOF",
						Position: pos,
					}
				}
				if c != '\n' {
					return false, &ParserError{
						Message:  "expected whitespace, got '" + string(c) + "'",
						Position: pos,
					}
				}
			}

			wasWhitespace = true
		} else {
			// If it wasn't whitespace, put it back and break.
			r.UnreadRune()
			break
		}
	}

	return wasWhitespace, nil
}

// Parses a comment. This assumes that the starting slash has already been read.
func parseComment(r *strings.Reader, tokens *[]any) *ParserError {
	// Get the position of the comment.
	pos := getReaderPos(r) - 1

	// Read the next Unicode character.
	c, _, err := r.ReadRune()
	if err != nil {
		// End of file.
		return nil
	}

	// Switch on the character.
	switch c {
	case '/':
		// Parse a single line comment.
		comment := ""
	commentLoop:
		for {
			// Read the next Unicode character.
			c, _, err := r.ReadRune()
			if err != nil {
				// End of file.
				break
			}

			// Switch on the character.
			switch c {
			case '\r':
				// Read the next Unicode character.
				c, _, err := r.ReadRune()
				if err != nil {
					// End of file.
					break commentLoop
				}
				if c == '\n' {
					// End of the comment.
					break commentLoop
				}

				// Add the character to the comment.
				comment += string(c)
			case '\n':
				// End of the comment.
				break commentLoop
			default:
				// Add the character to the comment.
				comment += string(c)
			}
		}

		// Add the comment token.
		*tokens = append(*tokens, CommentToken{
			Comment:  comment,
			Position: pos,
		})
	default:
		// Unexpected character.
		return &ParserError{
			Message:  "unexpected character '" + string(c) + "' after '/'",
			Position: pos,
		}
	}
	return nil
}

const unexpectedEofAfterDeco = "unexpected end of file after decorator"

// Parses a decorator. This assumes that the starting at has already been read.
func parseDecorator(r *strings.Reader) (DecoratorToken, *ParserError) {
	// Get the position of the decorator. We subtract 1 because the at has already been read.
	pos := getReaderPos(r) - 1

	// Get the decorator name.
	name := ""
nameReader:
	for {
		// Read the next unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			// End of file.
			return DecoratorToken{}, &ParserError{
				Message:  unexpectedEofAfterDeco,
				Position: pos,
			}
		}

		// Switch on the character.
		switch c {
		case ' ', '\t':
			// No whitespace allowed before brackets.
			return DecoratorToken{}, &ParserError{
				Message:  "unexpected whitespace before decorator brackets",
				Position: pos,
			}
		case '\n', '\r':
			// Gulps the whitespace.
			gulpWhitespace(r)

			// This means the decorator arguments are empty.
			return DecoratorToken{
				Method:   name,
				Position: pos,
			}, nil
		case '(':
			// We found the end of the decorator name.
			break nameReader
		default:
			// Add the character to the name.
			name += string(c)
		}
	}

	// Make sure the name is not empty.
	if name == "" {
		return DecoratorToken{}, &ParserError{
			Message:  "unexpected empty name after '@' character",
			Position: pos,
		}
	}

	// Consume everything in the token until the closing bracket.
	content := ""
contentReader:
	for {
		// Read the next unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			// End of file.
			return DecoratorToken{}, &ParserError{
				Message:  unexpectedEofAfterDeco,
				Position: pos,
			}
		}

		// Switch on the character.
		switch c {
		case ')':
			// We found the end of the decorator.
			break contentReader
		default:
			// Add the character to the content.
			content += string(c)
		}
	}

	// Trim the content.
	content = strings.TrimSpace(content)

	// Gobble the whitespace.
	gulpWhitespace(r)

	// Return the decorator token.
	return DecoratorToken{
		Method:    name,
		Position:  pos,
		Arguments: content,
	}, nil
}

// Parses the extends keyword. This assumes the starting e of extends has already been read.
func parseExtends(r *strings.Reader) *ParserError {
	// Get the position of the extends with 1 subtracted because we already read the character.
	pos := getReaderPos(r) - 1

	// Make sure the next content is 'xtends'.
	b := make([]byte, 6)
	_, err := r.Read(b)
	if err != nil {
		// End of file.
		return &ParserError{
			Message:  "unexpected end of file after 'e' character",
			Position: pos,
		}
	}

	// Make sure the next content is 'xtends'.
	content := string(b)
	if content != "xtends" {
		return &ParserError{
			Message:  "unexpected '" + content + "' after 'e' character - did you mean extends?",
			Position: pos,
		}
	}

	// Return no errors.
	return nil
}

var typeSplitRegex = regexp.MustCompile(":[ \t]*[a-zA-Z]")

// Parses a struct field or reference. c is the initial character that was consumed.
func parseStructFieldOrReference(r *strings.Reader, c rune, decorators []DecoratorToken) (any, *ParserError) {
	// Get the position of the field or reference with 1 subtracted because we already read the character.
	pos := getReaderPos(r) - 1

	// Parse until the newline.
	content := ""
	var err error
contentParse:
	for {
		// Switch on the character.
		switch c {
		case '\n':
			// End of the content.
			break contentParse
		case '\r':
			// Ignore.
		default:
			// Add the character to the content.
			content += string(c)
		}

		// Read the next Unicode character.
		c, _, err = r.ReadRune()
		if err != nil {
			// End of file.
			return nil, &ParserError{
				Message:  "unexpected end of file after start of field or reference",
				Position: pos,
			}
		}
	}

	// Trim the content.
	content = strings.TrimSpace(content)

	// Make sure the content is not empty.
	if content == "" {
		return nil, &ParserError{
			Message:  "unexpected end of file after start of field or reference",
			Position: pos,
		}
	}

	// Handle checking for the type split.
	split := typeSplitRegex.FindStringIndex(content)
	if split == nil {
		// This is a reference.
		return ReferenceToken{
			Name:       content,
			Position:   pos,
			Decorators: decorators,
		}, nil
	}

	// This is a field.
	return FieldToken{
		Name:       content[:split[0]],
		Type:       content[split[1]-1:],
		Position:   pos,
		Decorators: decorators,
	}, nil
}

// Parses the inners of a struct.
func parseInnerStruct(r *strings.Reader) ([]any, *ParserError) {
	tokens := []any{}
	decos := []DecoratorToken{}

	parserPos := getReaderPos(r)
	for {
		// Read the next Unicode character.
		c, s, err := r.ReadRune()
		if err != nil {
			// End of file before closing bracket.
			return nil, &ParserError{
				Message:  "unexpected end of file before closing bracket",
				Position: parserPos,
			}
		}

		// Switch on the character.
		switch c {
		case ' ', '\t', '\n', '\r':
			// Ignore whitespace.
		case '/':
			// Parse a comment.
			err := parseComment(r, &tokens)
			if err != nil {
				return nil, err
			}
		case '@':
			// Parse a decorator.
			deco, err := parseDecorator(r)
			if err != nil {
				return nil, err
			}
			decos = append(decos, deco)
		case '}':
			// Handle if there's unhandled decorators.
			if len(decos) > 0 {
				return nil, &ParserError{
					Message:  "unexpected decorators pointing to nothing",
					Position: parserPos,
				}
			}

			// Return the tokens.
			return tokens, nil
		default:
			// Parse the field or reference.
			token, err := parseStructFieldOrReference(r, c, decos)
			if err != nil {
				return nil, err
			}
			decos = []DecoratorToken{}
			tokens = append(tokens, token)
		}

		// Update the parser position.
		parserPos += s
	}
}

// Parses a struct. This assumes the s has already been read.
func parseStruct(r *strings.Reader, decorators []DecoratorToken) (StructToken, *ParserError) {
	// The position of the struct with 1 subtracted because we already read the character.
	pos := getReaderPos(r) - 1

	// Make sure the next content is 'truct'.
	b := make([]byte, 5)
	_, err := r.Read(b)
	if err != nil {
		// End of file.
		return StructToken{}, &ParserError{
			Message:  "unexpected end of file after 's' character",
			Position: pos,
		}
	}
	content := string(b)
	if content != "truct" {
		return StructToken{}, &ParserError{
			Message:  "unexpected '" + content + "' after 's' character - did you mean struct?",
			Position: pos,
		}
	}

	// Expect a space or newline.
	hasSpace, perr := gulpWhitespace(r)
	if perr != nil {
		return StructToken{}, perr
	}
	if !hasSpace {
		return StructToken{}, &ParserError{
			Message:  "unexpected lack of a space after 'struct' keyword - did you forget a space?",
			Position: pos,
		}
	}

	// Get the name of the struct.
	name := ""
nameReader:
	for {
		// Read the next Unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			// End of file.
			return StructToken{}, &ParserError{
				Message:  "unexpected end of file after start of struct definition",
				Position: pos,
			}
		}

		// Switch on the character.
		switch c {
		case '\n', ' ', '\t':
			// End of the name if there's anything in it.
			if name != "" {
				break nameReader
			}
		case '\r':
			// Read the next Unicode character.
			c, _, err := r.ReadRune()
			if err != nil {
				// End of file.
				return StructToken{}, &ParserError{
					Message:  "unexpected end of file after '\\r'",
					Position: pos,
				}
			}
			if c == '\n' {
				// End of the name if there's anything in it.
				if name != "" {
					break nameReader
				}
			} else {
				// A \r alone is not a valid return character.
				return StructToken{}, &ParserError{
					Message:  "unexpected '" + string(c) + "' after '\\r'",
					Position: pos,
				}
			}
		default:
			// Add the character to the name.
			name += string(c)
		}
	}

	// Make sure the name is not empty.
	if name == "" {
		return StructToken{}, &ParserError{
			Message:  "unexpected end of file after 'struct' keyword - did you forget the struct name?",
			Position: pos,
		}
	}

	// Make sure the name starts with a a-z or A-Z.
	if !(name[0] >= 'a' && name[0] <= 'z') && !(name[0] >= 'A' && name[0] <= 'Z') {
		return StructToken{}, &ParserError{
			Message:  "unexpected '" + string(name[0]) + "' as first character of struct name",
			Position: pos,
		}
	}

	// Look for the opening bracket.
openingBracketFind:
	for {
		// Read the next Unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			// End of file.
			return StructToken{}, &ParserError{
				Message:  "unexpected end of file after struct name",
				Position: pos,
			}
		}

		// Switch on the character.
		switch c {
		case ' ', '\t', '\n', '\r':
			// Do nothing.
		case '{':
			// We found the opening bracket.
			break openingBracketFind
		default:
			// Unexpected character.
			return StructToken{}, &ParserError{
				Message:  "unexpected '" + string(c) + "' after struct name",
				Position: pos,
			}
		}
	}

	// Parse the inner struct.
	fields, perr := parseInnerStruct(r)
	if perr != nil {
		return StructToken{}, perr
	}
	return StructToken{
		Name:       name,
		Position:   pos,
		Decorators: decorators,
		Fields:     fields,
	}, nil
}

// Parses tokens at the document root.
func parseDocRootToken(r *strings.Reader, tokens *[]any, c rune) *ParserError {
	// Defines all of the decorators that should be applied to the next non-decorator token.
	decorators := []DecoratorToken{}

	// Defines if the next token should be wrapped in a extends token.
	extends := -1

parseStart:
	// Get the position of the token. It is with 1 subtracted because we already read the character.
	pos := getReaderPos(r) - 1

	// Switch on the next character.
	switch c {
	case '/':
		// Parse a comment.
		err := parseComment(r, tokens)
		if err != nil {
			return err
		}

		// Gulp the whitespace.
		gulpWhitespace(r)

		// Get the next thing and handle EOF.
		var errIface error
		c, _, errIface = r.ReadRune()
		if errIface != nil {
			// End of file.
			return nil
		}

		// Go to the start of the parse.
		goto parseStart
	case '@':
		// Parse a decorator.
		deco, err := parseDecorator(r)
		if err != nil {
			return err
		}
		decorators = append(decorators, deco)
		var errIface error
		c, _, errIface = r.ReadRune()
		if errIface != nil {
			// End of file.
			return &ParserError{
				Message:  unexpectedEofAfterDeco,
				Position: pos,
			}
		}
		goto parseStart
	case 'e':
		// Parse an extends.
		err := parseExtends(r)
		if err != nil {
			return err
		}

		// Gulp the whitespace.
		gulpWhitespace(r)

		// Get the next thing.
		var errIface error
		c, _, errIface = r.ReadRune()
		if errIface != nil {
			// End of file.
			return &ParserError{
				Message:  "unexpected end of file after 'extends' keyword",
				Position: pos,
			}
		}

		// Handle if we were already in extends mode.
		if extends != -1 {
			return &ParserError{
				Message:  "unexpected second 'extends' keyword",
				Position: pos,
			}
		}
		extends = pos

		// Go to the start of the loop.
		goto parseStart
	case 's':
		// Parse a struct.
		s, err := parseStruct(r, decorators)
		if err != nil {
			return err
		}
		if extends == -1 {
			*tokens = append(*tokens, s)
		} else {
			*tokens = append(*tokens, ExtendsToken{
				Token:    s,
				Position: extends,
			})
		}
		return nil
	case 'c':
		// Parse a contract.
		c, err := parseContract(r, decorators)
		if err != nil {
			return err
		}
		if extends != -1 {
			return &ParserError{
				Message:  "unexpected 'contract' keyword after 'extends' keyword",
				Position: pos,
			}
		}
		*tokens = append(*tokens, c)
		return nil
	case 'm':
		// Parse a mapping.
		m, err := parseMapping(r, decorators)
		if err != nil {
			return err
		}
		if extends != -1 {
			return &ParserError{
				Message:  "unexpected 'mapping' keyword after 'extends' keyword",
				Position: pos,
			}
		}
		*tokens = append(*tokens, m)
		return nil
	default:
		// Unexpected character.
		return &ParserError{
			Message:  "unexpected character '" + string(c) + "' was hit",
			Position: pos,
		}
	}
}

// Handles a inline if/else statement.
func handleInlineIfElse(r *strings.Reader, readToken any) (any, *ParserError) {
	// Gulp just spaces specifically.
	var consumedChar rune
	for {
		// Read the next Unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			// Return here.
			return nil, nil
		}

		// Handle if the character is not a space or tab.
		if c != ' ' && c != '\t' {
			// Set the consumed character.
			consumedChar = c

			// Break the loop.
			break
		}
	}
	pos := getReaderPos(r) - 1

	// Switch on the character.
	switch consumedChar {
	case 'i':
		// Read the next Unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			// Return an error.
			return nil, &ParserError{
				Message:  "unexpected end of file after 'i' in inline if/else statement",
				Position: getReaderPos(r),
			}
		}

		// Handle if the character is not 'f'.
		if c != 'f' {
			// Return an error.
			return nil, &ParserError{
				Message:  "unexpected character '" + string(c) + "' after 'i' in inline if/else statement",
				Position: getReaderPos(r),
			}
		}

		// Gulps the whitespace.
		whitespace, _ := gulpWhitespace(r)
		if !whitespace {
			// Return an error.
			return nil, &ParserError{
				Message:  "expected whitespace after 'if' in inline if/else statement",
				Position: getReaderPos(r),
			}
		}

		// Parse the condition.
		condition, perr := parseInnerContractTokenWithOpGrouping(r, '}')
		if perr != nil {
			// Return the error.
			return nil, perr
		}

		// Return the inline if token.
		return InlineIfToken{
			Condition: condition,
			Position:  pos,
			Token:     readToken,
		}, nil
	case 'u':
		// Read the next 5 bytes.
		b := make([]byte, 5)
		if _, err := r.Read(b); err != nil {
			// Return an error.
			return nil, &ParserError{
				Message:  "unexpected end of file after 'u' in inline if/else statement",
				Position: getReaderPos(r),
			}
		}

		// Handle if the bytes are not 'nless'.
		if string(b) != "nless" {
			// Return an error.
			return nil, &ParserError{
				Message:  "unexpected character '" + string(b) + "' after 'u' in inline if/else statement",
				Position: getReaderPos(r),
			}
		}

		// Gulps the whitespace.
		whitespace, _ := gulpWhitespace(r)
		if !whitespace {
			// Return an error.
			return nil, &ParserError{
				Message:  "expected whitespace after 'unless' in inline if/else statement",
				Position: getReaderPos(r),
			}
		}

		// Parse the condition.
		condition, perr := parseInnerContractTokenWithOpGrouping(r, '}')
		if perr != nil {
			// Return the error.
			return nil, perr
		}

		// Return the inline unless token.
		return InlineUnlessToken{
			Condition: condition,
			Position:  pos,
			Token:     readToken,
		}, nil
	default:
		// Unread the rune and return the read token.
		_ = r.UnreadRune()
		return readToken, nil
	}
}

// Parse is used to parse a string into an AST. The any is all the types in tokens.go.
func Parse(input string) ([]any, *ParserError) {
	r := strings.NewReader(input)
	tokens := []any{}

	for {
		// Read the next Unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			// End of file.
			return tokens, nil
		}

		switch c {
		case ' ', '\t', '\n', '\r':
			// Ignore whitespace.
		default:
			// Parse tokens.
			if perr := parseDocRootToken(r, &tokens, c); perr != nil {
				return nil, perr
			}
		}
	}
}
