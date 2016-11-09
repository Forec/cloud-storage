/*
author: Forec
last edit date: 2016/11/09
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
	"testing"
	"time"
)

func TestNewCFile(t *testing.T) {
	c := NewCFile(int64(12345), int64(1234567))
	if c == nil {
		t.Errorf("CFile: NewCFile function failed")
	}
}

func TestAddRef(t *testing.T) {
	c := NewCFile(int64(12345), int64(1234567))
	if c == nil || c.AddRef(100) != true && c.GetRef() != 100 {
		t.Errorf("CFile: AddRef function failed")
	}
}

func TestGetID(t *testing.T) {
	c := NewCFile(int64(12345), int64(1234567))
	if c.GetId() != 12345 {
		t.Errorf("CFile: GetId function failed")
	}
}

func TestGetSize(t *testing.T) {
	c := NewCFile(int64(12345), int64(1234567))
	if c.GetSize() != 1234567 {
		t.Errorf("CFile: GetSize function failed")
	}
}

func TestGetTimestamp(t *testing.T) {
	c := NewCFile(int64(12345), int64(1234567))
	if time.Now().Sub(c.GetTimestamp()) > time.Second {
		t.Errorf("CFile: GetTimestamp function failed")
	}
}

func TestSetId(t *testing.T) {
	c := NewCFile(int64(12345), int64(1234567))
	if c.SetId(777) != true || c.GetId() != 777 {
		t.Errorf("CFile: SetId function failed")
	}
}

func TestSetSize(t *testing.T) {
	c := NewCFile(int64(12345), int64(1234567))
	if c.SetSize(777) != true || c.GetSize() != 777 {
		t.Errorf("CFile: SetSize function failed")
	}
}
