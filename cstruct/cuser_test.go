package cstruct

import (
	trans "Cloud/transmit"
	"testing"
)

func TestNewCuser(t *testing.T) {
	u := NewCUser("default", ".")
	if u == nil {
		t.Errorf("CUser: NewCuser function failed")
	}
}

func TestGetUsername(t *testing.T) {
	u := NewCUser("default", ".")
	if u == nil || u.GetUsername() != "default" {
		t.Errorf("CUser: GetUsername function failed")
	}
}

func TestGetId(t *testing.T) {
	// TODO
	u := NewCUser("default", ".")
	if u == nil || u.GetId() != 1 {
		t.Errorf("CUser: GetId function failed")
	}
}

func TestGetToken(t *testing.T) {
	// TODO
	u := NewCUser("default", ".")
	if u == nil || u.GetToken() != "" {
		t.Errorf("CUser: GetToken function failed")
	}
}

func TestVerify(t *testing.T) {
	// TODO
	u := NewCUser("default", ".")
	if u == nil || u.Verify("") != true {
		t.Errorf("CUser: GetToken function failed")
	}
}

func TestAddUFile(t *testing.T) {
	u := NewCUser("default", ".")
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

func TestAddTransmitter(t *testing.T) {
	u := NewCUser("default", ".")
	c := trans.NewTransmitter(nil, 0, nil)
	if u.AddTransmit(c) != true {
		t.Errorf("CUser: AddTransmitter function failed")
	}
}

func TestRemoveTransmitter(t *testing.T) {
	u := NewCUser("default", ".")
	c := trans.NewTransmitter(nil, 0, nil)
	if u.AddTransmit(c) != true || u.RemoveTransmit(c) != true {
		t.Errorf("CUser: RemoveTransmitter function failed")
	}
}
