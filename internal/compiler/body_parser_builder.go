// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package compiler

import (
	goAst "go/ast"
	"go/token"
	"strings"

	"remixdb.io/ast"
)

// Creates the body parser.
func (f *functionBody) createBodyParser(contract *ast.ContractToken) {
	// Check if the type ends in ?.
	type_ := contract.Argument.Type
	optional := strings.HasSuffix(type_, "?")
	if optional {
		type_ = type_[:len(type_)-1]
	}

	// Define a variable in the go/ast package of a blank ident.
	typeIdent := goAst.NewIdent("")
	f.body = append(f.body, &goAst.DeclStmt{
		Decl: &goAst.GenDecl{
			Tok: token.VAR,
			Specs: []goAst.Spec{
				&goAst.ValueSpec{
					Names: []*goAst.Ident{
						goAst.NewIdent("body"),
					},
					Type: typeIdent,
				},
			},
		},
	})

	// Set the type ident to * if it is optional.
	if optional {
		typeIdent.Name = "*"
	}

	// Defines the inner body if this is optional.
	var innerBody *goAst.BlockStmt
	if optional {
		// Put everything inside a 'if len(rawBody) != 1 || rawBody[0] != 0x00' block.
		innerBody = &goAst.BlockStmt{
			List: []goAst.Stmt{},
		}
		f.body = append(f.body, &goAst.IfStmt{
			Cond: &goAst.BinaryExpr{
				X: &goAst.BinaryExpr{
					X: &goAst.CallExpr{
						Fun: goAst.NewIdent("len"),
						Args: []goAst.Expr{
							goAst.NewIdent("rawBody"),
						},
					},
					Op: token.NEQ,
					Y: &goAst.BasicLit{
						Kind:  token.INT,
						Value: "1",
					},
				},
				Op: token.LOR,
				Y: &goAst.BinaryExpr{
					X: &goAst.IndexExpr{
						X: &goAst.Ident{
							Name: "rawBody",
						},
						Index: &goAst.BasicLit{
							Kind:  token.INT,
							Value: "0",
						},
					},
					Op: token.NEQ,
					Y: &goAst.BasicLit{
						Kind:  token.INT,
						Value: "0x00",
					},
				},
			},
			Body: innerBody,
		})
	}

	// Switch on the type.
	switch type_ {
	case "bool":
		// Add to the type name.
		typeIdent.Name += "bool"

		// Defines the error for a unexpected type.
		error_ := "\"Expected the type of a "
		if optional {
			error_ += "optional "
		}
		error_ += "bool for the input.\""

		// Check if the body is not 1 byte or the byte is not 0x01 or 0x02.
		ifStmt := &goAst.IfStmt{
			Cond: &goAst.BinaryExpr{
				X: &goAst.BinaryExpr{
					X: &goAst.CallExpr{
						Fun: goAst.NewIdent("len"),
						Args: []goAst.Expr{
							goAst.NewIdent("rawBody"),
						},
					},
					Op: token.NEQ,
					Y: &goAst.BasicLit{
						Kind:  token.INT,
						Value: "1",
					},
				},
				Op: token.LOR,
				Y: &goAst.BinaryExpr{
					X: &goAst.BinaryExpr{
						X: &goAst.IndexExpr{
							X: &goAst.Ident{
								Name: "rawBody",
							},
							Index: &goAst.BasicLit{
								Kind:  token.INT,
								Value: "0",
							},
						},
						Op: token.NEQ,
						Y: &goAst.BasicLit{
							Kind:  token.INT,
							Value: "0x01",
						},
					},
					Op: token.LAND,
					Y: &goAst.BinaryExpr{
						X: &goAst.IndexExpr{
							X: &goAst.Ident{
								Name: "rawBody",
							},
							Index: &goAst.BasicLit{
								Kind:  token.INT,
								Value: "0",
							},
						},
						Op: token.NEQ,
						Y: &goAst.BasicLit{
							Kind:  token.INT,
							Value: "0x02",
						},
					},
				},
			},
			Body: &goAst.BlockStmt{
				List: []goAst.Stmt{
					// Call the exception handler and return nil.
					&goAst.ExprStmt{
						X: &goAst.CallExpr{
							Fun: &goAst.SelectorExpr{
								X:   goAst.NewIdent("r"),
								Sel: goAst.NewIdent("RespondWithRemixDBException"),
							},
							Args: []goAst.Expr{
								&goAst.BasicLit{
									Kind:  token.INT,
									Value: "400",
								},
								&goAst.BasicLit{
									Kind:  token.STRING,
									Value: `"invalid_body"`,
								},
								&goAst.BasicLit{
									Kind:  token.STRING,
									Value: error_,
								},
							},
						},
					},
					&goAst.ReturnStmt{
						Results: []goAst.Expr{
							goAst.NewIdent("nil"),
						},
					},
				},
			},
		}

		// Check if this is optional.
		if optional {
			// Build a block in which we check if rawBody[0] == 0x02. Set the body to a pointer of that.
			block := &goAst.BlockStmt{
				List: []goAst.Stmt{
					&goAst.AssignStmt{
						Lhs: []goAst.Expr{
							goAst.NewIdent("b"),
						},
						Tok: token.DEFINE,
						Rhs: []goAst.Expr{
							&goAst.BinaryExpr{
								X: &goAst.IndexExpr{
									X: &goAst.Ident{
										Name: "rawBody",
									},
									Index: &goAst.BasicLit{
										Kind:  token.INT,
										Value: "0",
									},
								},
								Op: token.EQL,
								Y: &goAst.BasicLit{
									Kind:  token.INT,
									Value: "0x02",
								},
							},
						},
					},
					&goAst.AssignStmt{
						Lhs: []goAst.Expr{
							goAst.NewIdent("body"),
						},
						Tok: token.ASSIGN,
						Rhs: []goAst.Expr{
							goAst.NewIdent("&b"),
						},
					},
				},
			}
			innerBody.List = append(innerBody.List, ifStmt, block)
		} else {
			// Just set the body to the result of rawBody[0] == 0x02.
			f.body = append(f.body, ifStmt, &goAst.AssignStmt{
				Lhs: []goAst.Expr{
					goAst.NewIdent("body"),
				},
				Tok: token.ASSIGN,
				Rhs: []goAst.Expr{
					&goAst.BinaryExpr{
						X: &goAst.IndexExpr{
							X: &goAst.Ident{
								Name: "rawBody",
							},
							Index: &goAst.BasicLit{
								Kind:  token.INT,
								Value: "0",
							},
						},
						Op: token.EQL,
						Y: &goAst.BasicLit{
							Kind:  token.INT,
							Value: "0x02",
						},
					},
				},
			})
		}
	}
}
