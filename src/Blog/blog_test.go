package blog

import (
	"testing"
)

func TestRound(t *testing.T) {
	var tests = [...]int64{123456, 1234, 123, 12, 1, 0, -1}
	var wants = [...]int64{123000, 1230, 123, 12, 1, 0, -1}
	for i, test := range tests {
		if r := round(test); r != wants[i] {
			t.Errorf("\nTestRound(%d) want: %d, But got: %d", test, wants[i], r)
		}
	}
}
