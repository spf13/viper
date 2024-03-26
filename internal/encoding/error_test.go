package encoding

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_encodingError(t *testing.T) {
	err1 := fmt.Errorf("test error")
	err2 := encodingError("encoding error")
	assert.NotErrorIs(t, err1, err2)
	assert.NotErrorIs(t, err2, err1)
	assert.ErrorIs(t, err2, encodingError("encoding error"))
	assert.NotErrorIs(t, err2, encodingError("other encodingerror"))
}
