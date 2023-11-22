// Package clcdk contains reusable infrastructruce components using AWS CDK.
package clcdk

import (
	"fmt"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

var (
	// ScopeNameLeftBracket defined the left bracket for scope naming printing.
	ScopeNameLeftBracket = ""
	// ScopeNameRightBracket defined the right bracket for scope naming printing.
	ScopeNameRightBracket = ""
)

// ScopeName is the name of a scope.
type ScopeName string

// ChildScope returns a new scope named 'name'.
func (sn ScopeName) ChildScope(parent constructs.Construct) constructs.Construct {
	return constructs.NewConstruct(parent, jsii.String(sn.String()))
}

func (sn ScopeName) String() string {
	return fmt.Sprintf("%s%s%s", ScopeNameLeftBracket, string(sn), ScopeNameRightBracket)
}
