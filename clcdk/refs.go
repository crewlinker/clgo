package clcdk

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
)

type tokenRef struct {
	token *string
}

func (r tokenRef) ImportValue() *string {
	return r.token
}

// StrongRef represents a value that can be imported in another stack.
type StrongRef interface {
	// ImportValue will use the ref value through the use of Fn:ImportValue
	ImportValue() *string
}

// ExportValue uses the CDK's native method on the stack to export any 'v' that
// is a construct property.
func ExportValue(scope constructs.Construct, v any) StrongRef {
	return tokenRef{awscdk.Stack_Of(scope).ExportValue(v, nil)}
}
