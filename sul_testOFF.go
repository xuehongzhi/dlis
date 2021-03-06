package dlis

import (
	"strings"
	"testing"
)

var sultxt = "   1V1.00RECORD 8192Default Storage Set                                         "
var sul SULT

func TestRead(t *testing.T) {
	err := sul.Read(strings.NewReader(sultxt))
	if err != nil {
		t.Error(err)
	}
	t.Log(sul)
}
