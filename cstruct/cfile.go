package cstruct

import (
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
