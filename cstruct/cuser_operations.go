/*
author: Forec
last edit date: 2016/12/7
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
	auth "cloud-storage/authenticate"
	conf "cloud-storage/config"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

/*
	Logical
*/

type id_path struct {
	u_id   int
	u_path string
}

type path_name struct {
	u_path string
	u_name string
}

type ufile_record struct {
	uid                                                     int
	ownerid, cfileid, shared, downloaded                    int
	private, isdir                                          bool
	linkpass, path, created, filename, perlink, description string
}

func (u *cuser) DealWithRequests(db *sql.DB) {
	u.curpath = "/"
	fmt.Println(u.username + "Start deal with args")
	for {
		recvB, err := u.listen.RecvBytes()
		if err != nil {
			return
		}
		command := string(recvB)
		fmt.Println(command)
		switch {
		case len(command) >= 2 && strings.ToUpper(command[:2]) == "RM":
			u.rm(db, command)
		case len(command) >= 2 && strings.ToUpper(command[:2]) == "CP":
			u.cp(db, command)
		case len(command) >= 2 && strings.ToUpper(command[:2]) == "MV":
			u.mv(db, command)
		case len(command) >= 2 && strings.ToUpper(command[:2]) == "LS":
			u.ls(db, command)
		case len(command) >= 4 && strings.ToUpper(command[:4]) == "SEND":
			u.send(db, command)
		case len(command) >= 4 && strings.ToUpper(command[:4]) == "FORK":
			u.fork(db, command)
		case len(command) >= 5 && strings.ToUpper(command[:5]) == "TOUCH":
			u.touch(db, command)
		case len(command) >= 5 && strings.ToUpper(command[:5]) == "CHMOD":
			u.chmod(db, command)
		default:
			u.listen.SendBytes([]byte("Invalid Command"))
		}
	}
}

func generateArgs(command string, arglen int) []string {
	args := strings.Split(command, conf.SEPERATER)
	if arglen > 0 && len(args) != arglen {
		return nil
	}
	for i, arg := range args {
		args[i] = strings.Trim(arg, " ")
		args[i] = strings.Replace(args[i], "\r", "", -1)
		args[i] = strings.Replace(args[i], "\n", "", -1)
		if args[i] == "" {
			fmt.Println("got invalid arg")
			return nil
		}
		fmt.Printf("%s ", args[i])
	}
	fmt.Println()
	return args
}

func generateSubpaths(path string) []path_name {
	if !isPathFormatValid(path) {
		return nil
	}
	var record path_name
	var records []path_name = nil
	records = make([]path_name, 0, strings.Count(path, "/"))
	for {
		cp := strings.LastIndex(path, "/")
		path = path[:cp]
		cp = strings.LastIndex(path, "/")
		if cp < 0 {
			return records
		}
		record.u_name = path[cp+1:]
		record.u_path = path[:cp+1]
		records = append(records, record)
	}
	return records
}

func (u *cuser) send(db *sql.DB, command string) {
	// format: send+uid+message
	var uid int
	var message string = ""
	var err error
	var valid bool = true
	args := generateArgs(command, 0)
	if args == nil || len(args) < 2 {
		valid = false
		goto SEND_VERIFY
	}
	uid, err = strconv.Atoi(args[1])
	if err != nil {
		valid = false
		goto SEND_VERIFY
	}
	for i := 2; i < len(args); i++ {
		message += (args[i] + " ")
	}
	_, err = db.Exec(fmt.Sprintf(`insert into cmessages values(null, %d, %d, '%s', '%s', 0, 0,0, 0)`,
		uid, u.id, message, time.Now().Format("2006-01-02 15:04:05")))
	if err != nil {
		fmt.Println(err.Error())
		valid = false
	}
SEND_VERIFY:
	if !valid {
		u.listen.SendBytes(auth.Int64ToBytes(int64(400)))
	} else {
		u.listen.SendBytes(auth.Int64ToBytes(int64(200)))
	}
}

func (u *cuser) chmod(db *sql.DB, command string) {
	// format: chmod+uid+private
	var queryRow *sql.Row
	var queryRows *sql.Rows
	var uid, private, ownerid, recordCount int
	var isPrivate, isdir bool
	var valid bool = true
	var originPath, originName string
	var err, err1 error
	var uidList []int

	args := generateArgs(command, 3)
	if args == nil || strings.ToUpper(args[0]) != "CHMOD" {
		valid = false
		goto CHMOD_VERIFY
	}
	uid, err = strconv.Atoi(args[1])
	private, err1 = strconv.Atoi(args[2])
	if err != nil || err1 != nil {
		valid = false
		goto CHMOD_VERIFY
	}
	// check whether the current user valid
	queryRow = db.QueryRow(fmt.Sprintf(`select ownerid, private, path, filename, isdir from ufile where uid=%d`, uid))
	if queryRow == nil {
		valid = false
		goto CHMOD_VERIFY
	}
	err = queryRow.Scan(&ownerid, &isPrivate, &originPath, &originName, &isdir)
	if err != nil {
		valid = false
		fmt.Println(err.Error())
		goto CHMOD_VERIFY
	}
	if int64(ownerid) != u.id { // current user is not valid
		valid = false
		goto CHMOD_VERIFY
	}
	// change the origin record mod
	_, err = db.Exec(fmt.Sprintf(`update ufile set private=%d where uid=%d and ownerid=%d`, private, uid, u.id))
	if err != nil {
		fmt.Println("chmod error: ", err.Error())
		valid = false
		goto CHMOD_VERIFY
	}
	if isdir {
		// get the count of records under the origin folder
		queryRow = db.QueryRow(fmt.Sprintf(`select count (*) from ufile where path like '%s%%' and ownerid=%d`,
			originPath+originName+"/", u.id))
		if queryRow == nil {
			fmt.Println("get count from ufile error")
			valid = false
			goto CHMOD_VERIFY
		}
		err = queryRow.Scan(&recordCount)
		if err != nil {
			fmt.Println("get count from ufile format error: ", err.Error())
			valid = false
			goto CHMOD_VERIFY
		}
		uidList = make([]int, 0, recordCount)
		// get the uids under the origin folder
		queryRows, err = db.Query(fmt.Sprintf(`select uid from ufile where path like '%s%%' and ownerid=%d`,
			originPath+originName+"/", u.id))
		defer queryRows.Close()
		if err != nil {
			valid = false
			fmt.Println("select uid from ufile error: ", err.Error())
			goto CHMOD_VERIFY
		}
		for queryRows.Next() {
			err = queryRows.Scan(&uid)
			if err != nil {
				fmt.Println("ufile record format error: ", err.Error())
				valid = false
				goto CHMOD_VERIFY
			}
			uidList = append(uidList, uid)
		}
		for _, uid := range uidList {
			_, err = db.Exec(fmt.Sprintf(`update ufile set private=%d where uid=%d and ownerid=%d`,
				private, uid, u.id))
			if err != nil {
				valid = false
				fmt.Println("update ufile error: ", err.Error())
				goto CHMOD_VERIFY
			}
		}
	}
CHMOD_VERIFY:
	if !valid {
		u.listen.SendBytes(auth.Int64ToBytes(int64(400)))
	} else {
		u.listen.SendBytes(auth.Int64ToBytes(int64(200)))
	}
}

func (u *cuser) rm(db *sql.DB, command string) {
	// formant: rm+uid

	var queryRow *sql.Row
	var queryRows *sql.Rows

	// record items declaration
	var ref, recordCount, uid int
	var record ufile_record
	var crecords []int
	var valid int = -1
	var err error
	var tempSize int = 0
	args := generateArgs(command, 2)
	if args == nil || strings.ToUpper(args[0]) != "RM" {
		valid = 0 // command not valid
		goto RM_VERIFY
	}
	uid, err = strconv.Atoi(args[1])
	if err != nil {
		valid = 0 // command not valid
		goto RM_VERIFY
	}
	// find the record to fork
	queryRow = db.QueryRow(fmt.Sprintf(`select * from ufile where uid=%d and ownerid=%d`, uid, u.id))
	if queryRow == nil {
		valid = 1 // the record not exists
		goto RM_VERIFY
	}
	err = queryRow.Scan(&record.uid, &record.ownerid, &record.cfileid, &record.path,
		&record.perlink, &record.created, &record.shared, &record.downloaded, &record.filename,
		&record.private, &record.linkpass, &record.isdir, &record.description)
	fmt.Println("record.ownerid: ", record.ownerid, "  uid: ", uid)
	if err != nil || int64(record.ownerid) != u.id {
		valid = 2 // the record format not valid or the user is invalid
		fmt.Println("record format not valid or the user is invalid")
		fmt.Println(err.Error())
		goto RM_VERIFY
	}
	_, err = db.Exec(fmt.Sprintf(`delete from ufile where uid=%d and ownerid=%d`, uid, u.id))
	fmt.Println("delete from ufile")
	if err != nil {
		fmt.Println("delet failed: ", err.Error())
		valid = 4 // database operate error
		goto RM_VERIFY
	}
	if !record.isdir { // the record is a file
		if record.cfileid >= 0 {
			queryRow = db.QueryRow(fmt.Sprintf(`select ref, size from cfile where uid=%d`, record.cfileid))
			if queryRow == nil {
				valid = 1 // the record not exists
				fmt.Println("cannot find reference cfile")
				goto RM_VERIFY
			}
			err = queryRow.Scan(&ref, &tempSize)
			if err != nil {
				fmt.Println("format error: ", err.Error())
				valid = 2 // the record format is not valid
				goto RM_VERIFY
			}
			u.used -= int64(tempSize)
			if ref == 1 { // the cfile is not referred any more
				_, err = db.Exec(fmt.Sprintf(`delete from cfile where uid=%d`,
					record.cfileid))
				if err != nil {
					valid = 4 // database operate error
					goto RM_VERIFY
				}
			} else { // the cfile is still been referred
				_, err = db.Exec(fmt.Sprintf(`update cfile set ref=%d where uid=%d`,
					ref-1, record.cfileid))
				if err != nil {
					valid = 4 // database operate error
					goto RM_VERIFY
				}
			}
		}
	} else { // the record is a folder
		queryRow = db.QueryRow(fmt.Sprintf(`select count (*) from ufile where path like '%s%%' and ownerid=%d`,
			record.path+record.filename+"/", u.id))
		if queryRow == nil {
			valid = 1 // the record not exists
			fmt.Println("record not exists")
			goto RM_VERIFY
		}
		err = queryRow.Scan(&recordCount)
		if err != nil {
			fmt.Println("recordCount error: ", err.Error())
			valid = 2 // record format not valid
			goto RM_VERIFY
		}
		crecords = make([]int, 0, recordCount)
		queryRows, err = db.Query(fmt.Sprintf(`select cfileid from ufile where path like '%s%%' and ownerid=%d 
			and cfileid>=0 and isdir=0`, record.path+record.filename+"/", u.id))
		defer queryRows.Close()
		if err != nil {
			fmt.Println("select cfile error: ", err.Error())
			valid = 5 // database search error
			goto RM_VERIFY
		}
		for queryRows.Next() {
			err = queryRows.Scan(&ref)
			if err != nil {
				fmt.Println("record format error: ", err.Error())
				valid = 2 // record format not valid
				goto RM_VERIFY
			}
			crecords = append(crecords, ref)
		}
		_, err = db.Exec(fmt.Sprintf(`delete from ufile where path like '%s%%' and ownerid=%d`,
			record.path+record.filename+"/", u.id))
		if err != nil {
			fmt.Println("delete from ufile error:", err.Error())
			valid = 4 // database operate error
			goto RM_VERIFY
		}
		for _, cid := range crecords {
			queryRow = db.QueryRow(fmt.Sprintf(`select ref, size from cfile where uid=%d`, cid))
			if queryRow == nil {
				// the record not exists
				continue
			}
			err = queryRow.Scan(&ref, &tempSize)
			if err != nil {
				fmt.Println("foramt error :", err.Error())
				valid = 2 // record format invalid
				goto RM_VERIFY
			}
			u.used -= int64(tempSize)
			if ref == 1 { // the cfile is not referred any more
				_, err = db.Exec(fmt.Sprintf(`delete from cfile where uid=%d`,
					cid))
				if err != nil {
					fmt.Println(err.Error())
					valid = 4 // database operate error
					goto RM_VERIFY
				}
			} else { // the cfile is still been referred
				_, err = db.Exec(fmt.Sprintf(`update cfile set ref=%d where uid=%d`,
					ref-1, cid))
				if err != nil {
					valid = 4 // database operate error
					goto RM_VERIFY
				}
			}
		}
	}
RM_VERIFY:
	if valid != -1 {
		u.listen.SendBytes(auth.Int64ToBytes(int64(valid)))
		fmt.Println("valid is ", valid)
	} else {
		u.listen.SendBytes(auth.Int64ToBytes(int64(200)))
	}
}

func (u *cuser) fork(db *sql.DB, command string) {
	// format: fork+uid+pass+newpath
	var uid, recordCount, ref int
	var isdir_int int = 0
	var valid int = -1
	var err error
	var originPath, queryString string
	var subpaths []path_name
	var queryRow *sql.Row
	var queryRows *sql.Rows
	var tempSize int = 0

	// record items declaration
	var record ufile_record
	var recordList []ufile_record

	args := generateArgs(command, 4)
	if args == nil {
		valid = 0 // command not valid
		goto FORK_VERIFY
	}
	// change into uid
	uid, err = strconv.Atoi(args[1])
	// check uid, filename, path
	if err != nil || !isPathFormatValid(args[3]) || strings.ToUpper(args[0]) != "FORK" {
		valid = 0 // command not valid
		goto FORK_VERIFY
	}
	// find the record to fork
	queryRow = db.QueryRow(fmt.Sprintf(`select * from ufile where uid=%d`, uid))
	if queryRow == nil {
		valid = 1 // the record not exists
		fmt.Println("the record is not in database")
		goto FORK_VERIFY
	}
	err = queryRow.Scan(&record.uid, &record.ownerid, &record.cfileid, &record.path,
		&record.perlink, &record.created, &record.shared, &record.downloaded, &record.filename,
		&record.private, &record.linkpass, &record.isdir, &record.description)
	if err != nil {
		fmt.Println("format error:", err.Error())
		valid = 2 // the record format is not valid
		goto FORK_VERIFY
	}
	// if the record is private and password not match or the record belongs to the current user
	if record.private && record.linkpass != args[2] || int64(record.ownerid) == u.id {
		fmt.Println("password not match")
		valid = 3 // the password is unmatched or current user
		goto FORK_VERIFY
	}
	// check whether all the folders in newpath have been created
	subpaths = generateSubpaths(args[3])
	for _, subpath := range subpaths {
		queryRow = db.QueryRow(fmt.Sprintf(`select count(*) from ufile where ownerid=%d and 
		filename='%s' and path='%s' and isdir=1`, u.id, subpath.u_name, subpath.u_path))
		if queryRow == nil {
			valid = 5 // database write error
			goto FORK_VERIFY
		}
		err = queryRow.Scan(&recordCount)
		if err != nil {
			fmt.Println(err.Error())
			valid = 2 // record format not valid
			goto FORK_VERIFY
		}
		if recordCount <= 0 {
			_, err = db.Exec(fmt.Sprintf(`insert into ufile values(null, %d, -1, '%s', '', '%s', 0, 0, 
				'%s', 1, '', 1, '')`, u.id, subpath.u_path, time.Now().Format("2006-01-02 15:04:05"), subpath.u_name))
			if err != nil {
				fmt.Println(err.Error())
				valid = 4 //database write error
			}
		}
	}
	if record.isdir {
		isdir_int = 1
	}
	originPath = record.path
	// insert new record
	_, err = db.Exec(fmt.Sprintf(`insert into ufile values(null, %d, %d, '%s', '', '%s', 0, 0, '%s', 1, '', %d, '')`,
		u.id, record.cfileid, args[3], time.Now().Format("2006-01-02 15:04:05"), record.filename, isdir_int))
	if err != nil {
		fmt.Println("1:", err.Error())
		valid = 4 // database write error
		goto FORK_VERIFY
	}
	// update the origin ufile's shared
	_, err = db.Exec(fmt.Sprintf(`update ufile set shared=%d where uid=%d`, record.shared+1, record.uid))
	if err != nil {
		fmt.Println("update ufile error", err.Error())
		valid = 4 // database write error
		goto FORK_VERIFY
	}
	// update the cfile reference if the record is not a folder
	if !record.isdir && record.cfileid >= 0 {
		queryRow = db.QueryRow(fmt.Sprintf(`select ref, size from cfile where uid=%d`, record.cfileid))
		if queryRow == nil {
			fmt.Println("cfile record not exists")
			valid = 1 // record not exists
			goto FORK_VERIFY
		}
		err = queryRow.Scan(&ref, &tempSize)
		if err != nil {
			fmt.Println("cfile record format error: ", err.Error())
			valid = 2 // record format not valid
			goto FORK_VERIFY
		}
		u.used += int64(tempSize)
		_, err = db.Exec(fmt.Sprintf(`update cfile set ref=%d where uid=%d`, ref+1, record.cfileid))
		if err != nil {
			fmt.Println("update cfile error", err.Error())
			valid = 4 // database write error
			goto FORK_VERIFY
		}
	}
	if record.isdir {
		// if the record is a folder, all the files/folders belong to this folder should be
		// forked to the new folder path
		queryRow = db.QueryRow(fmt.Sprintf(`select count (*) from ufile where path like '%s%%' and ownerid=%d`,
			record.path+record.filename+"/", record.ownerid))
		if queryRow == nil {
			valid = 5 // database search error
			goto FORK_VERIFY
		}
		err = queryRow.Scan(&recordCount)
		if err != nil {
			fmt.Println("2:", err.Error())
			valid = 2 // record format not valid
			goto FORK_VERIFY
		}
		// get all the records under the parent record
		queryString = fmt.Sprintf(`select * from ufile where path like '%s%%' and ownerid=%d`,
			record.path+record.filename+"/", record.ownerid)
		fmt.Println(queryString)
		queryRows, err = db.Query(queryString)
		defer queryRows.Close()
		if err != nil {
			fmt.Println("3:", err.Error())
			valid = 5 // database search error
			goto FORK_VERIFY
		}
		recordList = make([]ufile_record, 0, recordCount)
		// get all records belong to the dir
		for queryRows.Next() {
			err = queryRow.Scan(&record.uid, &record.ownerid, &record.cfileid, &record.path,
				&record.perlink, &record.created, &record.shared, &record.downloaded, &record.filename,
				&record.private, &record.linkpass, &record.isdir, &record.description)
			if err != nil {
				fmt.Println("4:", err.Error())
				valid = 2 // record format not valid
				goto FORK_VERIFY
			}
			recordList = append(recordList, record)
		}
		// update
		for _, record = range recordList {
			if record.isdir {
				isdir_int = 1
			} else {
				isdir_int = 0
			}
			_, err = db.Exec(fmt.Sprintf(`insert into ufile values(null, %d, %d, '%s', '', '%s', 0, 0, '%s', 1, '', %d, '')`,
				u.id, record.cfileid, args[2]+record.path[len(originPath):], time.Now().Format("2006-01-02 15:04:05"), record.filename, isdir_int))
			if err != nil {
				fmt.Println("6:", err.Error())
				valid = 4 // database write error
				goto FORK_VERIFY
			}
			// update the shared number of origin ufile
			_, err = db.Exec(fmt.Sprintf(`update ufile set shared=%d where uid=%d`, record.shared+1, record.uid))
			if err != nil {
				valid = 4 // database write error
				goto FORK_VERIFY
			}
			// update the cfile reference, if the ufile record is not a folder and its cfileid >= 0
			if record.cfileid >= 0 && !record.isdir {
				queryRow = db.QueryRow(fmt.Sprintf(`select ref, size from cfile where uid=%d`, record.cfileid))
				if queryRow == nil {
					valid = 1 // record not exists
					goto FORK_VERIFY
				}
				err = queryRow.Scan(&ref, &tempSize)
				if err != nil {
					valid = 2 // record format not valid
					goto FORK_VERIFY
				}
				u.used += int64(tempSize)
				_, err = db.Exec(fmt.Sprintf(`update cfile set ref=%d where uid=%d`, ref+1, record.cfileid))
				if err != nil {
					valid = 4 // database write error
					goto FORK_VERIFY
				}
			}
		}
	}
FORK_VERIFY:
	if valid != -1 {
		u.listen.SendBytes(auth.Int64ToBytes(int64(valid)))
	} else {
		u.listen.SendBytes(auth.Int64ToBytes(int64(200)))
	}
}

func (u *cuser) cp(db *sql.DB, command string) {
	//format: cp+uid+newpath
	var valid, isdir, uprivate bool = true, false, true
	var uid, recordCount, isdir_int, ref int
	var err error
	var path, parentPath, originName, queryString string
	var uownerid, cfileid, ushared, udownloaded int
	var upath, uperlink, ufilename, ulinkpass, ucreated, udescription string
	var queryRow *sql.Row
	var queryRows *sql.Rows
	var recordList []id_path
	var record id_path
	var subpaths []path_name
	args := generateArgs(command, 3)
	if args == nil {
		valid = false
		goto CP_VERIFY
	}
	uid, err = strconv.Atoi(args[1])
	// check uid, filename, path
	if err != nil || !isPathFormatValid(args[2]) || strings.ToUpper(args[0]) != "CP" {
		valid = false
		goto CP_VERIFY
	}
	// find the original path of record to be copied
	queryRow = db.QueryRow(fmt.Sprintf(`select * from ufile where uid=%d`, uid))
	if queryRow == nil {
		valid = false
		goto CP_VERIFY
	}
	err = queryRow.Scan(&uid, &uownerid, &cfileid, &parentPath, &uperlink, &ucreated, &ushared, &udownloaded,
		&originName, &uprivate, &ulinkpass, &isdir, &udescription)
	if err != nil || int64(uownerid) != u.id {
		valid = false
		goto CP_VERIFY
	}

	// check whether all the folders in newpath have been created
	subpaths = generateSubpaths(args[2])
	for _, subpath := range subpaths {
		queryRow = db.QueryRow(fmt.Sprintf(`select count(*) from ufile where ownerid=%d and 
		filename='%s' and path='%s' and isdir=1`, u.id, subpath.u_name, subpath.u_path))
		if queryRow == nil {
			valid = false
			goto CP_VERIFY
		}
		err = queryRow.Scan(&recordCount)
		if err != nil {
			fmt.Println(err.Error())
			valid = false
			goto CP_VERIFY
		}
		if recordCount <= 0 {
			_, err = db.Exec(fmt.Sprintf(`insert into ufile values(null, %d, -1, '%s', '', '%s', 0, 0, 
				'%s', 1, '', 1, '')`, u.id, subpath.u_path, time.Now().Format("2006-01-02 15:04:05"), subpath.u_name))
			if err != nil {
				fmt.Println(err.Error())
				valid = false
			}
		}
	}
	if isdir {
		isdir_int = 1
	} else {
		isdir_int = 0
	}
	// if the record is a single file, add the cfile reference
	if !isdir && cfileid >= 0 {
		queryRow = db.QueryRow(fmt.Sprintf(`select ref from cfile where uid=%d`, cfileid))
		if queryRow == nil {
			valid = false
			fmt.Println("cannot select ref from cfile")
			goto CP_VERIFY
		}
		err = queryRow.Scan(&ref)
		if err != nil {
			valid = false
			fmt.Println("cfile record format error: ", err.Error())
			goto CP_VERIFY
		}
		_, err = db.Exec(fmt.Sprintf(`update cfile set ref=%d where uid=%d`, ref+1, cfileid))
		if err != nil {
			valid = false
			fmt.Println("update cfile error")
			goto CP_VERIFY
		}
	}
	// no matter the record is a file or folder, its filename and path should be copied
	_, err = db.Exec(fmt.Sprintf(`insert into ufile values(null, %d, %d, '%s', '', '%s', 0, 0, '%s', 1, '', %d, '')`,
		u.id, cfileid, args[2], time.Now().Format("2006-01-02 15:04:05"), originName, isdir_int))
	if err != nil {
		fmt.Println("1:", err.Error())
		valid = false
		goto CP_VERIFY
	}
	if isdir {
		// if the record is a folder, all the files/folders belong to this folder should be
		// copied to the new folder path
		queryRow = db.QueryRow(fmt.Sprintf(`select count (*) from ufile where path like '%s%%' and ownerid=%d`,
			parentPath+originName+"/", u.id))
		if queryRow == nil {
			valid = false
			goto CP_VERIFY
		}
		err = queryRow.Scan(&recordCount)
		if err != nil {
			fmt.Println("2:", err.Error())
			valid = false
			goto CP_VERIFY
		}
		// get all the records under the parent record
		queryString = fmt.Sprintf(`select uid, path from ufile where path like '%s%%' and ownerid=%d`,
			parentPath+originName+"/", u.id)
		fmt.Println(queryString)
		queryRows, err = db.Query(queryString)
		if err != nil {
			fmt.Println("3:", err.Error())
			valid = false
			goto CP_VERIFY
		}
		defer queryRows.Close()
		recordList = make([]id_path, 0, recordCount)
		// get all records belong to the dir
		for queryRows.Next() {
			err = queryRows.Scan(&uid, &path)
			if err != nil {
				fmt.Println("4:", err.Error())
				valid = false
				goto CP_VERIFY
			}
			record.u_id = uid
			record.u_path = path
			recordList = append(recordList, record)
		}
		// update
		for _, record = range recordList {
			queryRow = db.QueryRow(fmt.Sprintf(`select * from ufile where uid=%d and ownerid=%d`, uid, u.id))
			if queryRow == nil {
				valid = false
				goto CP_VERIFY
			}
			err = queryRow.Scan(&uid, &uownerid, &cfileid, &upath, &uperlink, &ucreated, &ushared, &udownloaded,
				&ufilename, &uprivate, &ulinkpass, &isdir, &udescription)
			if err != nil {
				fmt.Println("5:", err.Error())
				valid = false
				goto CP_VERIFY
			}
			if isdir {
				isdir_int = 1
			} else {
				isdir_int = 0
			}
			_, err = db.Exec(fmt.Sprintf(`insert into ufile values(null, %d, %d, '%s', '', '%s', 0, 0, '%s', 1, '', %d, '')`,
				u.id, cfileid, args[2]+record.u_path[len(parentPath):], time.Now().Format("2006-01-02 15:04:05"), ufilename, isdir_int))
			if err != nil {
				fmt.Println("6:", err.Error())
				valid = false
				goto CP_VERIFY
			}
			// update cfile reference
			if !isdir && cfileid >= 0 {
				queryRow = db.QueryRow(fmt.Sprintf(`select ref from cfile where uid=%d`, cfileid))
				if queryRow == nil {
					valid = false
					fmt.Println("cannot select ref from cfile")
					goto CP_VERIFY
				}
				err = queryRow.Scan(&ref)
				if err != nil {
					valid = false
					fmt.Println("cfile record format error: ", err.Error())
					goto CP_VERIFY
				}
				_, err = db.Exec(fmt.Sprintf(`update cfile set ref=%d where uid=%d`, ref+1, cfileid))
				if err != nil {
					valid = false
					fmt.Println("update cfile error")
					goto CP_VERIFY
				}
			}
		}
	}
CP_VERIFY:
	if !valid {
		u.listen.SendBytes(auth.Int64ToBytes(int64(400)))
	} else {
		u.listen.SendBytes(auth.Int64ToBytes(int64(200)))
	}
}

func (u *cuser) mv(db *sql.DB, command string) {
	//format: mv+uid+newfilename+newpath
	var valid, isdir bool = true, false
	var uid, recordCount, ownerid int
	var err error
	var path, parentPath, originName, queryString string
	var queryRow *sql.Row
	var queryRows *sql.Rows
	var recordList []id_path
	var record id_path
	var subpaths []path_name
	args := generateArgs(command, 4)
	if args == nil {
		valid = false
		goto MV_VERIFY
	}
	uid, err = strconv.Atoi(args[1])
	// check uid, filename, path
	if err != nil || !isFilenameValid(args[2]) || !isPathFormatValid(args[3]) || strings.ToUpper(args[0]) != "MV" {
		valid = false
		goto MV_VERIFY
	}

	// find the original path of file to be moved
	queryRow = db.QueryRow(fmt.Sprintf(`select ownerid, isdir, filename, path from ufile where uid=%d`, uid))
	if queryRow == nil {
		valid = false
		goto MV_VERIFY
	}
	err = queryRow.Scan(&ownerid, &isdir, &originName, &parentPath)
	if err != nil || int64(ownerid) != u.id {
		valid = false
		goto MV_VERIFY
	}
	// check whether all the folders in newpath have been created
	subpaths = generateSubpaths(args[3])
	for _, subpath := range subpaths {
		queryRow = db.QueryRow(fmt.Sprintf(`select count(*) from ufile where ownerid=%d and 
		filename='%s' and path='%s' and isdir=1`, u.id, subpath.u_name, subpath.u_path))
		if queryRow == nil {
			valid = false
			goto MV_VERIFY
		}
		err = queryRow.Scan(&recordCount)
		if err != nil {
			fmt.Println(err.Error())
			valid = false
			goto MV_VERIFY
		}
		if recordCount <= 0 {
			_, err = db.Exec(fmt.Sprintf(`insert into ufile values(null, %d, -1, '%s', '', '%s', 0, 0, 
				'%s', 1, '', 1, '')`, u.id, subpath.u_path, time.Now().Format("2006-01-02 15:04:05"), subpath.u_name))
			if err != nil {
				fmt.Println(err.Error())
				valid = false
			}
		}
	}
	// no matter the record is a file or folder, its filename and path should be changed
	_, err = db.Exec(fmt.Sprintf(`update ufile set path='%s' , filename='%s' where uid=%d`,
		args[3], args[2], uid))
	if err != nil {
		valid = false
		goto MV_VERIFY
	}
	if isdir {
		// if the record is a folder, all the files/folders belong to this folder should be
		// moved to the new folder path
		queryRow = db.QueryRow(fmt.Sprintf(`select count (*) from ufile where path like '%s%%' and ownerid=%d`,
			parentPath+originName+"/", u.id))
		if queryRow == nil {
			valid = false
			goto MV_VERIFY
		}
		err = queryRow.Scan(&recordCount)
		if err != nil {
			valid = false
			goto MV_VERIFY
		}
		queryString = fmt.Sprintf(`select uid, path from ufile where path like '%s%%' and ownerid=%d`,
			parentPath+originName+"/", u.id)
		fmt.Println(queryString)
		queryRows, err = db.Query(queryString)
		defer queryRows.Close()
		if err != nil {
			fmt.Println(err.Error())
			valid = false
			goto MV_VERIFY
		}
		recordList = make([]id_path, 0, recordCount)
		// get all records belong to the dir
		for queryRows.Next() {
			err = queryRows.Scan(&uid, &path)
			if err != nil {
				valid = false
				goto MV_VERIFY
			}
			record.u_id = uid
			record.u_path = path
			recordList = append(recordList, record)
		}
		// update
		for _, record = range recordList {
			_, err = db.Exec(fmt.Sprintf(`update ufile set path='%s' where uid=%d and ownerid=%d`,
				args[3]+args[2]+"/"+record.u_path[len(parentPath)+1+len(originName):], record.u_id, u.id))
			if err != nil {
				valid = false
				goto MV_VERIFY
			}
		}
	}
MV_VERIFY:
	if !valid {
		u.listen.SendBytes(auth.Int64ToBytes(int64(400)))
	} else {
		u.listen.SendBytes(auth.Int64ToBytes(int64(200)))
	}
}

func (u *cuser) touch(db *sql.DB, command string) {
	//touch+filename+path+isdir
	var valid bool = true
	var execString string
	var subpaths []path_name
	var queryRow *sql.Row
	var isdir, recordCount, uid int
	var err error
	args := generateArgs(command, 4)
	if args == nil {
		fmt.Println("args format wrong")
		valid = false
		goto TOUCH_VERIFY
	}
	isdir, err = strconv.Atoi(args[3])
	if err != nil || strings.ToUpper(args[0]) != "TOUCH" || !isFilenameValid(args[1]) ||
		!isPathFormatValid(args[2]) || isdir != 0 && isdir != 1 {
		if err != nil {
			fmt.Println(err.Error())
		}
		//fmt.Println(isdir)
		valid = false
		goto TOUCH_VERIFY
	}
	subpaths = generateSubpaths(args[2])
	for _, subpath := range subpaths {
		queryRow = db.QueryRow(fmt.Sprintf(`select count(*) from ufile where ownerid=%d and 
		filename='%s' and path='%s' and isdir=1`, u.id, subpath.u_name, subpath.u_path))
		if queryRow == nil {
			valid = false
			goto TOUCH_VERIFY
		}
		err = queryRow.Scan(&recordCount)
		if err != nil {
			fmt.Println(err.Error())
			valid = false
			goto TOUCH_VERIFY
		}
		if recordCount <= 0 {
			_, err = db.Exec(fmt.Sprintf(`insert into ufile values(null, %d, -1, '%s', '', '%s', 0, 0, 
				'%s', 1, '', 1, '')`, u.id, subpath.u_path, time.Now().Format("2006-01-02 15:04:05"), subpath.u_name))
			if err != nil {
				fmt.Println(err.Error())
				valid = false //database write error
			}
		}
	}

	execString = fmt.Sprintf(`insert into ufile values(null, %d, -1, '%s',
	 '', '%s', 0, 0, '%s', 1, '', %d, '')`,
		u.id, args[2], time.Now().Format("2006-01-02 15:04:05"), args[1], isdir)
	fmt.Println("touch string:", execString)
	_, err = db.Exec(execString)
	if err != nil {
		fmt.Println("touch err:", err.Error())
		valid = false
	}
	queryRow = db.QueryRow(fmt.Sprintf(`select uid from ufile where path='%s' and filename='%s'`,
		args[2], args[1]))
	if queryRow == nil {
		fmt.Println("return uid faild")
		valid = false
	}
	err = queryRow.Scan(&uid)
	if err != nil {
		fmt.Println("fuck touch error:", err.Error())
		valid = false
	}
TOUCH_VERIFY:
	if !valid {
		fmt.Println("fuck failed")
		codeBytes := auth.Int64ToBytes(int64(400))
		uidBytes := auth.Int64ToBytes(int64(0))
		codeBytes = append(codeBytes, uidBytes...)
		u.listen.SendBytes(codeBytes)
	} else {
		fmt.Println("uid:", uid)
		codeBytes := auth.Int64ToBytes(int64(200))
		uidBytes := auth.Int64ToBytes(int64(uid))
		codeBytes = append(codeBytes, uidBytes...)
		u.listen.SendBytes(codeBytes)
	}
}

func (u *cuser) ls(db *sql.DB, command string) {
	// format: ls+recurssive+path+args
	args := generateArgs(command, 0)
	valid := true
	argAll := "%"
	var queryString string
	var returnString string = fmt.Sprintf("UID%sPATH%sFILE%sCREATED TIME%sSIZE%sSHARED%sMODE",
		conf.SEPERATER, conf.SEPERATER, conf.SEPERATER, conf.SEPERATER, conf.SEPERATER, conf.SEPERATER)
	var uid, ownerid, cfileid, shared, downloaded, cuid, csize, cref int
	var private, isdir bool
	var path, perlink, filename, linkpass, created, cmd5, ccreated string
	var err error
	var ufilelist *sql.Rows
	var recurssive int
	if args == nil || len(args) < 3 || strings.ToUpper(args[0]) != "LS" || !isPathFormatValid(args[2]) {
		valid = false
		goto LS_VERIFY
	}
	recurssive, err = strconv.Atoi(args[1])
	if err != nil || recurssive != 0 && recurssive != 1 {
		valid = false
		goto LS_VERIFY
	}
	for i := 3; i < len(args); i++ {
		if args[i] != "" {
			argAll += (args[i] + "%")
		}
	}
	u.curpath = args[2]
	if recurssive == 0 {
		queryString = fmt.Sprintf(`select uid, ownerid, cfileid, path, perlink, created, shared, downloaded, filename, 
		 private, linkpass, isdir from ufile where ownerid=%d and path='%s' and filename like '%s'`,
			u.id, u.curpath, argAll)
	} else {
		queryString = fmt.Sprintf(`select uid, ownerid, cfileid, path, perlink, created, shared, downloaded, filename, 
		 private, linkpass, isdir from ufile where ownerid=%d and path like '%s%%' and filename like '%s'`,
			u.id, u.curpath, argAll)
	}
	fmt.Println(queryString)
	ufilelist, err = db.Query(queryString)

	if err != nil {
		fmt.Println("1:", err.Error())
		valid = false
		goto LS_VERIFY
	}

	for ufilelist.Next() {
		err = ufilelist.Scan(&uid, &ownerid, &cfileid, &path, &perlink, &created, &shared, &downloaded,
			&filename, &private, &linkpass, &isdir)
		if err != nil {
			fmt.Println("2:", err.Error())
			valid = false
			break
		}
		if cfileid >= 0 {
			tcfile := db.QueryRow(fmt.Sprintf("SELECT uid, md5, size, ref, created FROM cfile where uid='%d'",
				cfileid))
			if tcfile == nil {
				fmt.Println("3: tcfile is nil")
				continue
				//valid = false
				//break
			}
			err = tcfile.Scan(&cuid, &cmd5, &csize, &cref, &ccreated)
			if err != nil {
				fmt.Println("4:", err.Error())
				//valid = false
				//break
				continue
			}
		} else {
			csize = 0
		}
		returnString += fmt.Sprintf("\n%d%s%s%s%s%s%s%s%d%s%d%s", uid, conf.SEPERATER, path, conf.SEPERATER,
			filename, conf.SEPERATER, created, conf.SEPERATER, csize, conf.SEPERATER, shared, conf.SEPERATER)
		if isdir {
			returnString += "DIR"
		} else {
			returnString += "FILE"
		}
	}
LS_VERIFY:
	if !valid {
		u.listen.SendBytes([]byte("error happens when querying files"))
		return
	}
	u.listen.SendBytes([]byte(returnString))
}
