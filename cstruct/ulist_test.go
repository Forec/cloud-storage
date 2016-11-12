/*
author: Forec
last edit date: 2016/11/13
email: forec@bupt.edu.cn
LICENSE
Copyright (c) 2015-2017, Forec <forec@bupt.edu.cn>

Permission to use, copy, modify, and/or distribute this code for any
purpose with or without fee is hereby granted, provided that the above
copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
*/

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
