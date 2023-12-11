// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package compiler

import (
	goAst "go/ast"
	"go/token"
	"regexp"

	"remixdb.io/ast"
	"remixdb.io/engine"
)

// Handles adding to a interface.
func addToInterface(
	used map[string]struct{}, iface *goAst.InterfaceType, name string, funcType *goAst.FuncType,
) {
	// If it was previously used, skip it.
	if _, ok := used[name]; ok {
		return
	}

	// Add the method to the interface.
	iface.Methods.List = append(iface.Methods.List, &goAst.Field{
		Names: []*goAst.Ident{
			goAst.NewIdent(name),
		},
		Type: funcType,
	})

	// Mark it as used.
	used[name] = struct{}{}
}

// Shortcut for a function that returns an error with no parameters.
func noParamsJustError() *goAst.FuncType {
	return &goAst.FuncType{
		Params: &goAst.FieldList{},
		Results: &goAst.FieldList{
			List: []*goAst.Field{
				{
					Type: goAst.NewIdent("error"),
				},
			},
		},
	}
}

// Checks if the output is Cursor<T>. This is a special case where we return a cursor.
var cursorBuiltin = regexp.MustCompile(`^Cursor<(.+)>$`)

// Handles building the function body.
func buildFunctionBody(
	contract *ast.ContractToken, s engine.Session, iface *goAst.InterfaceType,
) (imports []string, body []goAst.Stmt, err error) {
	// Check if this is a cursor.
	matches := cursorBuiltin.FindStringSubmatch(contract.ReturnType)
	isCursor := false
	//outputType := contract.ReturnType
	if matches != nil {
		isCursor = true
		//outputType = matches[1]
	}

	// Defines all already used things in the interface.
	used := map[string]struct{}{}

	// If the contract is not of a cursor type, add a defer to close the cursor.
	if !isCursor {
		// Add Close to the interface.
		addToInterface(used, iface, "Close", noParamsJustError())

		// Add the defer into the body.
		body = append(body, &goAst.DeferStmt{
			Call: &goAst.CallExpr{
				Fun: &goAst.SelectorExpr{
					X:   goAst.NewIdent("r"),
					Sel: goAst.NewIdent("Close"),
				},
			},
		})
	}

	if contract.Argument != nil {
		// Capture the body into a variable.
		addToInterface(used, iface, "Body", &goAst.FuncType{
			Params: &goAst.FieldList{},
			Results: &goAst.FieldList{
				List: []*goAst.Field{
					{
						Type: goAst.NewIdent("[]byte"),
					},
				},
			},
		})
		body = append(body, &goAst.AssignStmt{
			Lhs: []goAst.Expr{
				goAst.NewIdent("body"),
			},
			Tok: token.DEFINE,
			Rhs: []goAst.Expr{
				&goAst.CallExpr{
					Fun: goAst.NewIdent("r.Body"),
				},
			},
		})
	}

	// In all cases, return nil for the error at the end.
	body = append(body, &goAst.ReturnStmt{
		Results: []goAst.Expr{
			goAst.NewIdent("nil"),
		},
	})

	// Just return the values created.
	return
}
