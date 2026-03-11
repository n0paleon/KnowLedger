package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripHTML(t *testing.T) {
	content := `Goodbye <a onblur="alert(secret)" href="https://en.wikipedia.org/wiki/Goodbye_Cruel_World_(Pink_Floyd_song)">Cruel</a> World`
	expectedOutput := "Goodbye Cruel World"

	output := StripHTML(content)
	assert.Equal(t, expectedOutput, output)
}
