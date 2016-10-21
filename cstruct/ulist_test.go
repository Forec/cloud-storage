package cstruct

import (
	"testing"
)

func TestAppendElements(t *testing.T) {
	testUFSlice := make([]*ufile, 0, 10)
	for i := 0; i < 20; i++ {
		uf := NewUFile(NewCFile(int64(i), int64(i+12345)), nil, "test1", "../files/test")
		testUFSlice = AppendElements(testUFSlice, uf)
	}
	if len(testUFSlice) != 20 {
		t.Errorf("UList: AppendElements function failed (single)")
	}
	testUFSlice = AppendElements(testUFSlice, testUFSlice...)
	if len(testUFSlice) != 40 {
		t.Errorf("UList: AppendElements function failed (multiple)")
	}
}

func TestIndexById(t *testing.T) {
	testUFSlice := make([]*ufile, 0, 10)
	f := NewCFile(int64(0), int64(12345))
	for i := 0; i < 20; i++ {
		uf := NewUFile(f, nil, "test1", "../files/test")
		testUFSlice = AppendElements(testUFSlice, uf)
	}
	testIndexSlice := IndexById(testUFSlice, int64(0))
	if testIndexSlice == nil || len(testIndexSlice) != 20 {
		t.Errorf("UList: IndexById function failed, %d", len(testIndexSlice))
	}
}
