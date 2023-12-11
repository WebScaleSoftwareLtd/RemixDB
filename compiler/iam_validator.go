// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package compiler

import (
	"go/ast"
	"go/token"
	"sort"
	"strconv"
)

type iamValidator struct {
	perms map[string]struct{}

	switchStmt *[]ast.Stmt
	binaryExpr *ast.BinaryExpr
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
			Tok: token.DEFINE,
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
	var ifStmts []ast.Stmt
	addToInterface("RespondWithRemixDBException", remixDbExceptionSig())
	noPermsResponse := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   ast.NewIdent("r"),
				Sel: ast.NewIdent("RespondWithRemixDBException"),
			},
			Args: []ast.Expr{
				ast.NewIdent("403"),
				ast.NewIdent("no_permission"),
				ast.NewIdent("You do not have permission to use this function."),
			},
		},
	}
	nilReturn := &ast.ReturnStmt{
		Results: []ast.Expr{
			ast.NewIdent("nil"),
		},
	}
	if isCursor {
		// Create the call for closing the underlying database session.
		closer := &ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ast.NewIdent("r"),
					Sel: ast.NewIdent("Close"),
				},
			},
		}

		// Make the body of the if statement include the above.
		ifStmts = []ast.Stmt{closer, noPermsResponse, nilReturn}
	} else {
		// Make the body of the if statement just the error and return.
		ifStmts = []ast.Stmt{noPermsResponse, nilReturn}
	}

	// Add the if statement to check if the user has the permission. At this stage it is just a placeholder.
	v.binaryExpr = &ast.BinaryExpr{}
	*s = append(*s, &ast.IfStmt{
		Cond: v.binaryExpr,
		Body: &ast.BlockStmt{
			List: ifStmts,
		},
	})

	// Add the label for the for post-IAM.
	*s = append(*s, &ast.LabeledStmt{
		Label: ast.NewIdent("postIam"),
	})
}

type stringStack struct {
	prev *stringStack
	val  string
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
			ast.NewIdent("*"),
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
		bitMapping[name] = bitMapping[name] | bit
	}

	// Now flip it around to find cases that can be combined.
	bitMappingReverse := map[uint64]*stringStack{}
	var bored uint64
	for name, bit := range bitMapping {
		bitMappingReverse[bit] = &stringStack{
			prev: bitMappingReverse[bit],
			val:  name,
		}
		bored = bored | bit
	}

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
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{
						&ast.BinaryExpr{
							X:  ast.NewIdent("userPerms"),
							Op: token.OR,
							Y:  ast.NewIdent(strconv.FormatUint(key, 10)),
						},
					},
				},
			},
		}

		// Add the case.
		switchReplace = append(switchReplace, caseClause)

		// Handle checking if we got all the permissions we expect.
		binaryExpr := ast.BinaryExpr{
			X:  ast.NewIdent("userPerms"),
			Op: token.EQL,
			Y:  ast.NewIdent(strconv.FormatUint(bored, 10)),
		}
		*v.binaryExpr = binaryExpr
	}

	// Replace the inside of the switch.
	*v.switchStmt = switchReplace
}
