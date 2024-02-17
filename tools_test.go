package toolkit

import "testing"

func TestTools_RandomStringg(t *testing.T) {
	var testTools Tools
	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Error("Wrong Length")
	}
}
