package common

import (
	"testing"
)

func TestHexToColor(t *testing.T) {
	str := "#cc0033"
	c, err := HexToColor(str)
	if err != nil {
		t.Error(err)
	}
	t.Log(c)
}
