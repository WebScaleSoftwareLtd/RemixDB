// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package structure

// StructField is used to define a field within a structure.
type StructField struct {
	// Comment is used to define the comment. Can be blank.
	Comment string `json:"comment"`

	// Type is used to define the type of the field. Built-in types are
	// "string", "uint", "int", "float", "bigint", "timestamp", "bool", and "bytes".
	// If the type is not built-in, it is a structure.
	Type string `json:"type"`

	// Array is used to define if the field is an array.
	Array bool `json:"array"`

	// Optional is used to define if the field is optional.
	Optional bool `json:"optional"`
}

// Struct is used to define a structure within the RPC.
type Struct struct {
	// Comment is used to define the comment. Can be blank.
	Comment string `json:"comment"`

	// Exception is used to define if the structure is an exception. In a exception,
	// 'Message' is used as the error message, or a repersentation of all of the
	// fields.
	Exception bool `json:"exception"`

	// Fields is used to define the fields within the structure.
	Fields map[string]StructField `json:"fields"`
}

// OutputBehaviour is used to define the behaviour of the output.
type OutputBehaviour string

const (
	// OutputBehaviourSingle is used to define that the output is a single value.
	OutputBehaviourSingle OutputBehaviour = "single"

	// OutputBehaviourArray is used to define that the output is an array of values.
	OutputBehaviourArray OutputBehaviour = "array"

	// OutputBehaviourCursor is used to define that the output is a cursor of values.
	OutputBehaviourCursor OutputBehaviour = "cursor"
)

// Method is used to define a method within the RPC.
type Method struct {
	// Comment is used to define the comment. Can be blank.
	Comment string `json:"comment"`

	// Input is used to define the input structure. Built-in types are
	// "string", "uint", "int", "float", "bigint", "timestamp", "bool", and "bytes".
	// If the type is not built-in, it is a structure. If it is blank, there is
	// no input.
	Input string `json:"input"`

	// InputName is used to define the name of the input structure. Required if
	// Input is not blank.
	InputName string `json:"input_name"`

	// InputOptional is used to define if the input is optional.
	InputOptional bool `json:"input_optional"`

	// Output is used to define the output structure. Built-in types are
	// "string", "uint", "int", "float", "bigint", "timestamp", "bool", and "bytes".
	// If the type is not built-in, it is a structure. If it is blank, there is
	// no output.
	Output string `json:"output"`

	// OutputOptional is used to define if the output is optional.
	OutputOptional bool `json:"output_optional"`

	// OutputBehaviour is used to define the behaviour of the output. Defaults to
	// OutputBehaviourSingle. Can be OutputBehaviourSingle, OutputBehaviourArray,
	// or OutputBehaviourCursor.
	OutputBehaviour OutputBehaviour `json:"output_behaviour"`
}

// Base is used to define the base RPC structure.
type Base struct {
	// Structs is used to define the structures within the language.
	Structs map[string]Struct `json:"structs"`

	// Methods is used to define the methods within the language.
	Methods map[string]Method `json:"methods"`

	// AuthenticationKeys are keys that should be included with every request to
	// authenticate the user. If they are not included, they will be blank in
	// the request.
	AuthenticationKeys []string `json:"authentication_keys"`
}
