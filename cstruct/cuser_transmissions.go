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
	auth "Cloud/authenticate"
	conf "Cloud/config"
	trans "Cloud/transmit"
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	case len(command) >= 6 && strings.ToUpper(command[:2]) == "UPDATE":
		u.update(db, command, t)
	default:
		t.SendBytes([]byte("Invalid"))
	}
	fmt.Println("finish transmission")
}

func (u *cuser) get(db *sql.DB, command string, t trans.Transmitable) {
	// format: get+uid+pass
	var err error
	var isdir, private bool
	var uid, valid int = 0, 0
	var recordCount, ownerid, cfileid, parentLength int
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
	queryRow = db.QueryRow(fmt.Sprintf(`select isdir, private, ownerid, linkpass, cfileid, filename, path
		from ufile where uid=%d`, uid))
	if queryRow == nil {
		valid = 2 // no such record
		goto GET_VERIFY
	}
	queryRow.Scan(&isdir, &private, &ownerid, &pass, &cfileid, &filename, &path)
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
	// TODO
}

func (u *cuser) update(db *sql.DB, command string, t trans.Transmitable) {
	// TODO
}
