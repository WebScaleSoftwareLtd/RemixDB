// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package languages

import (
	_ "embed"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
	"remixdb.io/internal/rpc/structure"
)

//go:embed templates/javascript.js
var jsTemplate string

//go:embed templates/javascript.d.ts
var jsDtsTemplate string

var methodsAnchorRegex = regexp.MustCompile(`([ \t]+)\/\/ AUTO-GENERATION MARKER: methods`)

func handleJsStructures(base *structure.Base, spacing string) string {
	// TODO
	return ""
}

func prefixJsComments(comments, prefix string) string {
	if comments == "" {
		return ""
	}
	return prefix + "// " + strings.TrimSpace(
		strings.ReplaceAll(comments, "\n", "\n"+prefix+"// "),
	) + "\n"
}

func jsRpcType(i string, nullable bool) string {
	switch i {
	case "string":
		i = "String"
	case "bool":
		i = "Boolean"
	case "int":
		i = "Number"
	case "bytes":
		i = "Uint8Array"
	case "timestamp":
		i = "Date"
	case "bigint":
		i = "BigInt"
	case "uint":
		i = "_uint"
	case "float":
		i = "_float"
	}
	if nullable {
		return "[" + i + ", null]"
	}
	return i
}

func generateJsMethods(base *structure.Base, spacing string) string {
	jsFuncs := ""
	for methodName, method := range base.Methods {
		// Handle comments.
		if method.Comment != "" {
			jsFuncs += prefixJsComments(method.Comment, spacing)
		}

		// Create the JS method signature.
		jsFuncs += spacing + strcase.ToCamel(methodName) + "("
		if method.Input != "" {
			jsFuncs += method.InputName
		}
		jsFuncs += ") {\n"

		// Multiply the spacing by 2.
		spacing2 := spacing + spacing

		// Validate the input.
		if method.Input != "" {
			jsFuncs += spacing2 + "_validateType(" + method.InputName + ", " + jsRpcType(method.Input, method.InputOptional) + ");\n"
		}

		// Encode the body if it is present.
		if method.Input == "" {
			jsFuncs += spacing2 + "const _body = new Uint8Array(0);\n"
		} else {
			jsFuncs += spacing2 + "const _body = _encode(" + method.InputName + ");\n"
		}

		// Get the schema hash.
		// TODO: Do the hash.
		schemaHash := "TODO"

		// Get the output type.
		outputType := method.Output
		if outputType == "" {
			outputType = "null"
		} else {
			outputType = jsRpcType(outputType, method.OutputOptional)
		}
		if method.OutputBehaviour == structure.OutputBehaviourArray {
			outputType = "[" + outputType + "]"
		}

		// Here goes!
		if method.OutputBehaviour == structure.OutputBehaviourCursor {
			jsFuncs += spacing2 + "return this._doCursorRequest(\"" + methodName + "\", _body, \"" + schemaHash + "\", " + outputType + ");\n"
		} else {
			jsFuncs += spacing2 + "return this._nonCursorRequest(\"" + methodName + "\", _body, \"" + schemaHash + "\", " + outputType + ");\n"
		}

		// Close the method.
		jsFuncs += spacing + "}\n\n"
	}
	return strings.TrimSpace(jsFuncs)
}

func jsGen(base *structure.Base, isNode, isEsm bool) string {
	// Deal with the imports marker.
	imports := ""
	if isNode {
		if isEsm {
			imports = "\n\nimport WebSocket from \"ws\";"
		} else {
			imports = "\n\nconst WebSocket = require(\"ws\");"
		}
	}
	jsTemplate := strings.Replace(jsTemplate, "\n// AUTO-GENERATION MARKER: imports", imports, 1)

	// Deal with the methods marker.
	spacing := methodsAnchorRegex.FindStringSubmatch(jsTemplate)[1]
	methods := generateJsMethods(base, spacing)
	jsTemplate = strings.Replace(jsTemplate, spacing+"// AUTO-GENERATION MARKER: methods", methods, 1)

	// Deal with the structures marker.
	structures := handleJsStructures(base, spacing)
	jsTemplate = strings.Replace(jsTemplate, "\n// AUTO-GENERATION MARKER: structures", structures, 1)

	// Deal with the exports object.
	exports := ""
	for structName := range base.Structs {
		exports += "\n" + spacing + structName + ","
	}
	jsTemplate = strings.Replace(jsTemplate, " // AUTO-GENERATION MARKER: exports", exports, 1)

	// Deal with the package exports marker.
	var packageExports string
	if isEsm {
		packageExports = "export _exports;"
	} else {
		packageExports = "module.exports = _exports;"
	}
	jsTemplate = strings.Replace(jsTemplate, "// AUTO-GENERATION MARKER: packageExports", packageExports, 1)

	// Return the JS.
	return jsTemplate
}

func jsDtsGen(base *structure.Base, isNode, isEsm bool) string {
	// TODO
	return jsDtsTemplate
}

func js(base *structure.Base, opts map[string]string) (map[Extension]string, error) {
	isEsm, err := opt2bool(opts["esm"])
	if err != nil {
		return nil, err
	}
	isNode, err := opt2bool(opts["node"])
	if err != nil {
		return nil, err
	}

	// Generate the JS.
	js := jsGen(base, isNode, isEsm)

	// Generate the JS DTS.
	jsDts := jsDtsGen(base, isNode, isEsm)

	// Return the JS and JS DTS.
	return map[Extension]string{"js": js, "d.ts": jsDts}, nil
}

var _ = initLanguage("js", js, map[string]Option{
	"esm": {
		Optional: true,
		Default:  ptr("true"),
	},
	"node": {
		Optional: false,
	},
})
