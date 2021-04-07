package olog

import (
	"testing"
)

func TestOlog(t *testing.T) {
	BuildLogMain("./testdata/makelog.dat", false, "./testdata/makelog.log")
}
