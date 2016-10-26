package cstruct

import (
	trans "Cloud/transmit"
	"testing"
)

func TestAppendElements(t *testing.T) {
	testUFSlice := make([]UFile, 0, 10)
	for i := 0; i < 20; i++ {
		uf := NewUFile(NewCFile(int64(i), int64(i+12345)), nil, "test1", "../files/test")
		testUFSlice = AppendUFile(testUFSlice, uf)
	}
	if len(testUFSlice) != 20 {
		t.Errorf("UList: AppendUFile function failed (single)")
	}
	testUFSlice = AppendUFile(testUFSlice, testUFSlice...)
	if len(testUFSlice) != 40 {
		t.Errorf("UList: AppendUFile function failed (multiple)")
	}
}

func TestAppendTransmitable(t *testing.T) {
	testTransmitableSlice := make([]trans.Transmitable, 19, 19)
	ts1 := trans.NewTransmitter(nil, 0, nil)
	ts2 := trans.NewTransmitter(nil, 0, nil)
	testTransmitableSlice = AppendTransmitable(testTransmitableSlice, ts1)
	if len(testTransmitableSlice) != 20 {
		t.Errorf("UList: AppendTransmitable function failed expected 20, got %d",
			len(testTransmitableSlice))
	}
	testTransmitableSlice = AppendTransmitable(testTransmitableSlice, ts2)
	if len(testTransmitableSlice) != 20 {
		t.Errorf("UList: AppendTransmitable function failed expected 20, got %d",
			len(testTransmitableSlice))
	}
}

func TestUFileIndexById(t *testing.T) {
	testUFSlice := make([]UFile, 0, 10)
	f := NewCFile(int64(0), int64(12345))
	for i := 0; i < 20; i++ {
		uf := NewUFile(f, nil, "test1", "../files/test")
		testUFSlice = AppendUFile(testUFSlice, uf)
	}
	testIndexSlice := UFileIndexById(testUFSlice, int64(0))
	if testIndexSlice == nil || len(testIndexSlice) != 20 {
		t.Errorf("UList: IndexById function failed, %d", len(testIndexSlice))
	}
}
