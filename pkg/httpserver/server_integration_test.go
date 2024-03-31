//go:build integration

package httpserver_test

import (
	"fmt"
	"testing"
)

func TestMytest(t *testing.T) {
	fmt.Println("My integration test running!")
	got := 5
	if got != 1 {
		t.Errorf("var = %d; want 1", got)
	}
}
