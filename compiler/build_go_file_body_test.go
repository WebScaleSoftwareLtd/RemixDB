// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package compiler

import (
	"testing"

	"github.com/jimeh/go-golden"
	_ "github.com/matryer/moq/pkg/moq"
	"github.com/stretchr/testify/assert"
	"remixdb.io/ast"
	"remixdb.io/compiler/mocksession"
)

//go:generate go run generate_mock_session_implementation.go

func Test_contract2go(t *testing.T) {
	tests := []struct {
		name string

		sessionMockSetup func(hn *mocksession.SessionMock)
		contract         *ast.ContractToken
	}{
		{
			name: "string output returning input",
			contract: &ast.ContractToken{
				Name: "Test",
				Argument: &ast.ContractArgumentToken{
					Name: "input",
					Type: "string",
				},
				ReturnType: "string",
				Statements: []any{
					ast.ReturnToken{
						Token: ast.ReferenceToken{
							Name: "input",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the mock session and do any setup which is wanted.
			mock := &mocksession.SessionMock{}
			if tt.sessionMockSetup != nil {
				tt.sessionMockSetup(mock)
			}

			// Run the contract2go function.
			goCode, err := contract2go(tt.contract, mock)
			assert.NoError(t, err)

			// Check the output with a golden file.
			if golden.Update() {
				golden.Set(t, []byte(goCode))
			}
			assert.Equal(t, string(golden.Get(t)), goCode)
		})
	}
}
