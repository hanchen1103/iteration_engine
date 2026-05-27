package engine

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hanchen1103/iteration_engine/domain"
)

func validateTarget(spec domain.SceneSpec, target domain.TargetRef) error {
	if strings.TrimSpace(target.Type) == "" {
		return invalidError("target.type is required")
	}
	if strings.TrimSpace(target.ID) == "" {
		return invalidError("target.id is required")
	}
	if spec.TargetType != "" && target.Type != spec.TargetType {
		return invalidError(fmt.Sprintf("target type %q does not match scene target type %q", target.Type, spec.TargetType))
	}
	return nil
}

func validatePlan(plan domain.IterationPlan) error {
	if strings.TrimSpace(string(plan.Source)) == "" {
		return invalidError("iteration plan source is required")
	}
	return validateDirectives(plan.Directives)
}

func validateDirectives(directives []domain.IterationDirective) error {
	seen := map[string]struct{}{}
	for _, directive := range directives {
		key := strings.TrimSpace(directive.Key)
		if key == "" {
			return invalidError("directive key is required")
		}
		if _, ok := seen[key]; ok {
			return invalidError("duplicate directive key: " + key)
		}
		seen[key] = struct{}{}
		if strings.TrimSpace(directive.Description) == "" {
			return invalidError("directive description is required")
		}
		if len(directive.Value) == 0 {
			return invalidError("directive value is required")
		}
		var object map[string]any
		if err := json.Unmarshal(directive.Value, &object); err != nil || object == nil {
			return invalidError("directive value must be a JSON object")
		}
	}
	return nil
}
