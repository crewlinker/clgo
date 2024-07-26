// Package cloapi provides an OpenAPI specced primitives.
package cloapi

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"html/template"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
)

// LoadSchemaTmpl will parse the OpenAPI3 schema.
func LoadSchemaTmpl(file []byte) (*openapi3.T, error) {
	doc, err := openapi3.NewLoader().LoadFromData(file)
	if err != nil {
		return nil, fmt.Errorf("failed to load schema: %w", err)
	}

	return doc, nil
}

// DecorateSchemaTmpl will add project specific extensions and output it as
// go template file that needs to be executed to fill in deploy-time values.
func DecorateSchemaTmpl(doc *openapi3.T) error {
	for _, path := range doc.Paths.InMatchingOrder() {
		item := doc.Paths.Find(path)
		for _, opr := range item.Operations() {
			if opr.Extensions == nil {
				opr.Extensions = map[string]any{}
			}

			// add AWS-specific integration attributes
			opr.Extensions["x-amazon-apigateway-integration"] = map[string]any{
				"httpMethod": http.MethodPost,
				"type":       "AWS_PROXY",
				"uri":        `{{.AwsProxyIntegrationURI}}`,
			}
		}
	}

	return nil
}

// SchemaDeployment describes the data necessary to deploy the schema from its template.
type SchemaDeployment struct {
	Title                  string
	Description            string
	AwsProxyIntegrationURI string
	AwsAuthorizerURI       string
}

// ExecuteSchemaTmpl executes the schema template file using deployment parameters. It retrurns any error
// and the hash of the template src so it can be used for content-based versioning.
func ExecuteSchemaTmpl(src []byte, depl SchemaDeployment) (string, [32]byte, error) {
	// in practice, the templated values will contain CDK Tokens that change on every deploy
	// so if we hash AFTER rendering the template the hash will change on every deploy.
	sum := sha256.Sum256(src)

	tmpl, err := template.New("schema").Parse(string(src))
	if err != nil {
		return "", sum, fmt.Errorf("failed to parse definition yml: %w", err)
	}

	buf := bytes.NewBuffer(nil)
	if err := tmpl.Option("missingkey=error").Execute(buf, depl); err != nil {
		return "", sum, fmt.Errorf("failed to render template: %w", err)
	}

	return buf.String(), sum, nil
}
