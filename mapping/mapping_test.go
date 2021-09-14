package mapping

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapping(t *testing.T) {
	Init(".")

	fmt.Println(len(mapping))
	assert.NotEqual(t, 0, len(mapping))

	assert.Equal(t, float64(1), mapping["city"]["beijing"])
	assert.Equal(t, float64(0), mapping["city"][""])
}
