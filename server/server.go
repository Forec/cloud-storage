package server

import (
	auth "Cloud/authenticate"
	conf "Cloud/config"
	cs "Cloud/cstruct"
	trans "Cloud/transmit"
	"fmt"
	"net"
	"time"
)

type Server struct {
	listener      net.Listener
	loginUserList []cs.User
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

func Login(t trans.Transmitable) cs.User {
	chRate := time.Tick(1e3)
	var recvL int64
	var err error
	recvL, err = t.RecvUntil(int64(16), recvL, chRate)
	if err != nil {
		return nil
	}
	srcLength := auth.BytesToInt64(t.GetBuf()[:8])
	encLength := auth.BytesToInt64(t.GetBuf()[8:16])
	nmLength := auth.BytesToInt64(t.GetBuf()[16:24])
	recvL, err = t.RecvUntil(encLength, recvL, chRate)
	if err != nil {
		return nil
	}
	var nameApass []byte
	nameApass, err = auth.AesDecode(t.GetBuf()[24:24+encLength], srcLength, t.GetBlock())
	if err != nil {
		return nil
	}
	rc := cs.NewCUser(string(nameApass[:nmLength]), "/")
	if rc == nil {
		return nil
	}
	if rc.Verify(string(nameApass[nmLength:])) == true {
		rc.SetListener(t)
		return rc
	}
	return nil
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
	rc := Login(mainT)
	if rc == nil {
		return
	}
	rc.SetToken(string(s_token))
	s.AddUser(rc)
	rc.DealWithRequests()
	rc.Logout()
	s.RemoveUser(rc)
	return
}

func (s *Server) Run(ip string, port int, level int) {
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
