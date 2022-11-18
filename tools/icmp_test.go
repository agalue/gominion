package tools

import (
	"fmt"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestPing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestPing")
	}
	duration, err := Ping("8.8.8.8", 2*time.Second)
	assert.NilError(t, err)
	fmt.Printf("Duration %d microseconds\n", duration.Microseconds())
	assert.Assert(t, duration.Microseconds() > 0)
}
