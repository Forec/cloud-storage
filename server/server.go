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

package server

import (
	auth "cloud-storage/authenticate"
	conf "cloud-storage/config"
	cs "cloud-storage/cstruct"
	trans "cloud-storage/transmit"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net"
	"strings"
	"time"
)

type Server struct {
	listener      net.Listener
	loginUserList []cs.User
	db            *sql.DB
}

func (s *Server) InitDB() bool {
	var err error
	s.db, err = sql.Open(conf.DATABASE_TYPE, conf.DATABASE_PATH)
	if err != nil {
		return false
	}
	s.db.Exec(`create table cuser (uid INTEGER PRIMARY KEY AUTOINCREMENT,
		username VARCHAR(64), password VARCHAR(128), created DATE)`)
	s.db.Exec(`create table ufile (uid INTEGER PRIMARY KEY AUTOINCREMENT, 
		ownerid INTEGER, cfileid INTEGER, path VARCHAR(256), perlink VARCHAR(128), 
		created DATE, shared INTEGER, downloaded INTEGER, filename VARCHAR(128),
		private BOOLEAN, linkpass VARCHAR(4)), isdir BOOLEAN`)
	s.db.Exec(`create table cfile (uid INTEGER PRIMARY KEY AUTOINCREMENT,
		md5 VARCHAR(32), size INTEGER, ref INTEGER, created DATE)`)
	s.db.Exec(`create table cmessages (mesid INTEGER PRIMARY KEY AUTOINCREMENT,
		targetid INTEGER, sendid INTEGER, message VARCHAR(512), created DATE)`)
	s.db.Exec(`crete table coperations (oprid INTEGER PRIMARY KEY AUTOINCREMENT,
		deletedUFileId INTEGER, deletedUFileName VARCHAR(128), 
		deletedUFilePath VARCHAR(256), relatedCFileId INTEGER, time DATE)`)
	return true
}

func (s *Server) CheckBroadCast() {
	chRate := time.Tick(conf.CHECK_MESSAGE_SEPERATE * time.Second)
	var queryRows *sql.Rows
	var queryRow *sql.Row
	var mesid, uid, messageCount int
	var message, created string
	var err error
	for {
		<-chRate
		for _, u := range s.loginUserList {
			queryRow = s.db.QueryRow(fmt.Sprintf(`select count (*) from cmessages where
				targetid=%d`, u.GetId()))
			if queryRow == nil {
				continue
			}
			err = queryRow.Scan(&messageCount)
			if err != nil {
				continue
			}
			id_list := make([]int, 0, messageCount)
			queryRows, err = s.db.Query(fmt.Sprintf(`select mesid, sendid, message, created
				 from cmessages where targetid=%d`, u.GetId()))
			if err != nil {
				fmt.Println("query error: ", err.Error())
				continue
			}
			for queryRows.Next() {
				err = queryRows.Scan(&mesid, &uid, &message, &created)
				if err != nil {
					fmt.Println("scan error: ", err.Error())
					break
				}
				if s.BroadCast(u, fmt.Sprintf("%d%s%s%s%s", uid, conf.SEPERATER, message,
					conf.SEPERATER, created)) {
					id_list = append(id_list, mesid)
				} else {
					break
				}
			}
			for _, id := range id_list {
				_, err = s.db.Exec(fmt.Sprintf(`delete from cmessages where mesid=%d`, id))
				if err != nil {
					fmt.Println("delete error: ", err.Error())
					continue
				}
			}
		}
	}
}

func (s *Server) BroadCastToAll(message string) {
	for _, u := range s.loginUserList {
		s.BroadCast(u, message)
	}
}

func (s *Server) BroadCast(u cs.User, message string) bool {
	if u.GetInfos() == nil {
		return false
	}
	return u.GetInfos().SendBytes([]byte(message))
}

func (s *Server) AddUser(u cs.User) {
	s.loginUserList = cs.AppendUser(s.loginUserList, u)
}

func (s *Server) RemoveUser(u cs.User) bool {
	for i, uc := range s.loginUserList {
		if uc == u {
			s.loginUserList = append(s.loginUserList[:i], s.loginUserList[i+1:]...)
			return true
		}
	}
	return false
}

func (s *Server) Login(t trans.Transmitable) (cs.User, int) {
	// mode : failed=-1, new=0, transmission=1
	chRate := time.Tick(1e3)
	var recvL int64 = 0
	var err error
	recvL, err = t.RecvUntil(int64(24), recvL, chRate)
	if err != nil {
		fmt.Println("1 error:", err.Error())
		return nil, -1
	}
	srcLength := auth.BytesToInt64(t.GetBuf()[:8])
	encLength := auth.BytesToInt64(t.GetBuf()[8:16])
	nmLength := auth.BytesToInt64(t.GetBuf()[16:24])
	recvL, err = t.RecvUntil(encLength, recvL, chRate)
	if err != nil {
		fmt.Println("2 error:", err.Error())
		return nil, -1
	}
	var nameApass []byte
	nameApass, err = auth.AesDecode(t.GetBuf()[24:24+encLength], srcLength, t.GetBlock())
	if err != nil {
		fmt.Println("decode error:", err.Error())
		return nil, -1
	}
	fmt.Println(string(nameApass[:nmLength]), string(nameApass[nmLength:]))

	pc := cs.UserIndexByName(s.loginUserList, string(nameApass[:nmLength]))
	// 该连接由已登陆用户建立
	if pc != nil {
		fmt.Println("userfined, ", pc.GetUsername())
		fmt.Println("pc token is ", pc.GetToken())
		if pc.GetToken() != string(nameApass[nmLength:]) {
			fmt.Println("token verify error! not valid!")
			return nil, -1
		} else {
			// background message receiver
			if pc.GetInfos() == nil {
				pc.SetInfos(t)
				return pc, 2
			} else {
				// transmission
				return pc, 1
			}
		}
	}
	// 该连接来自新用户
	username := string(nameApass[:nmLength])
	row := s.db.QueryRow(fmt.Sprintf("SELECT * FROM cuser where username='%s'", username))
	if row == nil {
		return nil, -1
	}
	var uid int
	var susername string
	var spassword string
	var screated string

	err = row.Scan(&uid, &susername, &spassword, &screated)
	if err != nil || spassword != strings.ToUpper(string(nameApass[nmLength:])) {
		return nil, -1
	}
	rc := cs.NewCUser(string(nameApass[:nmLength]), int64(uid), "/")
	if rc == nil {
		return nil, -1
	}
	rc.SetListener(t)
	rows, err := s.db.Query(fmt.Sprintf("SELECT cfileid FROM ufile where ownerid=%d", uid))
	if err != nil {
		return nil, -1
	}
	defer rows.Close()
	var cid, size int
	var totalSize int64 = 0
	for rows.Next() {
		err = rows.Scan(&cid)
		if err != nil {
			return nil, -1
		}
		if cid < 0 {
			continue
		}
		row = s.db.QueryRow(fmt.Sprintf("select size from cfile where uid=%d", cid))
		if row == nil {
			continue
		}
		err = row.Scan(&size)
		if err != nil {
			continue
		}
		totalSize += int64(size)
	}
	rc.SetUsed(int64(totalSize))
	return rc, 0
}

func (s *Server) Communicate(conn net.Conn, level uint8) {
	var err error
	s_token := auth.GenerateToken(level)
	length, err := conn.Write([]byte(s_token))
	fmt.Println("send toekn", string(s_token))
	if length != conf.TOKEN_LENGTH(level) ||
		err != nil {
		return
	}
	mainT := trans.NewTransmitter(conn, conf.AUTHEN_BUFSIZE, s_token)
	rc, mode := s.Login(mainT)
	if rc == nil || mode == -1 {
		mainT.Destroy()
		return
	}
	if !mainT.SendBytes(s_token) {
		return
	}
	if mode == 0 {
		rc.SetToken(string(s_token))
		s.AddUser(rc)
		rc.DealWithRequests(s.db)
		rc.Logout()
		s.RemoveUser(rc)
	} else if mode == 1 && mainT.SetBuflen(conf.BUFLEN) && rc.AddTransmit(mainT) {
		rc.DealWithTransmission(s.db, mainT)
	} else if mode != 2 {
		mainT.Destroy()
		fmt.Println("Remote client not valid")
	}
}

func (s *Server) Run(ip string, port int, level int) {
	if !trans.IsIpValid(ip) || !trans.IsPortValid(port) {
		return
	}
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		fmt.Println("test server starting with an error, break down...")
		return
	}
	defer s.listener.Close()
	s.loginUserList = make([]cs.User, 0, conf.START_USER_LIST)
	for {
		sconn, err := s.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting", err.Error())
			continue
		}
		fmt.Println("Rececive connection request from",
			sconn.RemoteAddr().String())
		go s.Communicate(sconn, uint8(level))
	}
}
