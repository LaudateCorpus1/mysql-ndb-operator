// Copyright (c) 2021, Oracle and/or its affiliates.
//
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/

package ndberrors

import (
	"errors"
	"fmt"
	"testing"
)

func Test_NdbError(t *testing.T) {

	err := error(&NdbError{code: 24, details: "test error occurred"})
	ndberr := &NdbError{}

	if errors.As(err, &ndberr) {
		fmt.Printf("Worked %s\n", ndberr)
	} else {
		t.Fail()
	}
}

// Test_AllErrors just makes sure all error types are correctly
// created and detected
func Test_AllErrors(t *testing.T) {

	type testS struct {
		reason string
		fp     func(error) bool
	}
	errorTests := []testS{
		{ErrReasonInvalidConfiguration, IsInvalidConfiguration},
		{ErrReasonNoManagementServerConnection, IsNoManagementServerConnection},
	}

	for _, errTest := range errorTests {
		err := &NdbError{code: 24, reason: errTest.reason}
		if !errTest.fp(err) {
			t.Errorf("%s wrongly detected", errTest.reason)
		}
	}

	// any random should give that
	bogusReason := " asdasd "
	ndbErr := &NdbError{reason: bogusReason}

	if getReason(ndbErr) != bogusReason {
		t.Errorf("Random error not detected as such")
	}

	// should not be detected as any of the known
	for _, errTest := range errorTests {
		if errTest.fp(ndbErr) {
			t.Errorf("Bogus reason wrongly detected as %s", errTest.reason)
		}
	}

	ndbErr = NewErrorNoManagementServerConnection("Test message")
	if ndbErr.reason != ErrReasonNoManagementServerConnection {
		t.Errorf("%s wrongly created as %s", ErrReasonNoManagementServerConnection, ndbErr.reason)
	}

}
