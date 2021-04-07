package blog

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func TestNewHeadFile(t *testing.T) {
	if f := NewHeadFile("./testdata/head"); f == nil {
		t.Errorf("CreateNewFile(%#v) get nil!", "head")
	} else {
		f.Close()
	}
}

func TestHead(t *testing.T) {
	if HeadInit("./testdata/head"); baseHead.fd == nil {
		t.Errorf("TestWrite of NewHeadFile(%#v) get nil!", "head")
	}

	if r := baseHead.WriteString("test1"); r != 0 {
		t.Errorf("TestWrite of Write(%#v) get r!=0 !", "test1")
	}

	if r := baseHead.WriteString("test2"); r == 0 {
		t.Errorf("TestWrite of Write(%#v) get r=0!", "test2")
	}

	HeadClose()

	if b, err := ioutil.ReadFile("./testdata/head"); err != nil {
		t.Errorf("TestWrite of ReadFile(%#v) get r=0!", "./testdata/head")
	} else {
		if string(b[0:5]) != "test1" || string(b[6:11]) != "test2" {
			t.Errorf("TestWrite check String(%#v) get %#v, %#v!", "test1test2", string(b[0:5]), string(b[5:11]))
		}
	}
}

func TestBody(t *testing.T) {
	if BodyInit("./testdata/body"); baseBody.fd == nil {
		t.Errorf("TestBody of NewBodyFile(%#v) get nil!", "body")
		return
	}

	var tests = []BodySubject{
		{
			Status: 8,
			Signal: 3,
			Pid:    7,
			Time:   987654321,
			Cwd:    0x1234,
			Cmd:    0x9876,
			Env:    []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			Args:   []string{"a", "ab", "abc", "abcd"},
		},
		{
			Status: 3,
			Signal: 15,
			Pid:    8,
			Time:   123456789,
			Cwd:    0x4321,
			Cmd:    0x5678,
			Env:    []uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			Args:   []string{"z", "zy", "zyx", "zyxw"},
		},
		{
			Status: 4,
			Signal: 5,
			Pid:    6,
			Time:   7,
			Cwd:    8,
			Cmd:    9,
			Env:    []uint32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			Args:   []string{""},
		},
		{
			Env:  []uint32{},
			Args: []string{""},
		},
		{
			Status: 4,
			Signal: 5,
			Pid:    6,
			Time:   7,
			Cwd:    8,
			Cmd:    9,
			Env:    []uint32{},
			Args:   []string{"z", "zy", "zyx", "zyxw"},
		},
		{
			Status: 99,
			Signal: 3,
			Pid:    7,
			Time:   987654321,
			Cwd:    0x1234,
			Cmd:    0x9876,
			Env:    []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			Args:   []string{"a", "ab", "abc", "abcd"},
		},
	}
	var offsets []int64
	for _, v := range tests {
		v := v
		offsets = append(offsets, baseBody.WriteSubject(&v))
	}
	BodyClose()
	if b, err := ioutil.ReadFile("./testdata/body"); err != nil {
		t.Errorf("TestBody of ReadFile(%#v) get r=0!", "./testdata/body")
	} else {

		lenth := len(offsets)
		for i := lenth - 1; i >= 0; i-- {
			got := baseBody.ReadSubject(b[offsets[i]:])
			if !reflect.DeepEqual(*got, tests[i]) {
				t.Errorf("TestBody want %#v, but get %#v", tests[i], got)
			}
		}
	}

}
