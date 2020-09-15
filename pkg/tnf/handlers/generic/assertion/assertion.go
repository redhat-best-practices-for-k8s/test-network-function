package assertion

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/generic/condition"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/generic/condition/intcondition"
	"github.com/redhat-nfvpe/test-network-function/pkg/tnf/handlers/generic/condition/stringcondition"
)

const (
	// AssertionsKey is the JSON key which indicates an assertions array payload.
	AssertionsKey = "assertions"
	// ConditionKey is the JSON key which indicates a condition payload.
	ConditionKey = "condition"
	// GroupIndexKey is the JSON key which represents a regular expression group index.
	GroupIndexKey = "groupIdx"
)

// Assertion provides the ability to assert a Condition for the string extracted from GroupIdx of the parent
// ResultContext Pattern
type Assertion struct {
	// GroupIdx is the index in the match string used in the Assertion.
	GroupIdx int `json:"groupIdx" yaml:"groupIdx"`
	// Condition is the condition.Condition asserted in this Assertion.
	Condition *condition.Condition `json:"condition" yaml:"condition"`
}

// unmarshalGroupIdxJSON is a helper function used to json.Unmarshal an Assertion.  The implementation is meant to mock
// the default implementation, and is only made necessary since other Assertion elements require special treatment.
func (a *Assertion) unmarshalGroupIdxJSON(objMap map[string]*json.RawMessage) error {
	if groupIdxJSONMessage, ok := objMap[GroupIndexKey]; ok {
		var groupIdx int
		if err := json.Unmarshal(*groupIdxJSONMessage, &groupIdx); err != nil {
			return err
		}
		a.GroupIdx = groupIdx
		return nil
	}
	return fmt.Errorf("required field \"%s\" is missing from the JSON payload", GroupIndexKey)
}

// unmarshalConditionTypeJSON is a helper function used to introspect on a condition.Condition payload.  Since there are
// many implementations of condition.Condition interface, custom strategy must be used to Unmarshal each implementation.
// unmarshalConditionTypeJSON returns the type (if it exists) and any encountered error.
func unmarshalConditionTypeJSON(conditionObjMap map[string]*json.RawMessage) (string, error) {
	if typJSONMessage, ok := conditionObjMap[condition.TypeKey]; ok {
		var typ string
		if err := json.Unmarshal(*typJSONMessage, &typ); err != nil {
			return "", err
		}
		return typ, nil
	}
	return "", fmt.Errorf("condition missing \"%s\"", condition.TypeKey)
}

// unmarshalEqualsCondition is a custom strategy used to json.Unmarshal an Assertion utilizing
// condition.EqualsCondition.
func (a *Assertion) unmarshalEqualsCondition(conditionJSONMessage *json.RawMessage) error {
	var equalsCondition stringcondition.EqualsCondition
	if err := json.Unmarshal(*conditionJSONMessage, &equalsCondition); err != nil {
		return err
	}
	var cond condition.Condition = equalsCondition
	a.Condition = &cond
	return nil
}

// unmarshalIsIntCondition is a custom strategy used to json.Unmarshal an Assertion utilizing
// condition.IsIntCondition.
func (a *Assertion) unmarshalIsIntCondition(conditionJSONMessage *json.RawMessage) error {
	var isIntCondition intcondition.IsIntCondition
	if err := json.Unmarshal(*conditionJSONMessage, &isIntCondition); err != nil {
		return err
	}
	var cond condition.Condition = isIntCondition
	a.Condition = &cond
	return nil
}

// unmarshalIntComparisonCondition is a custom strategy used to json.Unmarshal an Assertion utilizing
// condition.ComparisonCondition.
func (a *Assertion) unmarshalIntComparisonCondition(conditionJSONMessage *json.RawMessage) error {
	var intComparisonCondition intcondition.ComparisonCondition
	if err := json.Unmarshal(*conditionJSONMessage, &intComparisonCondition); err != nil {
		return err
	}
	var cond condition.Condition = intComparisonCondition
	a.Condition = &cond
	return nil
}

// unmarshalConditionJSON is a custom strategy used to json.Unmarshal an Assertion utilizing
// any known condition.Condition.
func (a *Assertion) unmarshalConditionJSON(objMap map[string]*json.RawMessage) error {
	if conditionJSONMessage, ok := objMap[ConditionKey]; ok {
		var conditionObjMap map[string]*json.RawMessage
		if err := json.Unmarshal(*conditionJSONMessage, &conditionObjMap); err != nil {
			return err
		}

		// Introspect the type of condition prior to attempting to Unmarshal.  This is necessary since JSON does
		// understand Polymorphism at the level of GoLang.
		typ, err := unmarshalConditionTypeJSON(conditionObjMap)
		if err != nil {
			return err
		}
		switch typ {
		case stringcondition.EqualsConditionKey:
			if err = a.unmarshalEqualsCondition(conditionJSONMessage); err != nil {
				return err
			}
		case intcondition.IsIntConditionKey:
			if err = a.unmarshalIsIntCondition(conditionJSONMessage); err != nil {
				return err
			}
		case intcondition.ComparisonConditionKey:
			if err = a.unmarshalIntComparisonCondition(conditionJSONMessage); err != nil {
				return err
			}
		default:
			return errors.New(fmt.Sprintf("unrecognized condition type: \"%s\"", typ))
		}
	}
	return nil
}

// UnmarshalJSON deserializes an Assertion.
func (a *Assertion) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		return err
	}

	if err = a.unmarshalGroupIdxJSON(objMap); err != nil {
		return err
	}

	if err = a.unmarshalConditionJSON(objMap); err != nil {
		return err
	}

	return nil
}
