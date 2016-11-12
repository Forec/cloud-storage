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
	conf "Cloud/config"
	trans "Cloud/transmit"
)

func AppendUser(slice []User, data ...User) []User {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) {
		newSlice := make([]User, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}

func AppendUFile(slice []UFile, data ...UFile) []UFile {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) {
		newSlice := make([]UFile, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}

func AppendTransmitable(slice []trans.Transmitable, data ...trans.Transmitable) []trans.Transmitable {
	if len(slice)+len(data) >= conf.MAXTRANSMITTER {
		slice = append(slice, data[:conf.MAXTRANSMITTER-len(slice)]...)
	} else {
		slice = append(slice, data...)
	}
	return slice
}

func UFileIndexByPath(slice []UFile, path string) []UFile {
	filter := make([]UFile, 0, 10)
	for _, uf := range slice {
		if uf.GetPath() == path {
			filter = AppendUFile(filter, uf)
		}
	}
	return filter
}

func UFileIndexById(slice []UFile, id int64) []UFile {
	filter := make([]UFile, 0, 10)
	for _, uf := range slice {
		if uf.GetPointer().GetId() == id {
			filter = AppendUFile(filter, uf)
		}
	}
	return filter
}

func UserIndexByName(slice []User, name string) User {
	for _, uc := range slice {
		if uc.GetUsername() == name {
			return uc
		}
	}
	return nil
}
