//go:build integration

package v1

import "testing"

// checkErr is a helper function to handle errors in tests
func checkErr(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
