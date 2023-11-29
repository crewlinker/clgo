// Package clcdk contains reusable infrastructruce components using AWS CDK.
package clcdk

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// ScopeName is the name of a scope.
type ScopeName string

// ChildScope returns a new scope named 'name'.
func (sn ScopeName) ChildScope(parent constructs.Construct) constructs.Construct {
	return constructs.NewConstruct(parent, jsii.String(sn.String()))
}

func (sn ScopeName) String() string {
	return string(sn)
}
