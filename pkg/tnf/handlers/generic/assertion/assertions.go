// Copyright (C) 2020 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package assertion

import (
	"encoding/json"
	"fmt"

	"github.com/test-network-function/test-network-function/pkg/tnf/handlers/generic/condition"
)

// Assertions provides the ability to compose BooleanLogic claims across any number of Assertion instances.
type Assertions struct {
	// Assertions provides a mechanism to define an arbitrary-length array of Assertion objects.
	Assertions []Assertion `json:"assertions,omitempty" yaml:"assertions,omitempty"`
	// Logic is the BooleanLogic implementation to that is asserted over Assertions.
	Logic *BooleanLogic `json:"logic,omitempty" yaml:"logic,omitempty"`
}

// unmarshalAssertionsJSON is a helper method to Unmarshal Assertions as Raw JSON.  This is needed in order to perform
// custom introspection and make decisions on how to json.Unmarshal the payload from the findings.
func (a *Assertions) unmarshalAssertionsJSON(objMap map[string]*json.RawMessage) error {
	if jsonMessage, ok := objMap[AssertionsKey]; ok {
		var assertions []Assertion
		if err := json.Unmarshal(*jsonMessage, &assertions); err != nil {
			return err
		}
		a.Assertions = assertions
	}
	return nil
}

// UnmarshalJSON deserializes a Assertions.
func (a *Assertions) UnmarshalJSON(b []byte) error {
	// Flatten out the messages to Raw JSON
	var objMap map[string]*json.RawMessage
	if err := json.Unmarshal(b, &objMap); err != nil {
		return err
	}

	// Unmarshall the assertions Array.
	if err := a.unmarshalAssertionsJSON(objMap); err != nil {
		return err
	}

	// Determine the boolean logic by introspecting the type and applying the correct Unmarshal strategy.
	logicMap, err := extractLogicMap(objMap)
	if err != nil {
		return err
	}

	typ, err := determineLogicType(logicMap)
	if err != nil {
		return err
	}

	switch typ {
	case AndBooleanLogicKey:
		if err := a.unmarshalAndBooleanLogic(objMap); err != nil {
			return err
		}
	case OrBooleanLogicKey:
		if err := a.unmarshalOrBooleanLogic(objMap); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown boolean logic type: %s", typ)
	}

	return nil
}

// unmarshalAndBooleanLogic is a strategy used to json.Unmarshal AndBooleanLogic.
func (a *Assertions) unmarshalAndBooleanLogic(objMap map[string]*json.RawMessage) error {
	var abl AndBooleanLogic
	if err := json.Unmarshal(*objMap[LogicKey], &abl); err != nil {
		return err
	}
	var bl BooleanLogic = abl
	a.Logic = &bl
	return nil
}

// unmarshalOrBooleanLogic is a strategy used to json.Unmarshal OrBooleanLogic.
func (a *Assertions) unmarshalOrBooleanLogic(objMap map[string]*json.RawMessage) error {
	var obl OrBooleanLogic
	if err := json.Unmarshal(*objMap[LogicKey], &obl); err != nil {
		return err
	}
	var bl BooleanLogic = obl
	a.Logic = &bl
	return nil
}

// extractLogicMap is used to json.Unmarshal BooleanLogic.
func extractLogicMap(objMap map[string]*json.RawMessage) (map[string]*json.RawMessage, error) {
	if logicJSONMessage, ok := objMap[LogicKey]; ok {
		var l map[string]*json.RawMessage
		if err := json.Unmarshal(*logicJSONMessage, &l); err != nil {
			return nil, err
		}
		return l, nil
	}
	return nil, fmt.Errorf("mandatory \"%s\" key is missing", LogicKey)
}

// determineLogicType introspects the raw JSON to determine what type of BooleanLogic is expressed.
func determineLogicType(logicMap map[string]*json.RawMessage) (string, error) {
	if typeJSONMessage, ok := logicMap[condition.TypeKey]; ok {
		var typ string
		if err := json.Unmarshal(*typeJSONMessage, &typ); err != nil {
			return "", err
		}
		return typ, nil
	}
	return "", fmt.Errorf("mandatory \"%s\" key is missing", TypeKey)
}
