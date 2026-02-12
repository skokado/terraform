// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package terraform

import (
	"fmt"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform/internal/configs/configschema"
	"github.com/hashicorp/terraform/internal/providers"
	"github.com/hashicorp/terraform/internal/tfdiags"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

func terraformDataActionSchema() providers.ActionSchema {
	return providers.ActionSchema{
		ConfigSchema: &configschema.Block{
			Attributes: map[string]*configschema.Attribute{
				"input": {
					Type:        cty.DynamicPseudoType,
					Optional:    true,
					Description: "A value to pass through and make available as output. Can be of any type.",
				},
			},
		},
	}
}

func validateTerraformDataActionConfig(req providers.ValidateActionConfigRequest) providers.ValidateActionConfigResponse {
	var resp providers.ValidateActionConfigResponse

	// Similar to terraform_data resource validation
	// Basic validation is handled by schema

	return resp
}

func planTerraformDataAction(req providers.PlanActionRequest) providers.PlanActionResponse {
	var resp providers.PlanActionResponse

	// terraform_data actions don't need special planning
	// They simply process input and produce output

	return resp
}

func invokeTerraformDataAction(req providers.InvokeActionRequest) providers.InvokeActionResponse {
	var resp providers.InvokeActionResponse

	// Create a channel for events
	events := make(chan providers.InvokeActionEvent)

	// Run in a goroutine to populate the iterator
	go func() {
		defer close(events)

		// Generate a unique ID for this action invocation
		idString, err := uuid.GenerateUUID()
		if err != nil {
			// Send error as completed event
			diag := tfdiags.AttributeValue(
				tfdiags.Error,
				"Error generating id",
				err.Error(),
				cty.GetAttrPath("id"),
			)
			events <- providers.InvokeActionEvent_Completed{
				Diagnostics: tfdiags.Diagnostics{diag},
			}
			return
		}

		// Extract input value
		var input cty.Value
		var output string

		if !req.PlannedActionData.IsNull() && req.PlannedActionData.Type().IsObjectType() {
			// Check if the input attribute exists
			attrTypes := req.PlannedActionData.Type().AttributeTypes()
			if _, hasInput := attrTypes["input"]; hasInput {
				input = req.PlannedActionData.GetAttr("input")
			} else {
				input = cty.NullVal(cty.DynamicPseudoType)
			}
		} else {
			input = cty.NullVal(cty.DynamicPseudoType)
		}

		// Format output based on input type
		if input.IsNull() {
			output = "null"
		} else {
			// Convert the input value to JSON representation for display
			inputJSON, err := ctyjson.Marshal(input, input.Type())
			if err != nil {
				output = fmt.Sprintf("<unable to display: %s>", err)
			} else {
				output = string(inputJSON)
			}
		}

		// Send progress event with the action details
		events <- providers.InvokeActionEvent_Progress{
			Message: fmt.Sprintf("terraform_data action invoked (id: %s)", idString),
		}

		// Send progress event with input/output information
		events <- providers.InvokeActionEvent_Progress{
			Message: fmt.Sprintf("input: %s", output),
		}

		events <- providers.InvokeActionEvent_Progress{
			Message: fmt.Sprintf("output: %s", output),
		}

		// Send completion event
		events <- providers.InvokeActionEvent_Completed{
			Diagnostics: nil,
		}
	}()

	// Convert channel to iterator
	resp.Events = func(yield func(providers.InvokeActionEvent) bool) {
		for event := range events {
			if !yield(event) {
				return
			}
		}
	}

	return resp
}
