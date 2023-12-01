// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package languages

import (
	"embed"
	"errors"
	"reflect"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"remixdb.io/rpc/structure"
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

// Processes a Go template.
func processGoTemplate(name, tmpl string, data any, variables map[string]string) (string, error) {
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
		"Subtemplate": func(name string, data interface{}) (string, error) {
			tmpl, ok := subtemplates[name]
			if !ok {
				return "", errors.New("subtemplate not found")
			}

			return processGoTemplate(name, tmpl, data, variables)
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
