package modelschema

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dapr/kubernetes-operator/pkg/generated/openapi"
	"github.com/spf13/cobra"
	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

const (
	cmdName = "modelschema"
)

func NewCmd() *cobra.Command {
	// Outputs openAPI schema JSON containing the schema definitions in zz_generated.openapi.go.
	cmd := cobra.Command{
		Use:   cmdName,
		Short: cmdName,
		RunE: func(cmd *cobra.Command, args []string) error {
			refFunc := func(name string) spec.Ref {
				return spec.MustCreateRef("#/definitions/" + friendlyName(name))
			}

			defs := openapi.GetOpenAPIDefinitions(refFunc)
			schemaDefs := make(map[string]spec.Schema, len(defs))

			for k, v := range defs {
				// Replace top-level schema with v2 if a v2 schema is embedded
				// so that the output of this program is always in OpenAPI v2.
				// This is done by looking up an extension that marks the embedded v2
				// schema, and, if the v2 schema is found, make it the resulting schema for
				// the type.
				if schema, ok := v.Schema.Extensions[common.ExtensionV2Schema]; ok {
					if v2Schema, isOpenAPISchema := schema.(spec.Schema); isOpenAPISchema {
						schemaDefs[friendlyName(k)] = v2Schema
						continue
					}
				}

				schemaDefs[friendlyName(k)] = v.Schema
			}

			data, err := json.Marshal(&spec.Swagger{
				SwaggerProps: spec.SwaggerProps{
					Definitions: schemaDefs,
					Info: &spec.Info{
						InfoProps: spec.InfoProps{
							Title:   "Gateway API",
							Version: "unversioned",
						},
					},
					Swagger: "2.0",
				},
			})

			if err != nil {
				return fmt.Errorf("error serializing api definitions: %w", err)
			}

			os.Stdout.Write(data)

			return nil
		},
	}

	return &cmd
}

// From k8s.io/apiserver/pkg/endpoints/openapi/openapi.go.
func friendlyName(name string) string {
	nameParts := strings.Split(name, "/")

	// Reverse first part. e.g., io.k8s... instead of k8s.io...
	if len(nameParts) > 0 && strings.Contains(nameParts[0], ".") {
		parts := strings.Split(nameParts[0], ".")
		for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
			parts[i], parts[j] = parts[j], parts[i]
		}

		nameParts[0] = strings.Join(parts, ".")
	}

	return strings.Join(nameParts, ".")
}
