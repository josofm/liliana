//go:build unit

package app_test

import (
	"fmt"
	"testing"
)

func TestMytest(t *testing.T) {
	fmt.Println("My unit test running!")
	got := 1
	if got != 1 {
		t.Errorf("var = %d; want 1", got)
	}
}
