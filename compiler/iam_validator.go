// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package compiler

import (
	"go/ast"
	"go/token"
	"sort"
	"strconv"
	"strings"
)

type iamValidator struct {
	perms map[string]struct{}

	switchStmt *[]ast.Stmt
}

func (v *iamValidator) addValidator(name string) {
	if v.perms == nil {
		v.perms = map[string]struct{}{}
	}
	v.perms[name] = struct{}{}
}

func (v *iamValidator) injectEntrypoint(
	s *[]ast.Stmt, addToInterface func(name string, fn *ast.FuncType), isCursor bool,
) {
	// Add a uint64 variable to count the user permissions.
	*s = append(*s, &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						ast.NewIdent("userPerms"),
					},
					Type: ast.NewIdent("uint64"),
				},
			},
		},
	})

	// Defines the switch statement.
	sw := &ast.SwitchStmt{
		Tag: ast.NewIdent("perm"),
		Body: &ast.BlockStmt{
			List: []ast.Stmt{},
		},
	}
	v.switchStmt = &sw.Body.List

	// Add the for range loop over r.Permissions().
	addToInterface("Permissions", &ast.FuncType{
		Params: &ast.FieldList{},
		Results: &ast.FieldList{
			List: []*ast.Field{
				{
					Type: ast.NewIdent("[]string"),
				},
			},
		},
	})
	*s = append(*s, &ast.RangeStmt{
		Key:   ast.NewIdent("_"),
		Value: ast.NewIdent("perm"),
		Tok:   token.DEFINE,
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent("r"),
				Sel: ast.NewIdent("Permissions"),
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{sw},
		},
	})

	// Handle the no permission errors.
	if isCursor {
		// Create the call for closing the underlying database session.
		addToInterface("Close", noParamsJustError())
		*s = append(*s, &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ast.NewIdent("r"),
					Sel: ast.NewIdent("Close"),
				},
			},
		})
	}

	// Send the no permission error since we didn't jump.
	addToInterface("RespondWithRemixDBException", remixDbExceptionSig())
	*s = append(*s, &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent("r"),
				Sel: ast.NewIdent("RespondWithRemixDBException"),
			},
			Args: []ast.Expr{
				ast.NewIdent("403"),
				ast.NewIdent(`"no_permission"`),
				ast.NewIdent(`"You do not have permission to use this contract."`),
			},
		},
	})

	// Return nil since we are done.
	*s = append(*s, &ast.ReturnStmt{
		Results: []ast.Expr{
			ast.NewIdent("nil"),
		},
	})

	// Add the label for the for post-IAM.
	*s = append(*s, &ast.LabeledStmt{
		Label: ast.NewIdent("postIam"),
		Stmt:  &ast.EmptyStmt{},
	})
}

type stringStack struct {
	prev *stringStack
	val  string
}

func iamPossibilitiesGenerator(s string, hn func(string)) {
	// Split the string into two parts to handle any single word cases.
	sp := strings.SplitN(s, ":", 2)
	if len(sp) == 1 {
		hn(s)
		return
	}

	// Call with the standard string.
	hn(s)

	// Call with a wildcard.
	hn(sp[0] + ":*")
}

func (v *iamValidator) compile() {
	// Defines what we are going to replace the switch statement with.
	switchReplace := []ast.Stmt{}

	// Get the permission names.
	permNames := make([]string, len(v.perms))
	i := 0
	for name := range v.perms {
		permNames[i] = name
		i++
	}
	sort.Strings(permNames)

	// Create the special * case.
	switchReplace = append(switchReplace, &ast.CaseClause{
		List: []ast.Expr{
			ast.NewIdent(`"*"`),
		},
		Body: []ast.Stmt{
			&ast.BranchStmt{
				Tok:   token.GOTO,
				Label: ast.NewIdent("postIam"),
			},
		},
	})

	// Create a mapping to handle duplication.
	bitMapping := map[string]uint64{}
	for i, name := range permNames {
		// Get the bit reperesentation of the permission index.
		bit := uint64(1 << i)

		// Do binary OR with the mapping.
		iamPossibilitiesGenerator(name, func(name string) { bitMapping[name] = bitMapping[name] | bit })
	}

	// Now flip it around to find cases that can be combined.
	bitMappingReverse := map[uint64]*stringStack{}
	var bored uint64
	for name, bit := range bitMapping {
		bitMappingReverse[bit] = &stringStack{
			prev: bitMappingReverse[bit],
			val:  name,
		}
		bored |= bit
	}
	boredS := strconv.FormatUint(bored, 10)

	// Get all of the keys and sort them.
	keys := make([]uint64, len(bitMappingReverse))
	i = 0
	for key := range bitMappingReverse {
		keys[i] = key
		i++
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	// Create the cases.
	for _, key := range keys {
		// Get the string stack.
		stack := bitMappingReverse[key]

		// Defines the case keys.
		caseKeys := []ast.Expr{}
		for stack != nil {
			caseKeys = append(caseKeys, ast.NewIdent(`"`+stack.val+`"`))
			stack = stack.prev
		}

		// Sort the case keys.
		sort.Slice(caseKeys, func(i, j int) bool {
			return caseKeys[i].(*ast.Ident).Name < caseKeys[j].(*ast.Ident).Name
		})

		// Create the case.
		caseClause := &ast.CaseClause{
			List: caseKeys,
			Body: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						ast.NewIdent("userPerms"),
					},
					Tok: token.OR_ASSIGN,
					Rhs: []ast.Expr{
						ast.NewIdent(strconv.FormatUint(key, 10)),
					},
				},
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{
						X:  ast.NewIdent("userPerms"),
						Op: token.EQL,
						Y:  ast.NewIdent(boredS),
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.BranchStmt{
								Tok:   token.GOTO,
								Label: ast.NewIdent("postIam"),
							},
						},
					},
				},
			},
		}

		// Add the case.
		switchReplace = append(switchReplace, caseClause)
	}

	// Replace the inside of the switch.
	*v.switchStmt = switchReplace
}
