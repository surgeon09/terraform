package terraform

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"

	"github.com/hashicorp/terraform/internal/addrs"
	"github.com/hashicorp/terraform/internal/configs"
	"github.com/hashicorp/terraform/internal/instances"
	"github.com/hashicorp/terraform/internal/lang"
	"github.com/hashicorp/terraform/internal/lang/marks"
	"github.com/hashicorp/terraform/internal/plans"
	"github.com/hashicorp/terraform/internal/tfdiags"
)

type checkType int

const (
	checkInvalid               checkType = 0
	checkResourcePrecondition  checkType = 1
	checkResourcePostcondition checkType = 2
	checkOutputPrecondition    checkType = 3
)

func (c checkType) FailureSummary() string {
	switch c {
	case checkResourcePrecondition:
		return "Resource precondition failed"
	case checkResourcePostcondition:
		return "Resource postcondition failed"
	case checkOutputPrecondition:
		return "Module output value precondition failed"
	default:
		// This should not happen
		return "Failed condition for invalid check type"
	}
}

func (c checkType) RuleAddr(self addrs.Checkable, i int) string {
	container := self.String()
	switch c {
	case checkResourcePrecondition:
		return fmt.Sprintf("%s.preconditions[%d]", container, i)
	case checkResourcePostcondition:
		return fmt.Sprintf("%s.postconditions[%d]", container, i)
	case checkOutputPrecondition:
		return fmt.Sprintf("%s.preconditions[%d]", container, i)
	default:
		// This should not happen
		return fmt.Sprintf("%s.conditions[%d]", container, i)
	}
}

func (c checkType) ConditionType() plans.ConditionType {
	switch c {
	case checkResourcePrecondition:
		return plans.ResourcePrecondition
	case checkResourcePostcondition:
		return plans.ResourcePostcondition
	case checkOutputPrecondition:
		return plans.OutputPrecondition
	default:
		// This should not happen
		return plans.InvalidCondition
	}
}

// evalCheckRules ensures that all of the given check rules pass against
// the given HCL evaluation context.
//
// If any check rules produce an unknown result then they will be silently
// ignored on the assumption that the same checks will be run again later
// with fewer unknown values in the EvalContext.
//
// If any of the rules do not pass, the returned diagnostics will contain
// errors. Otherwise, it will either be empty or contain only warnings.
func evalCheckRules(typ checkType, rules []*configs.CheckRule, ctx EvalContext, self addrs.Checkable, keyData instances.RepetitionData, diagSeverity tfdiags.Severity) tfdiags.Diagnostics {
	var diags tfdiags.Diagnostics

	if len(rules) == 0 {
		// Nothing to do
		return nil
	}

	severity := diagSeverity.ToHCL()

	for i, rule := range rules {
		ruleAddr := typ.RuleAddr(self, i)

		conditionResult, ruleDiags := evalCheckRule(typ, rule, ctx, self, keyData, severity)
		diags = diags.Append(ruleDiags)
		ctx.Conditions().SetResult(ruleAddr, conditionResult)
	}

	return diags
}

func evalCheckRule(typ checkType, rule *configs.CheckRule, ctx EvalContext, self addrs.Checkable, keyData instances.RepetitionData, severity hcl.DiagnosticSeverity) (*plans.ConditionResult, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics
	const errInvalidCondition = "Invalid condition result"

	refs, moreDiags := lang.ReferencesInExpr(rule.Condition)
	diags = diags.Append(moreDiags)
	moreRefs, moreDiags := lang.ReferencesInExpr(rule.ErrorMessage)
	diags = diags.Append(moreDiags)
	refs = append(refs, moreRefs...)

	conditionResult := &plans.ConditionResult{
		Address: self,
		Unknown: true,
		Type:    typ.ConditionType(),
	}

	var selfReference addrs.Referenceable
	// Only resource postconditions can refer to self
	if typ == checkResourcePostcondition {
		switch s := self.(type) {
		case addrs.AbsResourceInstance:
			selfReference = s.Resource
		default:
			panic(fmt.Sprintf("Invalid self reference type %t", self))
		}
	}
	scope := ctx.EvaluationScope(selfReference, keyData)

	hclCtx, moreDiags := scope.EvalContext(refs)
	diags = diags.Append(moreDiags)

	result, hclDiags := rule.Condition.Value(hclCtx)
	diags = diags.Append(hclDiags)

	errorValue, errorDiags := rule.ErrorMessage.Value(hclCtx)
	diags = diags.Append(errorDiags)

	if diags.HasErrors() {
		log.Printf("[TRACE] evalCheckRule: %s: %s", typ.FailureSummary(), diags.Err().Error())
	}

	if !result.IsKnown() {
		// We'll wait until we've learned more, then.
		return conditionResult, diags
	} else {
		conditionResult.Unknown = false
	}

	if result.IsNull() {
		diags = diags.Append(&hcl.Diagnostic{
			Severity:    severity,
			Summary:     errInvalidCondition,
			Detail:      "Condition expression must return either true or false, not null.",
			Subject:     rule.Condition.Range().Ptr(),
			Expression:  rule.Condition,
			EvalContext: hclCtx,
		})
		conditionResult.Result = false
		conditionResult.ErrorMessage = "Condition expression must return either true or false, not null."
		return conditionResult, diags
	}
	var err error
	result, err = convert.Convert(result, cty.Bool)
	if err != nil {
		detail := fmt.Sprintf("Invalid condition result value: %s.", tfdiags.FormatError(err))
		diags = diags.Append(&hcl.Diagnostic{
			Severity:    severity,
			Summary:     errInvalidCondition,
			Detail:      detail,
			Subject:     rule.Condition.Range().Ptr(),
			Expression:  rule.Condition,
			EvalContext: hclCtx,
		})
		conditionResult.Result = false
		conditionResult.ErrorMessage = detail
		return conditionResult, diags
	}

	// The condition result may be marked if the expression refers to a
	// sensitive value.
	result, _ = result.Unmark()
	conditionResult.Result = result.True()

	if conditionResult.Result {
		return conditionResult, diags
	}

	var errorMessage string
	if !errorDiags.HasErrors() && errorValue.IsKnown() && !errorValue.IsNull() {
		var err error
		errorValue, err = convert.Convert(errorValue, cty.String)
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity:    severity,
				Summary:     "Invalid error message",
				Detail:      fmt.Sprintf("Unsuitable value for error message: %s.", tfdiags.FormatError(err)),
				Subject:     rule.ErrorMessage.Range().Ptr(),
				Expression:  rule.ErrorMessage,
				EvalContext: hclCtx,
			})
		} else {
			if marks.Has(errorValue, marks.Sensitive) {
				diags = diags.Append(&hcl.Diagnostic{
					Severity: severity,

					Summary: "Error message refers to sensitive values",
					Detail: `The error expression used to explain this condition refers to sensitive values. Terraform will not display the resulting message.

You can correct this by removing references to sensitive values, or by carefully using the nonsensitive() function if the expression will not reveal the sensitive data.`,

					Subject:     rule.ErrorMessage.Range().Ptr(),
					Expression:  rule.ErrorMessage,
					EvalContext: hclCtx,
				})
				errorMessage = "The error message included a sensitive value, so it will not be displayed."
			} else {
				errorMessage = strings.TrimSpace(errorValue.AsString())
			}
		}
	}
	if errorMessage == "" {
		errorMessage = "Failed to evaluate condition error message."
	}
	diags = diags.Append(&hcl.Diagnostic{
		Severity:    severity,
		Summary:     typ.FailureSummary(),
		Detail:      errorMessage,
		Subject:     rule.Condition.Range().Ptr(),
		Expression:  rule.Condition,
		EvalContext: hclCtx,
	})
	conditionResult.ErrorMessage = errorMessage
	return conditionResult, diags
}
