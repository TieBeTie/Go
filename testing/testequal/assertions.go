//go:build !solution

package testequal

import (
	"bytes"
	"fmt"
	"reflect"
	"slices"
)

// AssertEqual checks that expected and actual are equal.
//
// Marks caller function as having failed but continues execution.
//
// Returns true iff arguments are equal.
func AssertEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()
	var msg string
	if len(msgAndArgs) > 0 {
		msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	} else {
		msg = fmt.Sprintf("expected: %v, actual: %v", expected, actual)
	}

	// Get types
	expType := reflect.TypeOf(expected)
	actType := reflect.TypeOf(actual)

	// Handle different types
	if expType != actType {
		t.Errorf(msg)
		return false
	}

	switch exp := expected.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		if exp != actual {
			t.Errorf(msg)
			return false
		}
	case string:
		if exp != actual {
			t.Errorf(msg)
			return false
		}
	case []int:
		act, ok := actual.([]int)
		if !ok {
			t.Errorf(msg)
			return false
		}
		if len(exp) == 0 && len(act) == 0 {
			t.Errorf(msg)
			return false // Empty slices should be considered not equal
		}
		if !slices.Equal(exp, act) {
			t.Errorf(msg)
			return false
		}
	case []byte:
		act, ok := actual.([]byte)
		if !ok {
			t.Errorf(msg)
			return false
		}
		if len(exp) == 0 && len(act) == 0 {
			t.Errorf(msg)
			return false // Empty slices should be considered not equal
		}
		if !bytes.Equal(exp, act) {
			t.Errorf(msg)
			return false
		}
	case map[string]string:
		act, ok := actual.(map[string]string)
		if !ok || !reflect.DeepEqual(exp, act) {
			t.Errorf(msg)
			return false
		}
	default:
		// For unsupported types (empty slices, structs, etc), consider them not equal
		t.Errorf(msg)
		return false
	}
	return true
}

// AssertNotEqual checks that expected and actual are not equal.
//
// Marks caller function as having failed but continues execution.
//
// Returns true iff arguments are not equal.
func AssertNotEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()
	var msg string
	if len(msgAndArgs) > 0 {
		msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	} else {
		msg = fmt.Sprintf("expected: %v, actual: %v", expected, actual)
	}

	// Handle nil cases
	if expected == nil || actual == nil {
		if expected == actual {
			t.Errorf(msg)
			return false
		}
		return true
	}

	// Get types
	expType := reflect.TypeOf(expected)
	actType := reflect.TypeOf(actual)

	// Handle different types
	if expType != actType {
		return true
	}

	switch exp := expected.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		if exp == actual {
			t.Errorf(msg)
			return false
		}
	case string:
		if exp == actual {
			t.Errorf(msg)
			return false
		}
	case []int:
		act, ok := actual.([]int)
		if !ok {
			return true
		}
		if len(exp) == 0 && len(act) == 0 {
			return true // Empty slices should be considered not equal
		}
		if slices.Equal(exp, act) {
			t.Errorf(msg)
			return false
		}
	case []byte:
		act, ok := actual.([]byte)
		if !ok {
			return true
		}
		if len(exp) == 0 && len(act) == 0 {
			return true // Empty slices should be considered not equal
		}
		if bytes.Equal(exp, act) {
			t.Errorf(msg)
			return false
		}
	case map[string]string:
		act, ok := actual.(map[string]string)
		if !ok {
			return true
		}
		if reflect.DeepEqual(exp, act) {
			t.Errorf(msg)
			return false
		}
	default:
		// For unsupported types (empty slices, structs, etc), consider them not equal
		return true
	}
	return true
}

// RequireEqual does the same as AssertEqual but fails caller test immediately.
func RequireEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if !AssertEqual(t, expected, actual, msgAndArgs...) {
		t.FailNow()
	}
}

// RequireNotEqual does the same as AssertNotEqual but fails caller test immediately.
func RequireNotEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if !AssertNotEqual(t, expected, actual, msgAndArgs...) {
		t.FailNow()
	}
}
