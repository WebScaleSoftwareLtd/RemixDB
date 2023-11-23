package ast

// CommentToken is used to define a line which is a comment.
type CommentToken struct {
	// Comment is the comment string.
	Comment string

	// Position is the position of the comment.
	Position int
}

// DecoratorToken is used to define a decorator.
type DecoratorToken struct {
	// Method is the method name.
	Method string

	// Arguments are the arguments of the decorator.
	Arguments string

	// Position is the position of the decorator.
	Position int
}

// FieldToken is used to define a field in a struct.
type FieldToken struct {
	// Name is the name of the field.
	Name string

	// Type is the type of the field.
	Type string

	// Position is the position of the field.
	Position int

	// Decorators are the decorators of the field.
	Decorators []DecoratorToken
}

// ReferenceToken is used to define a reference to something that is implied to be a variable or
// field.
type ReferenceToken struct {
	// Name is the name of the variable.
	Name string

	// Position is the position of the field.
	Position int

	// Decorators are the decorators of the field. This is only used for structs.
	Decorators []DecoratorToken
}

// StructToken is used to define a struct.
type StructToken struct {
	// Name is the name of the struct.
	Name string

	// Position is the position of the struct.
	Position int

	// Decorators are the decorators of the struct.
	Decorators []DecoratorToken

	// Fields are the fields of the struct. They can be any of
	// CommentToken, FieldToken, or ReferenceToken.
	Fields []any
}

// ExtendsToken is used to define something that extends another thing.
type ExtendsToken struct {
	// Token is the token that is being wrapped.
	Token any

	// Position is the position of the extends.
	Position int
}

// ReturnToken is used to define a return statement.
type ReturnToken struct {
	// Token is the token that is being returned.
	Token any

	// Position is the position of the return.
	Position int
}

// ContractArgumentToken is used to define an argument of a contract.
type ContractArgumentToken struct {
	// Name is the name of the argument.
	Name string

	// NameIndex is the index of the name in the file.
	NameIndex int

	// Type is the type of the argument.
	Type string

	// TypeIndex is the index of the type in the file.
	TypeIndex int
}

// StringLiteralToken is used to define a string literal.
type StringLiteralToken struct {
	// Value is the value of the string literal.
	Value string

	// Position is the position of the string literal.
	Position int
}

// NumberLiteralToken is used to define a number literal.
type NumberLiteralToken struct {
	// Value is the value of the number literal.
	Value int

	// Position is the position of the number literal.
	Position int
}

// FloatLiteralToken is used to define a float literal.
type FloatLiteralToken struct {
	// Value is the value of the float literal.
	Value float64

	// Position is the position of the float literal.
	Position int
}

// BigIntLiteralToken is used to define a big integer literal.
type BigIntLiteralToken struct {
	// Value is the value of the big integer literal.
	Value string

	// Position is the position of the big integer literal.
	Position int
}

// BooleanLiteralToken is used to define a boolean literal.
type BooleanLiteralToken struct {
	// Value is the value of the boolean literal.
	Value bool

	// Position is the position of the boolean literal.
	Position int
}

// ArrayLiteralToken is used to define an array literal.
type ArrayLiteralToken struct {
	// Values are the values of the array literal. They can be any of
	// CommentToken, FieldToken, ReferenceToken, StringLiteralToken,
	// NumberLiteralToken, BigIntLiteralToken, or BooleanLiteralToken.
	Values []any

	// Position is the position of the array literal.
	Position int
}

// ObjectLiteralToken is used to define an object literal.
type ObjectLiteralToken struct {
	// Values are the values of the object literal. Keys are strings
	// and values can be any of ReferenceToken, StringLiteralToken,
	// NumberLiteralToken, BigIntLiteralToken, BooleanLiteralToken,
	// ArrayLiteralToken, or ObjectLiteralToken.
	Values map[string]any

	// Comments are the comments of the object literal.
	Comments []CommentToken

	// Position is the position of the object literal.
	Position int
}

// NullLiteralToken is used to define a null literal.
type NullLiteralToken struct {
	// Position is the position of the null literal.
	Position int
}

// MethodCallToken is used to define a method call.
type MethodCallToken struct {
	// Name is the name of the method. This includes all dots for the location.
	Name string

	// Position is the position of the method.
	Position int

	// Arguments are the arguments of the method. They can be any of
	// CommentToken, FieldToken, ReferenceToken, StringLiteralToken,
	// NumberLiteralToken, BigIntLiteralToken, BooleanLiteralToken,
	// ArrayLiteralToken, AddToken, or ObjectLiteralToken.
	Arguments []any

	// ChainedCall is the chained call of the method. Can be nil, MethodCallToken,
	// or ReferenceToken.
	ChainedCall any
}

// AssignmentToken is used to define an assignment.
type AssignmentToken struct {
	// Name is the name of the variable.
	Name string

	// Value is the value of the variable.
	Value any

	// Position is the position of the assignment.
	Position int
}

// AddToken is used to define an addition.
type AddToken struct {
	// Left is the left side of the addition.
	Left any

	// Right is the right side of the addition.
	Right any

	// Position is the position of the addition.
	Position int
}

// LessThanToken is used to define a less than.
type LessThanToken struct {
	// Left is the left side of the less than.
	Left any

	// Right is the right side of the less than.
	Right any

	// Position is the position of the less than.
	Position int
}

// GreaterThanToken is used to define a greater than.
type GreaterThanToken struct {
	// Left is the left side of the greater than.
	Left any

	// Right is the right side of the greater than.
	Right any

	// Position is the position of the greater than.
	Position int
}

// LessThanOrEqualToken is used to define a less than or equal.
type LessThanOrEqualToken struct {
	// Left is the left side of the less than or equal.
	Left any

	// Right is the right side of the less than or equal.
	Right any

	// Position is the position of the less than or equal.
	Position int
}

// GreaterThanOrEqualToken is used to define a greater than or equal.
type GreaterThanOrEqualToken struct {
	// Left is the left side of the greater than or equal.
	Left any

	// Right is the right side of the greater than or equal.
	Right any

	// Position is the position of the greater than or equal.
	Position int
}

// EqualToken is used to define an equal.
type EqualToken struct {
	// Left is the left side of the equal.
	Left any

	// Right is the right side of the equal.
	Right any

	// Position is the position of the equal.
	Position int
}

// NotEqualToken is used to define a not equal.
type NotEqualToken struct {
	// Left is the left side of the not equal.
	Left any

	// Right is the right side of the not equal.
	Right any

	// Position is the position of the not equal.
	Position int
}

// AndToken is used to define an and.
type AndToken struct {
	// Left is the left side of the and.
	Left any

	// Right is the right side of the and.
	Right any

	// Position is the position of the and.
	Position int
}

// OrToken is used to define an or.
type OrToken struct {
	// Left is the left side of the or.
	Left any

	// Right is the right side of the or.
	Right any

	// Position is the position of the or.
	Position int
}

// MultiplyToken is used to define an multiplication.
type MultiplyToken struct {
	// Left is the left side of the multiplication.
	Left any

	// Right is the right side of the multiplication.
	Right any

	// Position is the position of the multiplication.
	Position int
}

// SubtractToken is used to define an subtraction.
type SubtractToken struct {
	// Left is the left side of the subtraction.
	Left any

	// Right is the right side of the subtraction.
	Right any

	// Position is the position of the subtraction.
	Position int
}

// DivideToken is used to define an division.
type DivideToken struct {
	// Left is the left side of the division.
	Left any

	// Right is the right side of the division.
	Right any

	// Position is the position of the division.
	Position int
}

// ModuloToken is used to define an modulo.
type ModuloToken struct {
	// Left is the left side of the modulo.
	Left any

	// Right is the right side of the modulo.
	Right any

	// Position is the position of the modulo.
	Position int
}

// ExponentToken is used to define an exponent.
type ExponentToken struct {
	// Left is the left side of the exponent.
	Left any

	// Right is the right side of the exponent.
	Right any

	// Position is the position of the exponent.
	Position int
}

// ContractThrowsToken is used to define an exception that a contract can raise.
type ContractThrowsToken struct {
	// Name is the name of the exception.
	Name string

	// Position is the position of the exception.
	Position int
}

// ThrowLiteralToken is used to define a throw statement.
type ThrowLiteralToken struct {
	// Token is the token that is being thrown.
	Token any

	// Position is the position of the exception.
	Position int
}

// ContractToken is used to define a contract.
type ContractToken struct {
	// Name is the name of the contract.
	Name string

	// Argument is used to define the argument of the contract if present.
	Argument *ContractArgumentToken

	// ReturnType is the return type of the contract.
	ReturnType string

	// Position is the position of the contract.
	Position int

	// Throws is used to define the exceptions that the contract can raise.
	Throws []ContractThrowsToken

	// Decorators are the decorators of the mapping.
	Decorators []DecoratorToken

	// Statements are the statements of the contract.
	Statements []any
}

// MappingPartialToken is used to define a partial mapping.
type MappingPartialToken struct {
	// Position is the position of the mapping.
	Position int

	// Key is the key of the mapping.
	Key string

	// Value is the value of the mapping. Either a PartialMappingToken or
	// a string to reference it.
	Value any

	// Comments are the comments of the mapping.
	Comments []CommentToken
}

// MappingToken is used to define a mapping.
type MappingToken struct {
	MappingPartialToken

	// Name is the name of the mapping.
	Name string

	// Using is the structs that the mapping uses.
	Using []string

	// Decorators are the decorators of the mapping.
	Decorators []DecoratorToken
}

// ElseToken is used to define an else statement.
type ElseToken struct {
	// Position is the position of the else statement.
	Position int

	// Statements are the statements of the else statement.
	Statements []any

	// Condition is the condition of the else statement.
	Condition any

	// Next is the next else statement of the else statement.
	Next *ElseToken
}

// UnlessToken is used to define an unless statement.
type UnlessToken struct {
	// Condition is the condition of the unless statement.
	Condition any

	// Position is the position of the unless statement.
	Position int

	// Statements are the statements of the unless statement.
	Statements []any

	// Else is the else statement of the unless statement.
	Else *ElseToken
}

// IfToken is used to define an if statement.
type IfToken struct {
	// Condition is the condition of the if statement.
	Condition any

	// Position is the position of the if statement.
	Position int

	// Statements are the statements of the if statement.
	Statements []any

	// Else is the else statement of the if statement.
	Else *ElseToken
}

// ForToken is used to define an for statement.
type ForToken struct {
	// Assignment is the assignment of the for statement.
	Assignment any

	// Condition is the condition of the for statement.
	Condition any

	// Increment is the increment of the for statement.
	Increment any

	// Position is the position of the for statement.
	Position int

	// Statements are the statements of the for statement.
	Statements []any
}

// WhileToken is used to define an while statement.
type WhileToken struct {
	// Condition is the condition of the while statement.
	Condition any

	// Position is the position of the while statement.
	Position int

	// Statements are the statements of the while statement.
	Statements []any
}

// InlineIfToken is used to define an inline if statement.
type InlineIfToken struct {
	// Condition is the condition of the inline if statement.
	Condition any

	// Position is the position of the inline if statement.
	Position int

	// Token is the token of the inline if statement.
	Token any
}

// InlineUnlessToken is used to define an inline unless statement.
type InlineUnlessToken struct {
	// Condition is the condition of the inline unless statement.
	Condition any

	// Position is the position of the inline unless statement.
	Position int

	// Token is the token of the inline unless statement.
	Token any
}

// CatchToken is used to define an catch statement.
type CatchToken struct {
	// Position is the position of the catch statement.
	Position int

	// Statements are the statements of the catch statement.
	Statements []any

	// Exception is the exception of the catch statement. Can be blank.
	Exception string

	// Variable is the variable of the catch statement. Can be blank.
	Variable string

	// Next is the next catch statement of the catch statement.
	Next *CatchToken
}

// TryToken is used to define an try statement.
type TryToken struct {
	// Position is the position of the try statement.
	Position int

	// Statements are the statements of the try statement.
	Statements []any

	// Catch is the catch statement of the try statement.
	Catch *CatchToken
}

// SwitchCaseToken is used to define an case statement.
type SwitchCaseToken struct {
	// Position is the position of the case statement.
	Position int

	// Name is the name of the case statement.
	Name any

	// Statements are the statements of the case statement.
	Statements []any
}

// SwitchToken is used to define an switch statement.
type SwitchToken struct {
	// Position is the position of the switch statement.
	Position int

	// Condition is the condition of the switch statement.
	Condition any

	// Cases are the cases of the switch statement.
	Cases []SwitchCaseToken

	// Comments are the comments of the switch statement.
	Comments []CommentToken
}

// NotToken is used to define a not statement.
type NotToken struct {
	// Position is the position of the not statement.
	Position int

	// Token is the token of the not statement.
	Token any
}
