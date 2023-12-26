// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package ast

import (
	"io"
	"strings"
)

// Handle any special tokens that start with 't'.
func parseContractTokenCasesStartingT(r *strings.Reader, eot rune) (any, *ParserError) {
	// Read ahead 4 bytes.
	pos := getReaderPos(r) - 1
	b := make([]byte, 4)
	_, err := r.Read(b)
	if err != nil {
		// Not really our business.
		return nil, nil
	}

	// Unread 1 byte.
	_, _ = r.Seek(-1, io.SeekCurrent)

	// Switch on the end byte.
	switch b[3] {
	case ',', '=', '/', ' ', '\t', '\n', '\r', byte(eot):
		// Check if the remaining bytes are 'rue'.
		if b[0] == 'r' && b[1] == 'u' && b[2] == 'e' {
			// Return the boolean literal.
			return BooleanLiteralToken{
				Value:    true,
				Position: pos,
			}, nil
		}

		// Fallthrough.
		fallthrough
	default:
		// Rewind the reader more.
		_, _ = r.Seek(-3, io.SeekCurrent)

		// Return nil.
		return nil, nil
	}
}

// Handle any special tokens that start with 'f'.
func parseContractTokenCasesStartingF(r *strings.Reader, eot rune) (any, *ParserError) {
	// Read ahead 5 bytes.
	pos := getReaderPos(r) - 1
	b := make([]byte, 5)

	// Read the bytes.
	_, err := r.Read(b)
	if err != nil {
		// Not really our business.
		return nil, nil
	}

	// Unread 1 byte.
	_, _ = r.Seek(-1, io.SeekCurrent)

	// Switch on the end byte.
	switch b[4] {
	case ',', '=', ' ', '/', '\t', '\n', '\r', byte(eot):
		// Check if the remaining bytes are 'alse'.
		if b[0] == 'a' && b[1] == 'l' && b[2] == 's' && b[3] == 'e' {
			// Return the boolean literal.
			return BooleanLiteralToken{
				Value:    false,
				Position: pos,
			}, nil
		}

		// Fallthrough.
		fallthrough
	default:
		// Rewind the reader more.
		_, _ = r.Seek(-4, io.SeekCurrent)

		// Return nil.
		return nil, nil
	}
}

// Handle any special tokens that start with 'n'.
func parseContractTokenCasesStartingN(r *strings.Reader, eot rune) (any, *ParserError) {
	// Read ahead 4 bytes.
	pos := getReaderPos(r) - 1
	b := make([]byte, 4)

	// Read the bytes.
	_, err := r.Read(b)
	if err != nil {
		// Not really our business.
		return nil, nil
	}

	// Unread 1 byte.
	_, _ = r.Seek(-1, io.SeekCurrent)

	// Switch on the end byte.
	switch b[3] {
	case ',', '=', ' ', '/', '\t', '\n', '\r', byte(eot):
		// Check if the remaining bytes are 'ull'.
		if b[0] == 'u' && b[1] == 'l' && b[2] == 'l' {
			// Return the null literal.
			return NullLiteralToken{
				Position: pos,
			}, nil
		}

		// Fallthrough.
		fallthrough
	default:
		// Rewind the reader more.
		_, _ = r.Seek(-3, io.SeekCurrent)

		// Return nil.
		return nil, nil
	}
}

// Handles checking for a chained call.
func handleChainedCall(r *strings.Reader) any {
	// Check if the next rune is a dot.
	c, _, err := r.ReadRune()
	if err != nil {
		// End of file.
		return nil
	}

	// If it's not a dot, rewind the rune and return.
	if c != '.' {
		_ = r.UnreadRune()
		return nil
	}

	// Grab the name.
	name := ""
	pos := getReaderPos(r)
	for {
		// Read the next Unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			// End of file.
			return nil
		}

		// Switch on the character.
		switch c {
		case ' ', '\t', '\n', '\r':
			// End of the name if there's anything in it.
			if name != "" {
				return ReferenceToken{
					Name:     name,
					Position: pos,
				}
			}
		case '(':
			// Handle if the name is blank.
			if name == "" {
				// This would be '.(' which is illegal. Rewind so
				// the next token can handle it.
				_ = r.UnreadRune()
				return nil
			}

			// Parse the method call.
			methodCall, err := parseMethodCall(r, getReaderPos(r), name)
			if err != nil {
				return nil
			}
			return methodCall
		default:
			// Add the character to the name.
			name += string(c)
		}
	}
}

// Parses a method call inside a contract.
func parseMethodCall(r *strings.Reader, pos int, name string) (any, *ParserError) {
	// Defines the arguments.
	args := []any{}

	// Consume all of the whitespace.
	gulpWhitespace(r)

	// Check if there's even any arguments.
	c, _, err := r.ReadRune()
	if err != nil {
		// End of file.
		return nil, &ParserError{
			Message:  "unexpected end of file after method call starting bracket",
			Position: getReaderPos(r),
		}
	}
	if c == ')' {
		// No arguments.
		return handleInlineIfElse(r, MethodCallToken{
			Name:        name,
			Position:    pos,
			Arguments:   args,
			ChainedCall: handleChainedCall(r),
		})
	}

	// Rewind the rune.
	_ = r.UnreadRune()

	// Loop to get the information we need.
	for {
		// Read the argument.
		arg, err := parseInnerContractTokenWithOpGrouping(r, ')')
		if err != nil {
			return MethodCallToken{}, err
		}

		// Add the argument.
		if arg == nil {
			break
		}
		args = append(args, arg)

		// Gulp all the whitespace.
		gulpWhitespace(r)

		// Read the next Unicode character.
		c, _, x := r.ReadRune()
		if x != nil {
			// End of file.
			return MethodCallToken{}, &ParserError{
				Message:  "unexpected end of file after method call argument",
				Position: getReaderPos(r),
			}
		}

		// If this is a closing bracket, return the method call.
		if c == ')' {
			return handleInlineIfElse(r, MethodCallToken{
				Name:        name,
				Position:    pos,
				Arguments:   args,
				ChainedCall: handleChainedCall(r),
			})
		}

		// If this isn't a comma, raise a error.
		if c != ',' {
			return MethodCallToken{}, &ParserError{
				Message:  "unexpected '" + string(c) + "' after method call argument",
				Position: getReaderPos(r),
			}
		}
	}

	// This is a syntax error because there's no closing bracket.
	return MethodCallToken{}, &ParserError{
		Message:  "unexpected end of file after method call",
		Position: getReaderPos(r),
	}
}

// Parses a contract reference or call, figuring out which one it is.
func parseContractReferenceOrCall(r *strings.Reader, eot, c rune) (any, *ParserError) {
	// Defines the name.
	name := ""

	// Defines the position.
	pos := getReaderPos(r) - 1

	for {
		// Switch on the character.
		switch c {
		case eot, ' ', '\t', '\n', '\r', ',':
			// Handle returning a reference token if the name is not blank.
			if name != "" {
				_ = r.UnreadRune()
				return ReferenceToken{
					Name:     name,
					Position: pos,
				}, nil
			}
		case '(':
			// Parse the method call.
			return parseMethodCall(r, pos, name)
		default:
			// Add the character to the name.
			name += string(c)
		}

		// Read the next Unicode character.
		var err error
		c, _, err = r.ReadRune()
		if err != nil {
			// End of file.
			return nil, &ParserError{
				Message:  "unexpected end of file after contract reference or call",
				Position: pos,
			}
		}
	}
}

// Parses a token which is inside a contract. eot is the end of the contract token or what
// this is inside (for example, a end bracket for a method call). Use parseInnerContractTokenWithOpGrouping
// if you want to parse a token that is inside a contract and uses operations.
func parseInnerContractToken(r *strings.Reader, eot rune) (any, *ParserError) {
	// Read the next Unicode character.
	c, _, err := r.ReadRune()
	if err != nil {
		// End of file before closing bracket.
		return nil, &ParserError{
			Message:  "unexpected end of file before closing bracket",
			Position: getReaderPos(r),
		}
	}

	// Switch on the character.
	switch c {
	case '\'':
		// Parse the single quoted string.
		return parseSingleQuotedString(r)
	case '"':
		// Parse the double quoted string.
		return parseDoubleQuotedString(r)
	case '[':
		// Parse the array literal.
		return parseArrayLiteral(r)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', '-', '+':
		// Rewind the rune.
		_ = r.UnreadRune()

		// Parse the number literal.
		return parseNumberLiteral(r, eot)
	case eot, ',':
		// Rewind the rune.
		_ = r.UnreadRune()

		// Return nil.
		return nil, nil
	case '{':
		// Parse the object literal.
		return parseObjectLiteral(r)
	case '/':
		// Parse the comment.
		a := []any{}
		e := parseComment(r, &a)
		if e != nil {
			return nil, e
		}
		return a[0], nil
	case 't':
		// Handle all special tokens that start with 't'.
		token, err := parseContractTokenCasesStartingT(r, eot)
		if err != nil {
			return nil, err
		}
		if token != nil {
			return token, nil
		}
	case 'f':
		// Handle all special tokens that start with 'f'.
		token, err := parseContractTokenCasesStartingF(r, eot)
		if err != nil {
			return nil, err
		}
		if token != nil {
			return token, nil
		}
	case 'n':
		// Handle all special tokens that start with 'n'.
		token, err := parseContractTokenCasesStartingN(r, eot)
		if err != nil {
			return nil, err
		}
		if token != nil {
			return token, nil
		}
	}

	// Make sure this is a-z or A-Z.
	if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') {
		return nil, &ParserError{
			Message:  "unexpected '" + string(c) + "' in contract token",
			Position: getReaderPos(r),
		}
	}

	// Handle all other characters.
	return parseContractReferenceOrCall(r, eot, c)
}

// Parses the maths or boolean operation inside a contract. Returns 0 if there's no math operation.
func parseMathOrBoolOp(r *strings.Reader) rune {
	// Gulps all of the whitespace.
	gulpWhitespace(r)

	// Read the next Unicode character.
	c, _, err := r.ReadRune()
	if err != nil {
		// Not really our business.
		return 0
	}

	// Switch on the character.
	switch c {
	case '/':
		// Look ahead for a slash.
		c2, _, err := r.ReadRune()
		if c2 == '/' {
			// Rewind twice.
			_, _ = r.Seek(-2, io.SeekCurrent)

			// Return nothing.
			return 0
		}

		// Rewind the reader.
		if err == nil {
			_ = r.UnreadRune()
		}

		// Return the slash.
		return c
	case '+', '-', '*', '%', '^':
		// We found a operator.
		return c
	case '<', '>':
		// Read ahead a rune.
		c2, _, err := r.ReadRune()
		if err != nil {
			// Return for this.
			return 0
		}

		// If c2 is not a =, rewind the rune and return the operator.
		if c2 != '=' {
			_ = r.UnreadRune()
			return c
		}

		// Return the operator.
		if c == '<' {
			return '≤'
		}
		return '≥'
	case '=', '!', '&', '|':
		// Read the next Unicode character.
		c2, _, err := r.ReadRune()
		if err != nil {
			// Return for this.
			return 0
		}

		// Set the comparison character to = if it is a !.
		ogC := c
		if c == '!' {
			c = '='
		}

		// If c isn't equal to c2, return 0 and rewind the last couple of runes.
		if c != c2 {
			_, _ = r.Seek(-2, io.SeekCurrent)
			return 0
		}

		// Return the operator.
		return ogC
	default:
		// Rewind the rune.
		_ = r.UnreadRune()

		// Return nothing.
		return 0
	}
}

// Ensure the next thing is a bracket.
func ensureClosingBracket(r *strings.Reader) *ParserError {
	// Gulps all of the whitespace.
	gulpWhitespace(r)

	// Read the next Unicode character.
	c, _, err := r.ReadRune()
	if err != nil {
		// End of file before closing bracket.
		return &ParserError{
			Message:  "unexpected end of file before closing bracket",
			Position: getReaderPos(r),
		}
	}

	// Make sure this is a closing bracket.
	if c != ')' {
		return &ParserError{
			Message:  "unexpected '" + string(c) + "' after contract token",
			Position: getReaderPos(r),
		}
	}

	// Return nil.
	return nil
}

// Parses a token and looks ahead for a group that should group with the token.
func parseInnerContractTokenWithOpGrouping(r *strings.Reader, eot rune) (any, *ParserError) {
	// Consume all of the whitespace.
	gulpWhitespace(r)

	// Read the next Unicode character.
	c, _, err := r.ReadRune()
	if err != nil {
		// End of file before closing bracket.
		return nil, &ParserError{
			Message:  "unexpected end of file before closing bracket",
			Position: getReaderPos(r),
		}
	}

	// Handle the result.
	var res any
	var perr *ParserError
	switch c {
	case '(':
		// Call ourselves with a end bracket.
		res, perr = parseInnerContractTokenWithOpGrouping(r, ')')

		if res == nil && perr == nil {
			// This is a special case where we have a empty bracket.
			return nil, &ParserError{
				Message:  "unexpected empty bracket",
				Position: getReaderPos(r),
			}
		}

		// Check for the closing bracket.
		if perr := ensureClosingBracket(r); perr != nil {
			return nil, perr
		}
	case '!':
		// Get the position of this bang.
		pos := getReaderPos(r) - 1

		// Check what the next rune is.
		ru, _, err := r.ReadRune()

		if ru == '(' {
			// Parse with op grouping until the end bracket.
			res, perr = parseInnerContractTokenWithOpGrouping(r, ')')

			// Check for the closing bracket.
			if perr := ensureClosingBracket(r); perr != nil {
				return nil, perr
			}
		} else {
			// Unread the rune.
			if err == nil {
				_ = r.UnreadRune()
			}

			// Parse the inner token without any op grouping.
			res, perr = parseInnerContractToken(r, eot)
		}

		if res != nil {
			// Wrap the token in a not token.
			res = NotToken{
				Token:    res,
				Position: pos,
			}
		}
	default:
		// Rewind the rune.
		_ = r.UnreadRune()

		// Parse the token.
		res, perr = parseInnerContractToken(r, eot)
	}

	// Return if there's a error.
	if perr != nil {
		return nil, perr
	}

	// If we didn't pick up a token, return nil.
	if res == nil {
		return nil, nil
	}

	// Look ahead for a operator.
	pos := getReaderPos(r)
	op := parseMathOrBoolOp(r)
	if op == 0 {
		// No operator.
		return res, nil
	}

	// Parse the next token.
	next, perr := parseInnerContractTokenWithOpGrouping(r, eot)
	if perr != nil {
		return nil, perr
	}

	// Return the math operation.
	switch op {
	case '+':
		return AddToken{
			Left:     res,
			Right:    next,
			Position: pos,
		}, nil
	case '-':
		return SubtractToken{
			Left:     res,
			Right:    next,
			Position: pos,
		}, nil
	case '*':
		return MultiplyToken{
			Left:     res,
			Right:    next,
			Position: pos,
		}, nil
	case '/':
		return DivideToken{
			Left:     res,
			Right:    next,
			Position: pos,
		}, nil
	case '%':
		return ModuloToken{
			Left:     res,
			Right:    next,
			Position: pos,
		}, nil
	case '^':
		return ExponentToken{
			Left:     res,
			Right:    next,
			Position: pos,
		}, nil
	case '<':
		return LessThanToken{
			Left:     res,
			Right:    next,
			Position: pos,
		}, nil
	case '>':
		return GreaterThanToken{
			Left:     res,
			Right:    next,
			Position: pos,
		}, nil
	case '=':
		return EqualToken{
			Left:     res,
			Right:    next,
			Position: pos,
		}, nil
	case '!':
		return NotEqualToken{
			Left:     res,
			Right:    next,
			Position: pos,
		}, nil
	case '&':
		return AndToken{
			Left:     res,
			Right:    next,
			Position: pos,
		}, nil
	case '|':
		return OrToken{
			Left:     res,
			Right:    next,
			Position: pos,
		}, nil
	case '≤':
		return LessThanOrEqualToken{
			Left:     res,
			Right:    next,
			Position: pos,
		}, nil
	case '≥':
		return GreaterThanOrEqualToken{
			Left:     res,
			Right:    next,
			Position: pos,
		}, nil
	default:
		// This should never happen.
		panic("unexpected operator value")
	}
}

// Parses a contract argument.
func parseContractArgument(r *strings.Reader) (*ContractArgumentToken, *ParserError) {
	// Get the next Unicode character.
	c, _, err := r.ReadRune()
	if err != nil {
		// End of file.
		return nil, &ParserError{
			Message:  "unexpected end of file after contract argument bracket",
			Position: getReaderPos(r),
		}
	}

	// Check if there's even any arguments.
	if c == ')' {
		// No arguments.
		return nil, nil
	}

	// Defines the argument name and type.
	name := ""
	nameIndex := getReaderPos(r)
	type_ := ""
	typeState := false
	typeIndex := -1

	// Loop to get the information we need.
	for {
		// Switch on the character.
		switch c {
		case ':':
			// Make sure we are not already in type mode.
			if typeState {
				return nil, &ParserError{
					Message:  "unexpected ':' after ':'",
					Position: getReaderPos(r),
				}
			}

			// Handle if no name is present.
			if name == "" {
				return nil, &ParserError{
					Message:  "unexpected ':' after '('",
					Position: getReaderPos(r),
				}
			}

			// We are now in type mode.
			typeState = true
			typeIndex = getReaderPos(r)
		case ' ', '\t', '\n', '\r':
			// Check if the mode we are in has content.
			hasContent := name != ""
			if typeState {
				hasContent = type_ != ""
			}

			// If there's content, this is in the middle of the argument.
			if hasContent {
				return nil, &ParserError{
					Message:  "unexpected whitespace in the middle of an argument",
					Position: getReaderPos(r),
				}
			}
		case ')':
			// Handle if no name is present.
			if name == "" {
				return nil, &ParserError{
					Message:  "unexpected ')' after '('",
					Position: getReaderPos(r),
				}
			}

			// Handle if no type is present.
			if typeState && type_ == "" {
				return nil, &ParserError{
					Message:  "unexpected ')' after ':'",
					Position: getReaderPos(r),
				}
			}

			// Handle if the colon for the type is missing.
			if !typeState {
				return nil, &ParserError{
					Message:  "unexpected ')' after argument name - did you forget the type?",
					Position: getReaderPos(r),
				}
			}

			// Return the token.
			return &ContractArgumentToken{
				Name:      name,
				NameIndex: nameIndex,
				Type:      type_,
				TypeIndex: typeIndex,
			}, nil
		default:
			// Add the character to the name or type.
			if typeState {
				type_ += string(c)
			} else {
				name += string(c)
			}
		}

		// Read the next Unicode character.
		c, _, err = r.ReadRune()
		if err != nil {
			// End of file.
			return nil, &ParserError{
				Message:  "unexpected end of file during contract argument",
				Position: getReaderPos(r),
			}
		}
	}
}

// Parses a contract return type.
func parseContractReturnType(r *strings.Reader) (string, *ParserError) {
	state := 0
	content := ""
	for {
		c, _, err := r.ReadRune()
		if err != nil {
			// End of file.
			return "", &ParserError{
				Message:  "unexpected end of file after contract return type bracket",
				Position: getReaderPos(r),
			}
		}

		switch c {
		case '-':
			if state != 0 {
				return "", &ParserError{
					Message:  "unexpected '-' after '-'",
					Position: getReaderPos(r),
				}
			}
			state = 1
		case '>':
			if state != 1 {
				return "", &ParserError{
					Message:  "unexpected '>' that is not preceded by '-'",
					Position: getReaderPos(r),
				}
			}
			state = 2
		case ' ', '\t', '\n', '\r':
			if state == 1 {
				return "", &ParserError{
					Message:  "unexpected whitespace after '-'",
					Position: getReaderPos(r),
				}
			}

			// Return when we are in state 2 unless its a blank string.
			if state == 2 && content != "" {
				// Rewind the rune.
				_ = r.UnreadRune()

				return content, nil
			}
		default:
			if state == 2 {
				content += string(c)
			} else {
				// Unexpected character.
				return "", &ParserError{
					Message:  "unexpected '" + string(c) + "' after contract return type",
					Position: getReaderPos(r),
				}
			}
		}
	}
}

// Parses the inners of a contract throws.
func parseInnerContractThrows(r *strings.Reader) ([]ContractThrowsToken, *ParserError) {
	// Defines all of the throw tokens.
	throws := []ContractThrowsToken{}

	// Defines the loop to get each throw token.
	for {
		// Defines the position of the throw token.
		pos := -1

		// Defines the buffer holding the throw token.
		buffer := ""

		// Defines the method to drain the buffer.
		drainBuffer := func() {
			// Drain the buffer.
			if buffer != "" {
				throws = append(throws, ContractThrowsToken{
					Name:     buffer,
					Position: pos,
				})
			}
		}

		// Gulp the whitespace.
		gulpWhitespace(r)

		// Defines the loop to consume the throw token.
	nameLoop:
		for {
			// Read the next Unicode character.
			c, _, err := r.ReadRune()
			if err != nil {
				// End of file.
				return nil, &ParserError{
					Message:  "unexpected end of file after start of throws token",
					Position: pos,
				}
			}

			// Switch on the character.
			switch c {
			case '{':
				// This is the end of the throw token. Rewind and return
				// since we are done.
				_ = r.UnreadRune()
				drainBuffer()
				return throws, nil
			case ' ', '\t', '\n', '\r':
				// If the buffer is blank, break.
				if buffer == "" {
					break
				}

				// Drain the buffer.
				drainBuffer()

				// Gulp the whitespace.
				gulpWhitespace(r)

				// Read the next Unicode character.
				c, _, err := r.ReadRune()
				if err != nil {
					// End of file.
					return nil, &ParserError{
						Message:  "unexpected end of file after start of throws token",
						Position: pos,
					}
				}

				// Switch on the character.
				switch c {
				case '{':
					// Rewind the rune and return.
					_ = r.UnreadRune()
					return throws, nil
				case ',':
					// Break the name loop.
					break nameLoop
				default:
					// Unexpected character.
					return nil, &ParserError{
						Message:  "unexpected '" + string(c) + "' after space in throws token",
						Position: getReaderPos(r),
					}
				}
			case ',':
				// Drain the buffer.
				drainBuffer()

				// Break the name loop.
				break nameLoop
			default:
				// If buffer is blank, set the position.
				if buffer == "" {
					pos = getReaderPos(r) - 1
				}

				// Add the character to the buffer.
				buffer += string(c)
			}
		}
	}
}

// Parses what a contract throws.
func parseContractThrows(r *strings.Reader) ([]ContractThrowsToken, *ParserError) {
	// Gulp the whitespace.
	whitespace, _ := gulpWhitespace(r)
	if !whitespace {
		// Cannot be 'throws'.
		return []ContractThrowsToken{}, nil
	}

	// Read the next Unicode character.
	c, _, err := r.ReadRune()
	if err != nil {
		// End of file.
		return nil, &ParserError{
			Message:  "unexpected end of file after contract return",
			Position: getReaderPos(r),
		}
	}

	// Make sure this is 't'.
	if c != 't' {
		// Rewind the rune and return.
		_ = r.UnreadRune()
		return []ContractThrowsToken{}, nil
	}

	// Check the next bit is 'hrows'.
	b := make([]byte, 5)
	_, err = r.Read(b)
	if err != nil {
		// End of file.
		return nil, &ParserError{
			Message:  "unexpected end of file after start of throws token",
			Position: getReaderPos(r),
		}
	}

	// Make sure this is 'hrows'.
	if string(b) != "hrows" {
		// If not, this is just an unexpected 't'.
		return nil, &ParserError{
			Message:  "unexpected 't' after contract return",
			Position: getReaderPos(r),
		}
	}

	// Gulp the whitespace.
	whitespace, _ = gulpWhitespace(r)
	if !whitespace {
		// Throws needs to be followed by a space.
		return nil, &ParserError{
			Message:  "unexpected lack of whitespace after 'throws' keyword",
			Position: getReaderPos(r),
		}
	}

	// Parse the inner contract throws.
	return parseInnerContractThrows(r)
}

// Parses a contract. The first c should be consumed.
func parseContract(r *strings.Reader, decorators []DecoratorToken) (ContractToken, *ParserError) {
	// Get the position of the start of the contract.
	pos := getReaderPos(r) - 1

	// Make sure the next content is 'ontract'.
	b := make([]byte, 7)
	_, err := r.Read(b)
	if err != nil {
		// End of file.
		return ContractToken{}, &ParserError{
			Message:  "unexpected end of file after 'c' character",
			Position: pos,
		}
	}
	content := string(b)
	if content != "ontract" {
		return ContractToken{}, &ParserError{
			Message:  "unexpected '" + content + "' after 'c' character - did you mean contract?",
			Position: pos,
		}
	}

	// Expect a space or newline.
	whitespace, perr := gulpWhitespace(r)
	if perr != nil {
		return ContractToken{}, perr
	}
	if !whitespace {
		return ContractToken{}, &ParserError{
			Message:  "unexpected non-space after 'contract' keyword - did you forget a space?",
			Position: pos,
		}
	}

	// Get the name of the contract.
	name := ""
nameReader:
	for {
		// Read the next Unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			// End of file.
			return ContractToken{}, &ParserError{
				Message:  "unexpected end of file after start of contract definition",
				Position: pos,
			}
		}

		// Switch on the character.
		switch c {
		case '(':
			// If the contract name is blank, this is illegal.
			if name == "" {
				return ContractToken{}, &ParserError{
					Message:  "unexpected '(' after 'contract' keyword - did you forget the contract name?",
					Position: pos,
				}
			}

			// We found the opening bracket.
			break nameReader
		case ' ', '\t':
			// Ignore whitespace.
		case '\n', '\r':
			// Error since arguments cannot be on a new line.
			return ContractToken{}, &ParserError{
				Message:  "unexpected line ending after 'contract' keyword - did you put a new line before the arguments?",
				Position: pos,
			}
		default:
			// Add the character to the name.
			name += string(c)
		}
	}

	// Make sure the name starts with a a-z or A-Z.
	if !(name[0] >= 'a' && name[0] <= 'z') && !(name[0] >= 'A' && name[0] <= 'Z') {
		return ContractToken{}, &ParserError{
			Message:  "unexpected '" + string(name[0]) + "' as first character of contract name",
			Position: pos,
		}
	}

	// Get the argument of the contract.
	arg, perr := parseContractArgument(r)
	if perr != nil {
		return ContractToken{}, perr
	}

	// Parse the return type of the contract.
	returnType, perr := parseContractReturnType(r)
	if perr != nil {
		return ContractToken{}, perr
	}

	// Parses anything this throws.
	throws, perr := parseContractThrows(r)
	if perr != nil {
		return ContractToken{}, perr
	}

	// Look for the opening bracket.
openingBracketFind:
	for {
		// Read the next Unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			// End of file.
			return ContractToken{}, &ParserError{
				Message:  "unexpected end of file after contract name",
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
			return ContractToken{}, &ParserError{
				Message:  "unexpected '" + string(c) + "' after contract definition",
				Position: pos,
			}
		}
	}

	// Parse the inner contract.
	inner, perr := parseInnerContract(r)
	if perr != nil {
		// This means the inside of the contract was invalid.
		return ContractToken{}, perr
	}

	// Return the contract.
	return ContractToken{
		Name:       name,
		Position:   pos,
		ReturnType: returnType,
		Argument:   arg,
		Throws:     throws,
		Decorators: decorators,
		Statements: inner,
	}, nil
}
