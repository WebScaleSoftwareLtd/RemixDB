// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package ast

import (
	"io"
	"strings"
)

type parserHn = func(*strings.Reader) ([]any, *ParserError)

// Handle parsing a else statement.
func parsePotentialElse(r *strings.Reader, parser parserHn) (*ElseToken, *ParserError) {
	var first *ElseToken
	var last *ElseToken
	for {
		// Gulp the whitespace.
		gulpWhitespace(r)

		// Get the position.
		pos := getReaderPos(r)

		// Read the next Unicode character.
		c, _, errIface := r.ReadRune()
		if errIface != nil {
			// Return first. Let the parent handle the error.
			return first, nil
		}

		// Handle if the character is not 'e'.
		if c != 'e' {
			// Unread the rune.
			_ = r.UnreadRune()

			// Return first.
			return first, nil
		}

		// Read the next 3 bytes.
		b := make([]byte, 3)
		if _, err := r.Read(b); err != nil {
			// Un-read the byte and return first.
			_, _ = r.Seek(-1, io.SeekCurrent)
			return first, nil
		}

		// Handle if the bytes are not 'lse' or 'lif'.
		if string(b) != "lse" && string(b) != "lif" {
			// Un-read the bytes and return first.
			_, _ = r.Seek(-4, io.SeekCurrent)
			return first, nil
		}

		// Gulp the whitespace.
		gulpWhitespace(r)

		// Read the condition.
		var condition any
		var perr *ParserError
		if string(b) != "lse" {
			condition, perr = parseInnerContractTokenWithOpGrouping(r, '{')
			if perr != nil {
				// Return the error.
				return nil, perr
			}
		}

		// Gulp the whitespace.
		gulpWhitespace(r)

		// Read the next Unicode character.
		c, _, errIface = r.ReadRune()
		if errIface != nil {
			// Return an error.
			return nil, &ParserError{
				Message:  "unexpected end of file in else statement",
				Position: getReaderPos(r),
			}
		}

		// Handle if the character is not '{'.
		if c != '{' {
			// Return an error.
			return nil, &ParserError{
				Message:  "unexpected character '" + string(c) + "' in else statement",
				Position: getReaderPos(r),
			}
		}

		// Parse the body.
		body, perr := parser(r)
		if perr != nil {
			return nil, perr
		}

		// Create the else token.
		prevLast := last
		last = &ElseToken{
			Condition:  condition,
			Position:   pos,
			Statements: body,
		}
		if prevLast == nil {
			first = last
		} else {
			prevLast.Next = last
		}
	}
}

// Handle parsing a unless statement.
func unlessParser(r *strings.Reader, pos int, parser parserHn) (any, *ParserError) {
	// Parse the token.
	token, err := parseInnerContractTokenWithOpGrouping(r, '{')
	if err != nil {
		return nil, err
	}

	// Gulp the whitespace.
	gulpWhitespace(r)

	// Read the next Unicode character.
	c, _, errIface := r.ReadRune()
	if errIface != nil {
		// Return an error.
		return nil, &ParserError{
			Message:  "unexpected end of file after 'unless' in unless statement",
			Position: getReaderPos(r),
		}
	}

	// Handle if the character is not '{'.
	if c != '{' {
		// Return an error.
		return nil, &ParserError{
			Message:  "unexpected character '" + string(c) + "' after 'unless' in unless statement",
			Position: getReaderPos(r),
		}
	}

	// Parse the body.
	body, err := parser(r)
	if err != nil {
		return nil, err
	}

	// Return the unless token.
	else_, err := parsePotentialElse(r, parser)
	if err != nil {
		return nil, err
	}
	return UnlessToken{
		Condition:  token,
		Position:   pos,
		Statements: body,
		Else:       else_,
	}, nil
}

// Handle parsing a if statement.
func ifParser(r *strings.Reader, pos int, parser parserHn) (any, *ParserError) {
	// Parse the token.
	token, err := parseInnerContractTokenWithOpGrouping(r, '{')
	if err != nil {
		return nil, err
	}

	// Gulp the whitespace.
	gulpWhitespace(r)

	// Read the next Unicode character.
	c, _, errIface := r.ReadRune()
	if errIface != nil {
		// Return an error.
		return nil, &ParserError{
			Message:  "unexpected end of file after 'if' in if statement",
			Position: getReaderPos(r),
		}
	}

	// Handle if the character is not '{'.
	if c != '{' {
		// Return an error.
		return nil, &ParserError{
			Message:  "unexpected character '" + string(c) + "' after 'if' in if statement",
			Position: getReaderPos(r),
		}
	}

	// Parse the body.
	body, err := parser(r)
	if err != nil {
		return nil, err
	}

	// Return the if token.
	else_, err := parsePotentialElse(r, parser)
	if err != nil {
		return nil, err
	}
	return IfToken{
		Condition:  token,
		Position:   pos,
		Statements: body,
		Else:       else_,
	}, nil
}

// Handle parsing a return statement.
func returnParser(r *strings.Reader, pos int, _ parserHn) (any, *ParserError) {
	// Parse the token.
	token, err := parseInnerContractTokenWithOpGrouping(r, '}')
	if err != nil {
		return nil, err
	}

	// Wrap in a return token.
	token = ReturnToken{
		Token:    token,
		Position: pos,
	}

	// Return with the inline if/else handler.
	return handleInlineIfElse(r, token)
}

// Handle parsing a throw statement.
func throwParser(r *strings.Reader, pos int, _ parserHn) (any, *ParserError) {
	// Parse the token.
	token, err := parseInnerContractTokenWithOpGrouping(r, '}')
	if err != nil {
		return nil, err
	}

	// Wrap in a throw token.
	token = ThrowLiteralToken{
		Token:    token,
		Position: pos,
	}

	// Return with the inline if/else handler.
	return handleInlineIfElse(r, token)
}

// Handle parsing a while statement.
func whileParser(r *strings.Reader, pos int, parser parserHn) (any, *ParserError) {
	// Parse the token.
	token, err := parseInnerContractTokenWithOpGrouping(r, '{')
	if err != nil {
		return nil, err
	}

	// Gulp the whitespace.
	gulpWhitespace(r)

	// Read the next Unicode character.
	c, _, errIface := r.ReadRune()
	if errIface != nil {
		// Return an error.
		return nil, &ParserError{
			Message:  "unexpected end of file after 'while' in while statement",
			Position: getReaderPos(r),
		}
	}

	// Handle if the character is not '{'.
	if c != '{' {
		// Return an error.
		return nil, &ParserError{
			Message:  "unexpected character '" + string(c) + "' after 'while' in while statement",
			Position: getReaderPos(r),
		}
	}

	// Parse the body.
	body, err := parser(r)
	if err != nil {
		return nil, err
	}

	// Return the while token.
	return WhileToken{
		Condition:  token,
		Position:   pos,
		Statements: body,
	}, nil
}

// Make sure the semi-colon is gobbled.
func gulpSemiColon(r *strings.Reader) {
	// Gulp the whitespace.
	gulpWhitespace(r)

	// Read the next Unicode character.
	c, _, _ := r.ReadRune()
	if c != ';' {
		// Rewind the rune.
		_ = r.UnreadRune()
	}
}

// Handle parsing a for statement.
func forParser(r *strings.Reader, pos int, parser parserHn) (any, *ParserError) {
	// Parse the initial assignment.
	initial, err := potentialAssignmentRef(r, ';', false)
	if err != nil {
		return nil, err
	}
	gulpSemiColon(r)

	// Parse the condition.
	condition, err := potentialAssignmentRef(r, ';', false)
	if err != nil {
		return nil, err
	}
	gulpSemiColon(r)

	// Parse the increment.
	increment, err := potentialAssignmentRef(r, '{', false)
	if err != nil {
		return nil, err
	}
	gulpSemiColon(r)

	// Gulp the whitespace.
	gulpWhitespace(r)

	// Read the next Unicode character.
	c, _, errIface := r.ReadRune()
	if errIface != nil {
		// Return an error.
		return nil, &ParserError{
			Message:  "unexpected end of file after 'for' in for statement",
			Position: getReaderPos(r),
		}
	}

	// Handle if the character is not '{'.
	if c != '{' {
		// Return an error.
		return nil, &ParserError{
			Message:  "unexpected character '" + string(c) + "' after 'for' in for statement",
			Position: getReaderPos(r),
		}
	}

	// Parse the body.
	body, err := parser(r)
	if err != nil {
		return nil, err
	}

	// Return the for token.
	return ForToken{
		Assignment: initial,
		Condition:  condition,
		Increment:  increment,
		Position:   pos,
		Statements: body,
	}, nil
}

type switchCasesResult struct {
	cases    []SwitchCaseToken
	comments []CommentToken
}

// Handle parsing switch cases.
func switchCasesParser(r *strings.Reader, parser parserHn) (switchCasesResult, *ParserError) {
	// Defines all of the token cases.
	tokens := []SwitchCaseToken{}

	// Defines all of the comment tokens.
	comments := []CommentToken{}

	// Gulp the whitespace.
	gulpWhitespace(r)

	for {
		// Check if the next character is '}'. This means that this is an empty switch.
		c, _, _ := r.ReadRune()
		if c == '}' {
			// Return here.
			return switchCasesResult{}, nil
		}
		_ = r.UnreadRune()

		// Get the position.
		pos := getReaderPos(r)

		// Parse the name.
		name, err := parseInnerContractTokenWithOpGrouping(r, '=')
		if err != nil {
			return switchCasesResult{}, err
		}

		// Handle if the name is a comment token.
		if x, ok := name.(CommentToken); ok {
			// Add the comment token.
			comments = append(comments, x)

			// Continue the loop.
			continue
		}

		// Handle if there is more content present.
		if name != nil {
		eqRead:
			// Gulp the whitespace.
			gulpWhitespace(r)

			// Read the next Unicode character.
			ru, _, errIface := r.ReadRune()
			if errIface != nil {
				// Return an error.
				return switchCasesResult{}, &ParserError{
					Message:  "unexpected end of file after name in switch statement",
					Position: getReaderPos(r),
				}
			}

			// Handle if the character is not '='.
			if ru != '=' {
				// If it is a slash, consume the comment and then go back.
				if ru == '/' {
					var x []any
					if err := parseComment(r, &x); err != nil {
						return switchCasesResult{}, err
					}
					comments = append(comments, x[0].(CommentToken))
					goto eqRead
				}

				// Return an error.
				return switchCasesResult{}, &ParserError{
					Message:  "unexpected character '" + string(ru) + "' after name in switch statement",
					Position: getReaderPos(r),
				}
			}

		valueRead:
			// Gulp the whitespace.
			gulpWhitespace(r)

			// Get the first byte of the next token.
			var inner []any
			ru, _, _ = r.ReadRune()
			if ru == '{' {
				// Inline statements.
				inner, err = parser(r)
				if err != nil {
					return switchCasesResult{}, err
				}
			} else {
				// Unread the rune.
				_ = r.UnreadRune()

				// Parse the token.
				x, err := parseInnerContractTokenWithOpGrouping(r, '}')
				if err != nil {
					return switchCasesResult{}, err
				}
				if x != nil {
					// Handle any comments.
					if x, ok := x.(CommentToken); ok {
						// Add the comment token.
						comments = append(comments, x)

						// Go back to the start of parsing the value.
						goto valueRead
					}

					// Set the inner to only this object.
					inner = []any{x}
				}
			}

			// Push the token.
			tokens = append(tokens, SwitchCaseToken{
				Position:   pos,
				Name:       name,
				Statements: inner,
			})
		}

		// Gulp the whitespace.
		gulpWhitespace(r)

		// Read the next Unicode character.
		c, _, errIface := r.ReadRune()
		if errIface != nil {
			// Return an error.
			return switchCasesResult{}, &ParserError{
				Message:  "unexpected end of file in switch statement",
				Position: getReaderPos(r),
			}
		}

		switch c {
		case '}':
			// Handle if the character is '}'.
			return switchCasesResult{
				cases:    tokens,
				comments: comments,
			}, nil
		default:
			// Rewind the rune.
			_ = r.UnreadRune()
		}
	}
}

// Handle parsing a switch statement.
func switchParser(r *strings.Reader, pos int, parser parserHn) (any, *ParserError) {
	// Parse the token.
	token, err := parseInnerContractTokenWithOpGrouping(r, '{')
	if err != nil {
		return nil, err
	}

	// Gulp the whitespace.
	gulpWhitespace(r)

	// Read the next Unicode character.
	c, _, errIface := r.ReadRune()
	if errIface != nil {
		// Return an error.
		return nil, &ParserError{
			Message:  "unexpected end of file after 'switch' in switch statement",
			Position: getReaderPos(r),
		}
	}

	// Handle if the character is not '{'.
	if c != '{' {
		// Return an error.
		return nil, &ParserError{
			Message:  "unexpected character '" + string(c) + "' after 'switch' in switch statement",
			Position: getReaderPos(r),
		}
	}

	// Handle switch cases.
	res, err := switchCasesParser(r, parser)
	if err != nil {
		return nil, err
	}

	// Return the switch token.
	return SwitchToken{
		Condition: token,
		Position:  pos,
		Cases:     res.cases,
		Comments:  res.comments,
	}, nil
}

// Handle parsing catch statements.
func catchesParser(r *strings.Reader, pos int, parser parserHn) (*CatchToken, *ParserError) {
	var first *CatchToken
	var last *CatchToken

	for {
		// Gulp the whitespace.
		gulpWhitespace(r)

		// Read the next Unicode character.
		c, _, errIface := r.ReadRune()
		if errIface != nil {
			// Return first. Let the parent handle the error.
			return first, nil
		}

		// Handle if the character is not 'c'.
		if c != 'c' {
			// Unread the rune.
			_ = r.UnreadRune()

			// Return first.
			return first, nil
		}

		// Read the next 4 bytes.
		b := make([]byte, 4)
		if _, err := r.Read(b); err != nil {
			// Un-read the bytes and return first.
			_, _ = r.Seek(-4, io.SeekCurrent)
			return first, nil
		}

		// Handle if the bytes are not 'atch'.
		if string(b) != "atch" {
			// Un-read the bytes and return first.
			_, _ = r.Seek(-4, io.SeekCurrent)
			return first, nil
		}

		// Gulp the whitespace.
		gulpWhitespace(r)

		// Read the exception.
		exception := ""
		for {
			// Read the next Unicode character.
			c, _, err := r.ReadRune()
			if err != nil {
				// Return an error.
				return nil, &ParserError{
					Message:  "unexpected end of file after 'catch' in catch statement",
					Position: getReaderPos(r),
				}
			}

			// Handle if the character is ' ', '{' or '-'.
			if c == '{' || c == '-' || c == ' ' {
				// Unread the rune.
				_ = r.UnreadRune()

				// Break the loop.
				break
			}

			// Add the character to the exception.
			exception += string(c)
		}

		// Gulp the whitespace.
		gulpWhitespace(r)

		// Read the next Unicode character.
		c, _, errIface = r.ReadRune()
		if errIface != nil {
			// Return an error.
			return nil, &ParserError{
				Message:  "unexpected end of file after 'catch' in catch statement",
				Position: getReaderPos(r),
			}
		}

		// Defines the variable name.
		variable := ""

		// Switch on the character.
		switch c {
		case '-':
			// Read the next Unicode character.
			c, _, err := r.ReadRune()
			if err != nil {
				// Return an error.
				return nil, &ParserError{
					Message:  "unexpected end of file after name in catch statement",
					Position: getReaderPos(r),
				}
			}

			// Handle if the character is not '>'.
			if c != '>' {
				// Return an error.
				return nil, &ParserError{
					Message:  "unexpected character '" + string(c) + "' after name in catch statement",
					Position: getReaderPos(r),
				}
			}

			// Gulp the whitespace.
			gulpWhitespace(r)

			// Read the variable name.
			for {
				// Read the next Unicode character.
				c, _, err := r.ReadRune()
				if err != nil {
					// Return an error.
					return nil, &ParserError{
						Message:  "unexpected end of file after name in catch statement",
						Position: getReaderPos(r),
					}
				}

				// Handle if the character is a end of variable.
				if c == '{' || c == ' ' || c == '\t' || c == '\n' || c == '\r' {
					// Unread the rune.
					_ = r.UnreadRune()

					// Break the loop.
					break
				}

				// Add the character to the variable.
				variable += string(c)
			}

			// Gulp the whitespace.
			gulpWhitespace(r)

			// Read the next Unicode character.
			c, _, errIface = r.ReadRune()
			if errIface != nil {
				// Return an error.
				return nil, &ParserError{
					Message:  "unexpected end of file after variable in catch statement",
					Position: getReaderPos(r),
				}
			}

			// Handle if the character is not '{'.
			if c != '{' {
				// Return an error.
				return nil, &ParserError{
					Message:  "unexpected character '" + string(c) + "' after variable in catch statement",
					Position: getReaderPos(r),
				}
			}

			// Fallthrough.
			fallthrough
		case '{':
			// Do the parse.
			body, err := parser(r)
			if err != nil {
				return nil, err
			}

			// Create the catch token.
			prevLast := last
			last = &CatchToken{
				Exception:  exception,
				Position:   pos,
				Statements: body,
				Variable:   variable,
			}
			if prevLast == nil {
				first = last
			} else {
				prevLast.Next = last
			}
		default:
			// Unexpected character.
			return nil, &ParserError{
				Message:  "unexpected character '" + string(c) + "' after exception in catch statement",
				Position: getReaderPos(r),
			}
		}
	}
}

// Handle parsing a try statement.
func tryParser(r *strings.Reader, pos int, parser parserHn) (any, *ParserError) {
	// Gulp the whitespace.
	gulpWhitespace(r)

	// Read the next Unicode character.
	c, _, errIface := r.ReadRune()
	if errIface != nil {
		// Return an error.
		return nil, &ParserError{
			Message:  "unexpected end of file after 'try' in try statement",
			Position: getReaderPos(r),
		}
	}

	// Handle if the character is not '{'.
	if c != '{' {
		// Return an error.
		return nil, &ParserError{
			Message:  "unexpected character '" + string(c) + "' after 'try' in try statement",
			Position: getReaderPos(r),
		}
	}

	// Parse the body.
	body, err := parser(r)
	if err != nil {
		return nil, err
	}

	// Handle any caches.
	catches, err := catchesParser(r, pos, parser)
	if err != nil {
		return nil, err
	}
	if catches == nil {
		// No catches with the try statement.
		return nil, &ParserError{
			Message:  "expected 'catch' after 'try' in try statement",
			Position: getReaderPos(r),
		}
	}

	// Return the try token.
	return TryToken{
		Position:   pos,
		Statements: body,
		Catch:      catches,
	}, nil
}

// Handle cases where we break into another branch.
var branchingCases = map[string]func(*strings.Reader, int, parserHn) (any, *ParserError){
	"unless": unlessParser,
	"if":     ifParser,
	"throw":  throwParser,
	"return": returnParser,
	"while":  whileParser,
	"for":    forParser,
	"switch": switchParser,
	"try":    tryParser,
}

// Parses a potential assignment or branching case.
func parsePotentialAssignmentOrBranchingCase(r *strings.Reader, eot rune, allowBigStatements bool) (any, *ParserError) {
	// Defines the reference buffer. The reference buffer is stored in if something could
	// be either a assignment or reference. If it is the later, it is drained back into
	// the tokens array. If it is the former, it is parsed as a method call. refPos will be
	// -1 if there is no reference buffer.
	refBuf := ""
	refPos := -1

	// Attempts to drain the buffer if needed.
	drain := func() (any, *ParserError) {
		if refPos == -1 {
			// Nothing to drain.
			return nil, nil
		}

		// Set the position of the reader to the start of the thing that is not a method call.
		r.Seek(int64(refPos), io.SeekStart)

		// Parse the token.
		token, err := parseInnerContractTokenWithOpGrouping(r, eot)
		if err != nil {
			return nil, err
		}

		// Return the token.
		return token, nil
	}

	// Gulp the whitespace.
	gulpWhitespace(r)

	for {
		// Read the position.
		pos := getReaderPos(r)

		// Read the next Unicode character.
		c, _, err := r.ReadRune()
		if err != nil {
			// EOF in the middle of a contract body before the ending brace.
			return nil, &ParserError{
				Message:  "unexpected end of file before closing bracket",
				Position: pos,
			}
		}

		// Switch on the character.
		switch c {
		case eot:
			// Drain our local buffer and return.
			return drain()
		case '=':
			// Handle if the reference buffer is blank.
			if refPos == -1 {
				return nil, &ParserError{
					Message:  "unexpected '=' without assignment name",
					Position: pos,
				}
			}

			// This is not a reference, this is a assignment. Go ahead and parse the next token.
			next, err := parseInnerContractTokenWithOpGrouping(r, '}')
			if err != nil {
				// Assigned thing is not valid.
				return nil, err
			}

			// next == nil is ignored because it is picked up next loop iteration and will be a
			// syntax error.

			// Return the assignment token.
			return AssignmentToken{
				Name:     refBuf,
				Position: refPos,
				Value:    next,
			}, nil
		case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j',
			'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't',
			'u', 'v', 'w', 'x', 'y', 'z', '_', 'A', 'B', 'C',
			'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W',
			'X', 'Y', 'Z':
			if refPos == -1 {
				// If the reference buffer is blank, set the position.
				refPos = pos
			}

			// Add the character to the reference buffer.
			refBuf += string(c)
		case '\r', '\n', '\t', ' ':
			// If this is \r, handle this.
			useThisCheck := true
			if c == '\r' {
				// Initially, don't use this check.
				useThisCheck = false

				// Read the next char.
				x, _, _ := r.ReadRune()
				if x == '\n' {
					// +1 to the position and use this check.
					pos++
					useThisCheck = true
				} else {
					// Unread the rune.
					_ = r.UnreadRune()
				}
			}

			// Handle any branching cases.
			if useThisCheck && allowBigStatements {
				x := branchingCases[refBuf]
				if x != nil {
					// Get the position minus the length of the branching case.
					posMinusLen := pos - len(refBuf)

					// Gulp the whitespace.
					gulpWhitespace(r)

					// Return the method.
					return x(r, posMinusLen, parseInnerContract)
				}
			}
		default:
			// Rewind the rune.
			_ = r.UnreadRune()

			// Any other characters result in the existing buffer being drained.
			x, err := drain()
			if x != nil || err != nil {
				// Return the content because draining at least returned something.
				return x, err
			}

			// Use the default parser. Not a possible assignment but a token hit off the bat.
			return parseInnerContractTokenWithOpGrouping(r, eot)
		}
	}
}

var potentialAssignmentRef func(*strings.Reader, rune, bool) (any, *ParserError)

func init() {
	// Create a ref copy to stop circular dependency.
	potentialAssignmentRef = parsePotentialAssignmentOrBranchingCase
}

// Parses the inner of the contract and ending brace. Expects everything up to after the starting brace
// to be consumed.
func parseInnerContract(r *strings.Reader) ([]any, *ParserError) {
	// Defines the tokens that have been parsed.
	tokens := []any{}

	for {
		// Run the parser.
		token, err := parsePotentialAssignmentOrBranchingCase(r, '}', true)
		if err != nil {
			// Return the parser error.
			return nil, err
		}
		if token == nil {
			// End of contract. Return the tokens.
			return tokens, nil
		}

		// Add the token to the tokens.
		tokens = append(tokens, token)
	}
}
