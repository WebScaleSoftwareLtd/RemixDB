// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package languages

import (
	"embed"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"remixdb.io/internal/rpc/structure"
)

// Extension is the type used to define the extension of the compiled code.
type Extension string

// LanguageCompiler is the type used to define a compiler for the language. Returned is
// the extension and the compiled code.
type LanguageCompiler func(base *structure.Base, opts map[string]string) (map[Extension]string, error)

// Option is the type used to define an option for the compiler.
type Option struct {
	// Optional is used to define if the option is optional.
	Optional bool

	// Default is used to define the default value of the option.
	Default *string
}

// LanguageCompilerBase is the type used to define a compiler init for this language.
type LanguageCompilerBase struct {
	// Options is used to define the options for the compiler.
	Options map[string]Option

	// Compiler is used to define the compiler for the language.
	Compiler LanguageCompiler
}

// Languages is a global map of languages to their compilers.
var Languages = map[string]LanguageCompilerBase{}

// Small little thing I can add to the language files to init them in a single line.
func initLanguage(name string, compiler LanguageCompiler, opts map[string]Option) struct{} {
	Languages[name] = LanguageCompilerBase{
		Options:  opts,
		Compiler: compiler,
	}
	return struct{}{}
}

func ptr[T any](x T) *T { return &x }

//go:embed templates/switches/*.txt
var switchesFs embed.FS

var switches = map[string]map[string]string{}

func init() {
	files, err := switchesFs.ReadDir("templates/switches")
	if err != nil {
		panic(err)
	}
	for _, v := range files {
		name := v.Name()

		b, err := switchesFs.ReadFile("templates/switches/" + name)
		name = name[:len(name)-4]
		if err != nil {
			panic(err)
		}
		s := map[string]string{}
		for _, v := range strings.Split(string(b), "\n") {
			if v == "" {
				continue
			}

			if strings.HasPrefix(v, "#") {
				continue
			}

			index := strings.Index(v, " ")
			if index == -1 {
				s[v] = v
			} else {
				s[v[:index]] = v[index+1:]
			}
		}
		switches[name] = s
	}
}

//go:embed templates/subtemplates/*.tmpl
var subtemplatesFs embed.FS

var subtemplates = map[string]string{}

func init() {
	files, err := subtemplatesFs.ReadDir("templates/subtemplates")
	if err != nil {
		panic(err)
	}
	for _, v := range files {
		name := v.Name()

		b, err := subtemplatesFs.ReadFile("templates/subtemplates/" + name)
		name = name[:len(name)-5]
		if err != nil {
			panic(err)
		}
		subtemplates[name] = string(b)
	}
}

//go:embed templates/static/*.txt
var staticFs embed.FS

var static = map[string]string{}

func init() {
	files, err := staticFs.ReadDir("templates/static")
	if err != nil {
		panic(err)
	}
	for _, v := range files {
		name := v.Name()

		b, err := staticFs.ReadFile("templates/static/" + name)
		name = name[:len(name)-4]
		if err != nil {
			panic(err)
		}
		static[name] = string(b)
	}
}

var bracketRegex = regexp.MustCompile(`\{(.*?)\}`)

// Processes a Go template.
func processGoTemplate(root *structure.Base, name, tmpl string, data any, variables map[string]string) (string, error) {
	if data == nil {
		data = root
	}

	tpl, err := template.New(name).Funcs(template.FuncMap{
		"Switchfile": func(name string, case_ string) string {
			cases, ok := switches[name]
			if !ok {
				return case_
			}

			res, ok := cases[case_]
			if !ok {
				return case_
			}

			return res
		},
		"Variable": func(name string) string {
			return variables[name]
		},
		"PadToMax": func(items []string, s string) string {
			max := 0
			s = strcase.ToCamel(s)
			for _, v := range items {
				v = strcase.ToCamel(v)
				if len(v) > max {
					max = len(v)
				}
			}

			return strings.Repeat(" ", (max-len(s))+1)
		},
		"TitleCase": func(s string) string {
			return strcase.ToCamel(s)
		},
		"SplitLines": func(s string) []string {
			return strings.Split(s, "\n")
		},
		"HasTime": func(base *structure.Base) bool {
			for _, v := range base.Structs {
				for _, v := range v.Fields {
					if v.Type == "timestamp" {
						return true
					}
				}
			}
			for _, v := range base.Methods {
				if v.Input == "timestamp" || v.Output == "timestamp" {
					return true
				}
			}
			return false
		},
		"HasBigInt": func(base *structure.Base) bool {
			for _, v := range base.Structs {
				for _, v := range v.Fields {
					if v.Type == "bigint" {
						return true
					}
				}
			}
			for _, v := range base.Methods {
				if v.Input == "bigint" || v.Output == "bigint" {
					return true
				}
			}
			return false
		},
		"HasCursor": func(base *structure.Base) bool {
			for _, v := range base.Methods {
				if v.OutputBehaviour == structure.OutputBehaviourCursor {
					return true
				}
			}
			return false
		},
		"Subtemplate": func(name string, data any) (string, error) {
			tmpl, ok := subtemplates[name]
			if !ok {
				return "", errors.New("subtemplate not found")
			}

			return processGoTemplate(root, name, tmpl, data, variables)
		},
		"SwitchyTemplate": func(nameTpl, nameVar string, data any) (string, error) {
			var possibilities []string

			// Check if there's any curly brackets (optionals).
			if bracketRegex.MatchString(nameTpl) {
				// Create 2 templates. One with them removed from their curly brackets and just
				// in the string, and one with them removed from the string.
				t0 := bracketRegex.ReplaceAllString(nameTpl, "$1")
				t1 := bracketRegex.ReplaceAllString(nameTpl, "")

				// Define the possibilities with the optional taking priority.
				possibilities = []string{
					strings.Replace(t0, "$", nameVar, 1),
					strings.Replace(t0, "$", "default", 1),
					strings.Replace(t1, "$", nameVar, 1),
					strings.Replace(t1, "$", "default", 1),
				}
			} else {
				// Only 2 possibilities.
				possibilities = []string{
					strings.Replace(nameTpl, "$", nameVar, 1),
					strings.Replace(nameTpl, "$", "default", 1),
				}
			}

			// Loop through the possibilities and check if they exist.
			for _, v := range possibilities {
				if tmpl, ok := subtemplates[v]; ok {
					variables["__case_name"] = nameVar
					return processGoTemplate(root, v, tmpl, data, variables)
				}
			}

			return "", errors.New("subtemplate not found - tried: " + strings.Join(possibilities, ", "))
		},
		"Static": func(name string) string {
			return static[name]
		},
		"Keys": func(m any) []string {
			r := reflect.ValueOf(m)
			keys := r.MapKeys()
			res := make([]string, len(keys))
			for i, v := range keys {
				res[i] = v.String()
			}
			return res
		},
		"HasKey": func(m any, key string) bool {
			r := reflect.ValueOf(m)
			keys := r.MapKeys()
			for _, v := range keys {
				if v.String() == key {
					return true
				}
			}
			return false
		},
		"Tabify": func(c int, s string) string {
			split := strings.Split(s, "\n")
			for i, v := range split {
				if v == "" {
					continue
				}
				split[i] = strings.Repeat("\t", c) + v
			}
			return strings.Join(split, "\n")
		},
		"KeyAndValue": func(k, v any) map[string]any {
			return map[string]any{"Key": k, "Value": v}
		},
		"HashSchema": func(method structure.Method) string {
			// TODO
			return "method_hash_here"
		},
		"OutputOptCheck": func(root any) bool {
			switch x := root.(type) {
			case structure.Method:
				return x.OutputOptional
			case structure.StructField:
				return x.Optional
			default:
				// Just explode here!
				panic("OutputOptCheck: unknown type")
			}
		},
		"GetOutputType": func(root any) string {
			switch x := root.(type) {
			case structure.Method:
				return x.Output
			case structure.StructField:
				return x.Type
			default:
				// Just explode here!
				panic("GetOutputType: unknown type")
			}
		},
		"Root": func() *structure.Base {
			return root
		},
	}).Parse(tmpl)
	if err != nil {
		return "", err
	}

	var s strings.Builder
	err = tpl.Execute(&s, data)
	if err != nil {
		return "", err
	}
	return s.String(), nil
}

var optBoolMappings = map[string]bool{
	"true":  true,
	"false": false,
	"yes":   true,
	"no":    false,
	"1":     true,
	"0":     false,
	"yarr":  true,
	"narr":  false,
}

// Turns a option into a boolean.
func opt2bool(opt string) (bool, error) {
	if b, ok := optBoolMappings[strings.ToLower(opt)]; ok {
		return b, nil
	}
	return false, errors.New("invalid option: must be a boolean")
}
