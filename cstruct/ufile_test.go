package cstruct

import (
	"testing"
)

func TestNewUFile(t *testing.T) {
	uf := NewUFile(nil, nil, "userfile", "./user/files/")
	if uf == nil {
		t.Errorf("UFile: NewUFile function failed")
	}
}

func TestGetFilename(t *testing.T) {
	uf := NewUFile(nil, nil, "userfile", "./user/files/")
	if uf == nil || uf.GetFilename() != "userfile" {
		t.Errorf("UFile: GetFilename function failed")
	}
}

func TestGetPath(t *testing.T) {
	uf := NewUFile(nil, nil, "userfile", "./user/files/")
	if uf == nil || uf.GetPath() != "./user/files/" {
		t.Errorf("UFile: GetPath function failed")
	}
}

func TestGetShared(t *testing.T) {
	uf := NewUFile(nil, nil, "userfile", "./user/files/")
	if uf == nil || uf.GetShared() != 0 {
		t.Errorf("UFile: GetShared function failed")
	}
}

func TestGetDownloaded(t *testing.T) {
	uf := NewUFile(nil, nil, "userfile", "./user/files/")
	if uf == nil || uf.GetDownloaded() != 0 {
		t.Errorf("UFile: GetDownloaded function failed")
	}
}

func TestGetOwner(t *testing.T) {
	c := NewCUser(int64(10086), "forec", "../")
	uf := NewUFile(nil, c, "userfile", "./user/files/")
	if uf == nil || uf.GetOwner() == nil || uf.GetOwner().GetId() != c.GetId() {
		t.Errorf("UFile: GetOwner function failed")
	}
}

func TestGetPointer(t *testing.T) {
	f := NewCFile(int64(10086), int64(1000000))
	uf := NewUFile(f, nil, "userfile", "./user/files/")
	if uf == nil || uf.GetPointer() == nil || uf.GetPointer().GetId() != f.GetId() {
		t.Errorf("UFile: GetPointer function failed")
	}
}

func TestIncShared(t *testing.T) {
	f := NewCFile(int64(10086), int64(1000000))
	uf := NewUFile(f, nil, "userfile", "./user/files/")
	if uf == nil || uf.IncShared() != true || uf.GetShared() != 1 {
		t.Errorf("UFile: IncShared function failed")
	}
}

func TestIncDownloaded(t *testing.T) {
	uf := NewUFile(nil, nil, "userfile", "./user/files/")
	if uf == nil || uf.IncDowned() != true || uf.GetDownloaded() != 1 {
		t.Errorf("UFile: IncDowned function failed")
	}
}

func TestSetOwner(t *testing.T) {
	c := NewCUser(int64(10086), "forec", "../")
	uf := NewUFile(nil, nil, "userfile", "./user/files/")
	if uf == nil || uf.SetOwner(c) != true ||
		uf.GetOwner() == nil || uf.GetOwner().GetId() != c.GetId() {
		t.Errorf("UFile: SetOwner function failed")
	}
}

func TestSetPointer(t *testing.T) {
	f := NewCFile(int64(10086), int64(1000000))
	uf := NewUFile(nil, nil, "userfile", "./user/files/")
	if uf == nil || uf.SetPointer(f) != true ||
		uf.GetPointer() == nil || uf.GetPointer().GetId() != f.GetId() {
		t.Errorf("UFile: SetPointer function failed")
	}
}

func TestPerlink(t *testing.T) {
	uf := NewUFile(nil, nil, "userfile", "./user/files/")
	if uf == nil || uf.SetPerlink("https://127.0.0.1/test") != true ||
		uf.GetPerlink() != "https://127.0.0.1/test" {
		t.Errorf("UFile: Set/GetPerlink function failed")
	}
}
