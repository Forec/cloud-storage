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
	"time"
)

/* SHOW HOW UFILE WORKS
--cfile is stored in disk, any reference to the cfile
   is recorded as a ufile for different cusers.
 -------            -------          -------------
| cfile |  <-----  | ufile |  <---  | cuser (id=1)|
 -------            -------          -------------
   /|               -------          -------------
    |<-----------  | ufile |  <---  | cuser (id=2)|
                    -------          -------------
*/

/* UFILE DECLARATION
 * pointer     : point to the read cfile
 * owner       : owner of this record
 * filename    : filename for this record
 * path        : path for this record
 * perlink     : perlink for this record
 * timestamp   : timestamp for this record
 * shared      : share times for this ufile(from its related cuser)
 * downloaded  : download times for this ufile (from its related cuser)
 */
type ufile struct {
	pointer    *cfile
	owner      *cuser
	filename   string
	path       string
	perlink    string
	timestamp  time.Time
	shared     int32
	downloaded int32
}

type UFile interface {
	GetFilename() string
	GetShared() int32
	GetDownloaded() int32
	GetPath() string
	GetPerlink() string
	GetTime() time.Time
	GetPointer() *cfile
	GetOwner() *cuser
	IncShared() bool
	IncDowned() bool
	SetPath(string) bool
	SetPerlink(string) bool
	SetPointer(*cfile) bool
	SetOwner(*cuser) bool
}

/* UFILE METHODS
 * CONSTRUCTOR: NewUfile(upointer *cfile,
						 uname    string,
						 upath	  string) *ufile
 * MODIFIER: Set[VALUE] (v 	  VALUE_TYPE) bool
			 IncShared  ()				  bool
			 IncDowned	()				  bool
*/

func NewUFile(upointer *cfile, uowner *cuser, uname string, upath string) *ufile {
	u := new(ufile)
	u.downloaded = 0
	u.shared = 0
	u.pointer = upointer
	u.owner = uowner
	u.filename = uname
	u.path = upath
	u.timestamp = time.Now()
	u.perlink = ""
	if upointer != nil {
		upointer.AddRef(1)
	}
	return u
}

func (u *ufile) GetFilename() string {
	return u.filename
}

func (u *ufile) GetShared() int32 {
	return u.shared
}

func (u *ufile) GetTime() time.Time {
	return u.timestamp
}

func (u *ufile) GetDownloaded() int32 {
	return u.downloaded
}

func (u *ufile) GetPath() string {
	return u.path
}

func (u *ufile) GetPerlink() string {
	return u.perlink
}

func (u *ufile) GetPointer() *cfile {
	return u.pointer
}

func (u *ufile) GetOwner() *cuser {
	return u.owner
}

func (u *ufile) IncShared() bool {
	u.shared++
	//if u.pointer != nil {
	//	u.pointer.AddRef(1)
	//}
	return true
}

func (u *ufile) IncDowned() bool {
	u.downloaded++
	return true
}

func (u *ufile) SetPath(upath string) bool {
	u.path = upath
	return true
}

func (u *ufile) SetPerlink(uperlink string) bool {
	u.perlink = uperlink
	return true
}

func (u *ufile) SetPointer(upointer *cfile) bool {
	u.pointer = upointer
	return true
}

func (u *ufile) SetOwner(uowner *cuser) bool {
	u.owner = uowner
	return true
}
