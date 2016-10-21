package cstruct

import (
	"testing"
)

func TestNewCuser(t *testing.T) {
	u := NewCUser(int64(0), "default", ".")
	if u == nil {
		t.Errorf("CUser: NewCuser function failed")
	}
}

func TestGetUsername(t *testing.T) {
	u := NewCUser(int64(0), "default", ".")
	if u == nil || u.GetUsername() != "default" {
		t.Errorf("CUser: GetUsername function failed")
	}
}

func TestGetId(t *testing.T) {
	u := NewCUser(int64(12345), "default", ".")
	if u == nil || u.GetId() != 12345 {
		t.Errorf("CUser: GetId function failed")
	}
}

func TestGetToken(t *testing.T) {
	// TODO
	u := NewCUser(int64(12345), "default", ".")
	if u == nil || u.GetId() != 12345 {
		t.Errorf("CUser: GetToken function failed")
	}
}

func TestVerify(t *testing.T) {
	// TODO
	u := NewCUser(int64(12345), "default", ".")
	if u == nil || u.Verify("") != true {
		t.Errorf("CUser: GetToken function failed")
	}
}

func TestAddUFile(t *testing.T) {
	u := NewCUser(int64(12345), "default", ".")
	c := NewCFile(int64(10086), int64(65536))
	uf := NewUFile(c, u, "userfile", "./user/files/")
	if u.AddUFile(uf) != true {
		t.Errorf("CUser: AddUFile function failed")
	}
}

func TestRemoveUFile(t *testing.T) {
	u := NewCUser(int64(12345), "default", ".")
	c := NewCFile(int64(10086), int64(65536))
	uf := NewUFile(c, u, "userfile", "./user/files/")
	if u.AddUFile(uf) != true || u.RemoveUFile(uf) != true {
		t.Errorf("CUser: RemoveUFile function failed")
	}
}
