package blog

import (
	"os"
	"testing"
)

func TestDir(t *testing.T) {
	if !IsDir("/tmp/") {
		t.Errorf("\nTestDir want: %v,\nBut got:%v", true, false)
	}
}

func TestMd5(t *testing.T) {
	want := "d41d8cd98f00b204e9800998ecf8427e"
	got, _ := Md5("./testdata/makelog.dat2")
	if got != string(want) {
		t.Errorf("\nTestMd5 want: %s,\nBut got:%s", want, got)
	}
}

func TestGz(t *testing.T) {
	Gz("./testdata/makelog.dat", "./testdata/makelog.gz")
	Ungz("./testdata/makelog.gz", "./testdata/makelog.dat2")
	gz, _ := Md5("./testdata/makelog.dat")
	ungz, _ := Md5("./testdata/makelog.dat2")
	if gz != ungz {
		t.Errorf("\nTestGz want: %s,\nBut got:%s", gz, ungz)
	} else {
		os.Remove("./testdata/makelog.dat2")
		os.Remove("./testdata/makelog.gz")
	}
}
