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
