//go:build integration

package app_test

import (
	"fmt"
	"testing"
)

func TestMytest(t *testing.T) {
	fmt.Println("My integration test running!")
	got := 3
	if got != 1 {
		t.Errorf("var = %d; want 1", got)
	}
}
