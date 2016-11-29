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
	trans "cloud-storage/transmit"
	"testing"
)

func TestNewCuser(t *testing.T) {
	u := NewCUser("default", 0, ".")
	if u == nil {
		t.Errorf("CUser: NewCuser function failed")
	}
}

func TestGetUsername(t *testing.T) {
	u := NewCUser("default", 0, ".")
	if u == nil || u.GetUsername() != "default" {
		t.Errorf("CUser: GetUsername function failed")
	}
}

func TestGetId(t *testing.T) {
	u := NewCUser("default", 0, ".")
	if u == nil || u.GetId() != 0 {
		t.Errorf("CUser: GetId function failed")
	}
}

func TestGetToken(t *testing.T) {
	u := NewCUser("default", 0, ".")
	if u == nil || u.GetToken() != "" {
		t.Errorf("CUser: GetToken function failed")
	}
}

func TestAddUFile(t *testing.T) {
	u := NewCUser("default", 0, ".")
	c := NewCFile(int64(10086), int64(65536))
	uf := NewUFile(c, u, "userfile", "./user/files/")
	if u.AddUFile(uf) != true {
		t.Errorf("CUser: AddUFile function failed")
	}
}

func TestRemoveUFile(t *testing.T) {
	u := NewCUser("default", int64(12345), ".")
	c := NewCFile(int64(10086), int64(65536))
	uf := NewUFile(c, u, "userfile", "./user/files/")
	if u.AddUFile(uf) != true || u.RemoveUFile(uf) != true {
		t.Errorf("CUser: RemoveUFile function failed")
	}
}

func TestAddTransmitter(t *testing.T) {
	u := NewCUser("default", 0, ".")
	c := trans.NewTransmitter(nil, 0, nil)
	if u.AddTransmit(c) != true {
		t.Errorf("CUser: AddTransmitter function failed")
	}
}

func TestRemoveTransmitter(t *testing.T) {
	u := NewCUser("default", 0, ".")
	c := trans.NewTransmitter(nil, 0, nil)
	if u.AddTransmit(c) != true || u.RemoveTransmit(c) != true {
		t.Errorf("CUser: RemoveTransmitter function failed")
	}
}
