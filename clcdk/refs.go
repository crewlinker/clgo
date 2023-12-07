package clcdk

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsssm"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
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

// WeakRef represents a reference to a value stored in the AWS SSM parameter store. It can be read
// by the receiving stack without a token being involved. It is a weak link though and when the value
// is changed the receiving stack doesn't automatically change with it. Prefer "ExportValue" if
// possible and use this only in case this is not possible.
type WeakRef interface {
	LookupValue(scope constructs.Construct) *string
}

type weakRef struct {
	src       awscdk.Stack
	paramName string
}

func (f weakRef) LookupValue(scope constructs.Construct) *string {
	awscdk.Stack_Of(scope).AddDependency(f.src, jsii.String("SSM Parameter: "+f.paramName))

	return awsssm.StringParameter_ValueFromLookup(scope, jsii.String(f.paramName))
}

// WeakExportAttribute uses the SSM parameter store to make an attribute available. It creates a weak link as
// the parameter can be changed without the receiving stack being updated of the fact.
func WeakExportAttribute(scope constructs.Construct, resource constructs.IConstruct, attributeName string) WeakRef {
	rv := reflect.ValueOf(resource)

	met := rv.MethodByName(attributeName)
	if !met.IsValid() {
		panic("clcdk: invalid method for attribute: " + attributeName)
	}

	res := met.Call(nil)
	if len(res) != 1 {
		panic("clcdk: attribute method must return exactly one value, got: " + strconv.Itoa(len(res)))
	}

	val, ok := res[0].Interface().(*string)
	if !ok {
		panic(fmt.Sprintf("clckd: the attribute method must return a *string value, got: %T", val))
	}

	name := *awscdk.Names_UniqueResourceName(resource, &awscdk.UniqueResourceNameOptions{}) + attributeName
	paramName := "/clcdk/" + name

	awsssm.NewStringParameter(scope, jsii.String(name), &awsssm.StringParameterProps{
		ParameterName: jsii.String(paramName),
		StringValue:   val,
	})

	return weakRef{src: awscdk.Stack_Of(scope), paramName: paramName}
}
