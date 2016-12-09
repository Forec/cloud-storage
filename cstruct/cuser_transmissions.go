/*
author: Forec
last edit date: 2016/12/3
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
	"bufio"
	auth "cloud-storage/authenticate"
	conf "cloud-storage/config"
	trans "cloud-storage/transmit"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type downloadItem struct {
	cfileid  int
	filename string
	size     int
}

func (u *cuser) DealWithTransmission(db *sql.DB, t trans.Transmitable) {
	fmt.Println(u.username + "Start deal with transmissions")
	defer u.RemoveTransmit(t)
	recvB, err := t.RecvBytes()
	if err != nil {
		u.RemoveTransmit(t)
		return
	}
	command := string(recvB)
	fmt.Println(command)
	switch {
	case len(command) >= 3 && strings.ToUpper(command[:3]) == "GET":
		u.get(db, command, t)
	case len(command) >= 3 && strings.ToUpper(command[:3]) == "PUT":
		u.put(db, command, t)
	default:
		t.SendBytes(auth.Int64ToBytes(300))
		fmt.Println("command not valid")
	}
	fmt.Println("finish transmission")
}

func (u *cuser) get(db *sql.DB, command string, t trans.Transmitable) {
	// format: get+uid+pass
	var err error
	var isdir, private bool
	var uid, valid int = 0, 0
	var recordCount, ownerid, cfileid, parentLength, downloaded int
	var pass, filename, originFilename, path, subpath string
	var queryRow *sql.Row
	var queryRows *sql.Rows
	var fileReader *bufio.Reader
	args := generateArgs(command, 3)
	if args == nil {
		valid = 1 // command not valid
		goto GET_VERIFY
	}
	uid, err = strconv.Atoi(args[1])
	if err != nil || strings.ToUpper(args[0]) != "GET" {
		valid = 1 // command not valid
		goto GET_VERIFY
	}
	queryRow = db.QueryRow(fmt.Sprintf(`select isdir, private, ownerid, linkpass, cfileid, filename, path, 
		downloaded from ufile where uid=%d`, uid))
	if queryRow == nil {
		valid = 2 // no such record
		goto GET_VERIFY
	}
	queryRow.Scan(&isdir, &private, &ownerid, &pass, &cfileid, &filename, &path, &downloaded)
	if int64(ownerid) != u.id && private && pass != args[2] {
		valid = 3 // no permission
		goto GET_VERIFY
	}
GET_VERIFY:
	if valid != 0 {
		t.SendBytes([]byte("NOTPERMITTED"))
		fmt.Println("invalid code: ", valid)
		return
	} else {
		t.SendBytes([]byte("VALID"))
	}
	db.Exec(fmt.Sprintf(`update ufile set downloaded=%d where uid=%d`, downloaded+1, uid))
	var totalFileLength int = 0
	if !isdir {
		// only a single file, send 1 indicate
		if !t.SendBytes(auth.Int64ToBytes(int64(1))) {
			return
		}
		// send filename
		if !t.SendBytes([]byte(filename)) {
			return
		}
		// send the isdir = 0
		if !t.SendBytes([]byte(auth.Int64ToBytes(int64(0)))) {
			return
		}
		if cfileid < 0 {
			t.SendFromReader(nil, int64(0))
		} else {
			queryRow = db.QueryRow(fmt.Sprintf(`select size from cfile where uid=%d`, cfileid))
			if queryRow == nil {
				return
			}
			err = queryRow.Scan(&totalFileLength)
			if err != nil {
				fmt.Println("scan cfile format error, ", err.Error())
			}
			file, err := os.Open(fmt.Sprintf("%s%d", conf.STORE_PATH, cfileid))
			if err != nil {
				fmt.Println(" Cannot Open Cfile")
				return
			}
			defer file.Close()
			fileReader = bufio.NewReader(file)
			t.SendFromReader(fileReader, int64(totalFileLength))
		}
	} else {
		// a folder
		// calculate how many records should be sent
		queryRow = db.QueryRow(fmt.Sprintf(`select count (*) from ufile where path like '%s%%'`,
			path+filename+"/"))
		originFilename = filename
		//	defer queryRows.Close()
		if queryRow == nil {
			return
		}
		err = queryRow.Scan(&recordCount)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		recordCount += 1
		// send count
		if !t.SendBytes(auth.Int64ToBytes(int64(recordCount))) {
			return
		}
		// send filename
		if !t.SendBytes([]byte(filename)) {
			return
		}
		// send the isdir=1
		if !t.SendBytes(auth.Int64ToBytes(int64(1))) {
			return
		}
		parentLength = len(path)

		// create the sub folders first
		queryRows, err = db.Query(fmt.Sprintf(`select filename, path from ufile where path like '%s%%' 
			and isdir=1 and ownerid=%d order by length(path)`, path+filename+"/", ownerid))
		if err != nil {
			return
		}
		for queryRows.Next() {
			err = queryRows.Scan(&filename, &subpath)
			if err != nil {
				continue
			}
			// send filename
			filename = subpath[parentLength:] + filename
			// if the system is windows, we need to replace "/" with "\\"
			if conf.CLIENT_VERSION == "Windows" {
				filename = strings.Replace(filename, "/", "\\", -1)
			}
			if !t.SendBytes([]byte(filename)) {
				return
			}
			// send the isdir=1
			if !t.SendBytes(auth.Int64ToBytes(int64(1))) {
				return
			}
		}
		// send files left
		queryRow = db.QueryRow(fmt.Sprintf(`select count (*) from ufile where path like '%s%%'
			and isdir=0 and ownerid=%d`, path+originFilename+"/", ownerid))
		if queryRow == nil {
			return
		}
		err = queryRow.Scan(&recordCount)
		if err != nil {
			return
		}
		file_list := make([]downloadItem, 0, recordCount)
		var fileItem downloadItem
		queryRows, err = db.Query(fmt.Sprintf(`select filename, path, cfileid from ufile where path like '%s%%' 
			and isdir=0 and ownerid=%d order by length(path)`, path+originFilename+"/", ownerid))
		if err != nil {
			return
		}
		for queryRows.Next() {
			err = queryRows.Scan(&filename, &subpath, &cfileid)
			if err != nil {
				continue
			}
			// get file size
			if cfileid < 0 {
				totalFileLength = 0
			} else {
				queryRow = db.QueryRow(fmt.Sprintf(`select size from cfile where uid=%d`, cfileid))
				if queryRow == nil {
					totalFileLength = 0
					continue
				}
				err = queryRow.Scan(&totalFileLength)
				if err != nil {
					totalFileLength = 0
					continue
				}
			}
			fileItem.size = totalFileLength
			filename = subpath[parentLength:] + filename
			// if the system is windows, we need to replace "/" with "\\"
			if conf.CLIENT_VERSION == "Windows" {
				filename = strings.Replace(filename, "/", "\\", -1)
			}
			fileItem.filename = filename
			fileItem.cfileid = cfileid
			file_list = append(file_list, fileItem)
		}
		for _, fileItem = range file_list {
			// send filename
			if !t.SendBytes([]byte(fileItem.filename)) {
				break
			}
			fmt.Println("send file name: ", fileItem.filename)
			// send the isdir=0
			if !t.SendBytes(auth.Int64ToBytes(int64(0))) {
				break
			}
			if fileItem.size > 0 && fileItem.cfileid >= 0 {
				// cfileid >= 0
				file, err := os.Open(fmt.Sprintf("%s%d", conf.STORE_PATH, fileItem.cfileid))
				if err != nil {
					fileItem.size = 0
					fmt.Println(" Cannot Open Cfile")
					fileReader = nil
				} else {
					defer file.Close()
					fileReader = bufio.NewReader(file)
				}
			} else {
				fileReader = nil
			}
			t.SendFromReader(fileReader, int64(fileItem.size))
		}
	}
}

func (u *cuser) put(db *sql.DB, command string, t trans.Transmitable) {
	/*
	* receive put+uid+size(bytes)+md5
	* send isTransmit
	* start transmitting or end connection
	 */
	var err1, err2, err error
	var uid, _cid, cid, size, _ref, ref int
	var shouldTransmit, valid bool = true, true
	var queryRow *sql.Row
	args := generateArgs(command, 4)
	if args == nil {
		valid = false // command not valid
	} else {
		uid, err1 = strconv.Atoi(args[1])
		size, err2 = strconv.Atoi(args[2])
		if err1 != nil || err2 != nil || strings.ToUpper(args[0]) != "PUT" ||
			!auth.IsMD5(args[3]) {
			fmt.Println("input not valid")
			fmt.Println("ismdt is ", auth.IsMD5(args[3]))
			valid = false // command not valid
		} else {
			queryRow = db.QueryRow(fmt.Sprintf(`select uid, ref from cfile 
			where md5='%s' and size=%d`, strings.ToUpper(args[3]), size))
			if queryRow == nil {
				shouldTransmit = true
			} else {
				err = queryRow.Scan(&cid, &ref)
				if err == nil {
					shouldTransmit = false
				} else {
					fmt.Println("scan cid,ref err:", err.Error())
				}
			}
		}
	}
	fmt.Println(shouldTransmit)
	if valid != true {
		t.SendBytes(auth.Int64ToBytes(300))
		fmt.Println("command not valid")
		return
	} else if shouldTransmit {
		t.SendBytes(auth.Int64ToBytes(201))
		fmt.Println("should start transmition")
		file, err := os.OpenFile(conf.STORE_PATH+args[3], os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			fmt.Println("Cannot Open File ", conf.STORE_PATH+args[3])
			t.SendBytes(auth.Int64ToBytes(500))
			fmt.Println("server internal error: cannot create openfile, ", err.Error())
			return
		}

		// transmitting files
		fileWriter := bufio.NewWriter(file)
		if t.RecvToWriter(fileWriter) {
			fmt.Println("file has been transmitted over")
		} else {
			t.SendBytes(auth.Int64ToBytes(203))
			fmt.Println("file transmission failed")
			return
		}
		_, err = db.Exec(fmt.Sprintf(`insert into cfile values(null, '%s', %d, 0, '%s')`,
			strings.ToUpper(args[3]), size, time.Now().Format("2006-01-02 15:04:05")))
		if err != nil {
			t.SendBytes(auth.Int64ToBytes(500))
			fmt.Println("server internal error: cannot insert cfile record", err.Error())
			return
		}
		queryRow = db.QueryRow(`select max(uid) from cfile`)
		if queryRow == nil {
			t.SendBytes(auth.Int64ToBytes(500))
			fmt.Println("server internal error: cannot find max cfiles' uid")
			return
		} else {
			err = queryRow.Scan(&cid)
			if err != nil {
				t.SendBytes(auth.Int64ToBytes(500))
				fmt.Println("server internal error: cannot scan max cfiles' uid", err.Error())
				return
			}
		}
		file.Close()
		err = os.Rename(conf.STORE_PATH+args[3], fmt.Sprintf("%s%d", conf.STORE_PATH, cid))
		if err != nil {
			t.SendBytes(auth.Int64ToBytes(500))
			fmt.Println("server internal error: cannot rename file to cid", err.Error())
			return
		}
		file, err = os.Open(fmt.Sprintf("%s%d", conf.STORE_PATH, cid))
		if err != nil {
			t.SendBytes(auth.Int64ToBytes(500))
			fmt.Println("calcMD5 : Cannot open target file", err.Error())
			return
		}
		fileReader := bufio.NewReader(file)
		_md5 := auth.CalcMD5ForReader(fileReader)
		file.Close()
		if _md5 == nil {
			t.SendBytes(auth.Int64ToBytes(500))
			fmt.Println("_md5 is nil!")
			return
		}
		if strings.ToUpper(string(args[3])) != strings.ToUpper(string(_md5)) {
			fmt.Println(strings.ToUpper(string(args[3])), strings.ToUpper(string(_md5)))
			t.SendBytes(auth.Int64ToBytes(403))
			fmt.Println("Your origin md5 is not valid!")
			db.Exec(fmt.Sprintf("delete from cfile where uid=%d", cid))
			os.Remove(fmt.Sprintf("%s%s", conf.STORE_PATH, cid))
			return
		}
		ref = 0
	}
	fmt.Println("transmission is over or there is no need to transmit")
	queryRow = db.QueryRow(fmt.Sprintf(`select cfileid from ufile where uid=%d and ownerid=%d`,
		uid, u.id))
	if queryRow == nil {
		t.SendBytes(auth.Int64ToBytes(301))
		fmt.Println("file not created yet")
		return
	} else {
		err = queryRow.Scan(&_cid)
		if err != nil {
			t.SendBytes(auth.Int64ToBytes(500))
			fmt.Println("server internal error: cannot scan _cid", err.Error())
			return
		}
	}
	if _cid == cid {
		t.SendBytes(auth.Int64ToBytes(200))
		fmt.Println("no need to transmit")
		return
	}
	queryRow = db.QueryRow(fmt.Sprintf(`select ref from cfile where uid=%d`, _cid))
	if queryRow != nil {
		err = queryRow.Scan(&_ref)
		if err == nil {
			if _ref != 1 {
				db.Exec(fmt.Sprintf(`update cfile set ref=%d where uid=%d`, _ref-1, _cid))
			} else {
				db.Exec(fmt.Sprintf(`delete from cfile where uid=%d`, _cid))
			}
		}
	}
	_, err = db.Exec(fmt.Sprintf(`update ufile set cfileid=%d where uid=%d and ownerid=%d`,
		cid, uid, u.id))
	fmt.Println(fmt.Sprintf(`update ufile set cfileid=%d where uid=%d and ownerid=%d`,
		cid, uid, u.id))
	if err != nil {
		t.SendBytes(auth.Int64ToBytes(500))
		fmt.Println("server internal error:cannot update ufile's cfileid", err.Error())
		return
	}
	_, err = db.Exec(fmt.Sprintf(`update cfile set ref=%d where uid=%d`,
		ref+1, cid))
	if err != nil {
		t.SendBytes(auth.Int64ToBytes(500))
		fmt.Println("server internal error:cannot update cfile's ref", err.Error())
	} else {
		t.SendBytes(auth.Int64ToBytes(200))
		fmt.Println("all put mission has been done")
	}
}
