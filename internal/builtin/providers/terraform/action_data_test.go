// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package terraform

import (
	"testing"

	"github.com/hashicorp/terraform/internal/providers"
	"github.com/zclconf/go-cty/cty"
)

func TestProviderTerraformDataActionIntegration(t *testing.T) {
	p := NewProvider()

	// Test that the schema is registered
	schema := p.GetProviderSchema()
	actionSchema, exists := schema.Actions["terraform_data"]
	if !exists {
		t.Fatal("expected 'terraform_data' action to be registered in provider schema")
	}

	if actionSchema.ConfigSchema == nil {
		t.Fatal("expected action schema to have ConfigSchema")
	}

	// Test with string input
	t.Run("string input", func(t *testing.T) {
		config := cty.ObjectVal(map[string]cty.Value{
			"input": cty.StringVal("test string"),
		})

		// Test ValidateActionConfig
		validateResp := p.ValidateActionConfig(providers.ValidateActionConfigRequest{
			TypeName: "terraform_data",
			Config:   config,
		})

		if validateResp.Diagnostics.HasErrors() {
			t.Errorf("unexpected validation errors: %s", validateResp.Diagnostics.Err())
		}

		// Test PlanAction
		planResp := p.PlanAction(providers.PlanActionRequest{
			ActionType:         "terraform_data",
			ProposedActionData: config,
		})

		if planResp.Diagnostics.HasErrors() {
			t.Errorf("unexpected plan errors: %s", planResp.Diagnostics.Err())
		}

		// Test InvokeAction
		invokeResp := p.InvokeAction(providers.InvokeActionRequest{
			ActionType:        "terraform_data",
			PlannedActionData: config,
		})

		if invokeResp.Diagnostics.HasErrors() {
			t.Errorf("unexpected invoke errors: %s", invokeResp.Diagnostics.Err())
		}

		// Verify we got events
		eventCount := 0
		for range invokeResp.Events {
			eventCount++
		}

		if eventCount == 0 {
			t.Error("expected to receive events from InvokeAction")
		}
	})

	// Test with map input
	t.Run("map input", func(t *testing.T) {
		config := cty.ObjectVal(map[string]cty.Value{
			"input": cty.ObjectVal(map[string]cty.Value{
				"key1": cty.StringVal("value1"),
				"key2": cty.NumberIntVal(42),
			}),
		})

		invokeResp := p.InvokeAction(providers.InvokeActionRequest{
			ActionType:        "terraform_data",
			PlannedActionData: config,
		})

		if invokeResp.Diagnostics.HasErrors() {
			t.Errorf("unexpected invoke errors: %s", invokeResp.Diagnostics.Err())
		}

		hasInput := false
		hasOutput := false
		for event := range invokeResp.Events {
			if prog, ok := event.(providers.InvokeActionEvent_Progress); ok {
				if len(prog.Message) > 0 {
					if len(prog.Message) >= 6 && prog.Message[:6] == "input:" {
						hasInput = true
					}
					if len(prog.Message) >= 7 && prog.Message[:7] == "output:" {
						hasOutput = true
					}
				}
			}
		}

		if !hasInput {
			t.Error("expected to see input message in progress events")
		}
		if !hasOutput {
			t.Error("expected to see output message in progress events")
		}
	})

	// Test with null input
	t.Run("null input", func(t *testing.T) {
		config := cty.ObjectVal(map[string]cty.Value{
			"input": cty.NullVal(cty.DynamicPseudoType),
		})

		invokeResp := p.InvokeAction(providers.InvokeActionRequest{
			ActionType:        "terraform_data",
			PlannedActionData: config,
		})

		if invokeResp.Diagnostics.HasErrors() {
			t.Errorf("unexpected invoke errors: %s", invokeResp.Diagnostics.Err())
		}

		eventCount := 0
		for range invokeResp.Events {
			eventCount++
		}

		if eventCount == 0 {
			t.Error("expected to receive events from InvokeAction")
		}
	})

	// Test with empty config
	t.Run("empty config", func(t *testing.T) {
		config := cty.ObjectVal(map[string]cty.Value{})

		invokeResp := p.InvokeAction(providers.InvokeActionRequest{
			ActionType:        "terraform_data",
			PlannedActionData: config,
		})

		if invokeResp.Diagnostics.HasErrors() {
			t.Errorf("unexpected invoke errors: %s", invokeResp.Diagnostics.Err())
		}

		eventCount := 0
		for range invokeResp.Events {
			eventCount++
		}

		if eventCount == 0 {
			t.Error("expected to receive events from InvokeAction")
		}
	})
}
