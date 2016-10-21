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
