package server

import (
	auth "Cloud/authenticate"
	conf "Cloud/config"
	cs "Cloud/cstruct"
	trans "Cloud/transmit"
	//"database/sql"
	//	"database/sql/driver"
	"fmt"
	//_ "github.com/mattn/go-sqlite3"
	"net"
	"time"
)

type Server struct {
	listener      net.Listener
	loginUserList []cs.User
	//db            *sql.DB
}

func (s *Server) InitDB() bool {
	var err error
	//s.db, err = sql.Open(conf.DATABASE_TYPE, conf.DATABASE_PATH)
	if err != nil {
		return false
	}
	return true
}

func (s *Server) AddUser(u cs.User) {
	s.loginUserList = cs.AppendUser(s.loginUserList, u)
}

func (s *Server) RemoveUser(u cs.User) bool {
	for i, uc := range s.loginUserList {
		if uc == u {
			s.loginUserList = append(s.loginUserList[:i], s.loginUserList[i:]...)
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
		return nil, -1
	}
	srcLength := auth.BytesToInt64(t.GetBuf()[:8])
	encLength := auth.BytesToInt64(t.GetBuf()[8:16])
	nmLength := auth.BytesToInt64(t.GetBuf()[16:24])
	recvL, err = t.RecvUntil(encLength, recvL, chRate)
	if err != nil {
		return nil, -1
	}
	var nameApass []byte
	nameApass, err = auth.AesDecode(t.GetBuf()[24:24+encLength], srcLength, t.GetBlock())
	if err != nil {
		return nil, -1
	}
	// test
	if string(nameApass[:nmLength]) != conf.TEST_USERNAME ||
		string(nameApass[nmLength:]) != conf.TEST_PASSWORD {
		fmt.Println("FAILED")
		return nil, -1
	}

	pc := cs.UserIndexByName(s.loginUserList, string(nameApass[:nmLength]))
	// 该连接由已登陆用户建立
	if pc != nil {
		if pc.GetToken() != string(nameApass[nmLength:]) {
			return nil, -1
		} else {
			if pc.AddTransmit(t) {
				return pc, 1
			} else {
				return nil, -1
			}
		}
	}
	// 该连接来自新用户
	rc := cs.NewCUser(string(nameApass[:nmLength]), "/")
	if rc == nil {
		return nil, -1
	}
	if rc.Verify(string(nameApass[nmLength:])) == true {
		rc.SetListener(t)
		return rc, 0
	}
	return nil, -1
}

func (s *Server) Communicate(conn net.Conn, level uint8) {
	var err error
	s_token := auth.GenerateToken(level)
	length, err := conn.Write([]byte(s_token))
	if length != conf.TOKEN_LENGTH(level) ||
		err != nil {
		return
	}
	mainT := trans.NewTransmitter(conn, conf.AUTHEN_BUFSIZE, s_token)
	rc, mode := s.Login(mainT)
	if rc == nil || mode == -1 {
		return
	}
	if !mainT.SendBytes(s_token) {
		return
	}
	if mode == 0 {
		rc.SetToken(string(s_token))
		s.AddUser(rc)
		rc.DealWithRequests()
		rc.Logout()
		s.RemoveUser(rc)
	} else {
		rc.DealWithTransmission(mainT)
	}
	return
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
