// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package languages

import "remixdb.io/rpc/structure"

// Extension is the type used to define the extension of the compiled code.
type Extension string

// LanguageCompiler is the type used to define a compiler for the language. Returned is
// the extension and the compiled code.
type LanguageCompiler func(base *structure.Base) (map[Extension]string, error)

// Languages is a global map of languages to their compilers.
var Languages = map[string]LanguageCompiler{}

// Small little thing I can add to the language files to init them in a single line.
func initLanguage(name string, compiler LanguageCompiler) struct{} {
	Languages[name] = compiler
	return struct{}{}
}
