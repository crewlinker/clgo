package clcdk

import (
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type tokenRef struct {
	token *string
}

func (r tokenRef) Dereference() *string {
	return awscdk.Fn_ImportValue(r.token)
}

// StdRef replaces [Ref] and uses the CDK native tooling for exporting values.
type StdRef interface {
	// Dereference the value carried by this reference. It results in the use
	// of Fn::ImportValue.
	Dereference() *string
}

// ExportValue uses the CDK's native method on the stack to export any 'v' that
// is a construct property. It replaces "Export".
func ExportValue(scope constructs.Construct, v any) StdRef {
	return tokenRef{awscdk.Stack_Of(scope).ExportValue(v, nil)}
}

// Ref can be passed between stacks to easy imports and exports.
type Ref interface {
	Import(scope constructs.Construct) *string
}

type ref struct {
	ident  string
	source awscdk.Stack
}

// Import the referenced value. Will also create a dependency from the stack of 'scope' ON the
// stack that exported the reference.
func (r ref) Import(scope constructs.Construct) *string {
	awscdk.Stack_Of(scope).AddDependency(r.source, jsii.String("reference: "+r.ident))

	return awscdk.Fn_ImportValue(jsii.String(r.ident))
}

// Export a value from the stack owning 'scope' in a standardized way.
//
// Deprecated: Export is deprecated, instead "ExportValue" should be used.
func Export(scope constructs.Construct, name ScopeName, conv Conventions, val *string) Ref {
	stack := awscdk.Stack_Of(scope)
	scope = name.ChildScope(scope)
	scope.ToString()

	desc := conv.Qualifier() + *scope.ToString()
	ident := strings.ReplaceAll(desc, "/", "")

	awscdk.NewCfnOutput(scope, jsii.String("Export"), &awscdk.CfnOutputProps{
		Value:       val,
		Description: scope.Node().Path(),
		ExportName:  jsii.String(ident),
	})

	return ref{
		ident:  ident,
		source: stack,
	}
}
