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
	"strings"
	"time"
)

/* CFILE DECLARATION
 * id         : primary key for cfile
 * ref        : the number of users forking this cfile
 * size       : file size
 * downloaded : the number of times this cfile has been downloaded
 * timestamp  : cfile created time
 * userlist   : list of users forking this cfile
 */

type cfile struct {
	id         int64
	ref        int32
	size       int64
	downloaded int64
	timestamp  time.Time
	userlist   []*cuser
}

type CFile interface {
	GetId() int64
	GetTimestamp() time.Time
	GetSize() int64
	GetRef() int32
	SetId(int64) bool
	SetSize(int64) bool
	AddRef(int32) bool
}

/* CFILE METHODS
 * CONSTRUCTOR: NewCFile(fid 	int64,
						fname 	string,
						fpath 	string,
						fsize 	int64) 	*cfile

 * MODIFIER: Set[VALUE] (v VALUE_TYPE) 	bool
			 AddUser	(u *cuser) 	   	bool
*/

func NewCFile(fid int64, fsize int64) *cfile {
	f := new(cfile)
	f.ref = 0
	f.id = fid
	f.size = fsize
	f.userlist = nil
	f.timestamp = time.Now()
	return f
}

func (f *cfile) GetId() int64 {
	return f.id
}

func (f *cfile) GetTimestamp() time.Time {
	return f.timestamp
}

func (f *cfile) GetSize() int64 {
	return f.size
}

func (f *cfile) GetRef() int32 {
	return f.ref
}

func (f *cfile) SetId(fid int64) bool {
	f.id = fid
	return true
}

func (f *cfile) SetSize(fsize int64) bool {
	f.size = fsize
	return true
}

func (f *cfile) AddRef(offset int32) bool {
	f.ref += offset
	return true
}

func isFilenameValid(filename string) bool {
	if len(filename) > 128 ||
		strings.Count(filename, "/") > 0 ||
		strings.Count(filename, "\\") > 0 ||
		strings.Count(filename, "+") > 0 ||
		strings.Count(filename, ":") > 0 ||
		strings.Count(filename, "*") > 0 ||
		strings.Count(filename, "?") > 0 ||
		strings.Count(filename, "<") > 0 ||
		strings.Count(filename, ">") > 0 ||
		strings.Count(filename, "\"") > 0 {
		return false
	} else {
		return true
	}
}

func isPathFormatValid(path string) bool {
	if len(path) < 2 ||
		len(path) > 256 ||
		path[0] != '/' ||
		path[1] != '/' ||
		strings.Count(path, "../") > 0 ||
		strings.Count(path, "/..") > 0 ||
		strings.Count(path, "+") > 0 ||
		strings.Count(path, ":") > 0 ||
		strings.Count(path, "*") > 0 ||
		strings.Count(path, "?") > 0 ||
		strings.Count(path, "<") > 0 ||
		strings.Count(path, ">") > 0 ||
		strings.Count(path, "\"") > 0 {
		return false
	} else {
		return true
	}
}
