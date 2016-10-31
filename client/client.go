package main

import (
	auth "Cloud/authenticate"
	conf "Cloud/config"
	trans "Cloud/transmit"
	"fmt"
	"net"
	"time"
)

type Client struct {
	remote   trans.Transmitable
	level    uint8
	worklist []trans.Transmitable
	token    []byte
}

func NewClient(level int) *Client {
	c := new(Client)
	c.level = uint8(level)
	c.worklist = nil
	c.token = make([]byte, conf.TOKEN_LENGTH(c.level))
	c.remote = nil
	return c
}

func (c *Client) Connect(ip string, port int) bool {
	if !trans.IsIpValid(ip) || !trans.IsPortValid(port) {
		return false
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		fmt.Println("ERROR: Error dialing", err.Error())
		return false
	}
	init := 0
	buf := make([]byte, conf.TOKEN_LENGTH(c.level)*2)
	chR := time.Tick(1e3)
	for {
		if init >= conf.TOKEN_LENGTH(c.level) {
			break
		}
		<-chR
		length, err := conn.Read(buf[init:])
		if err != nil {
			fmt.Println("ERROR: Transmission Error.")
			return false
		}
		init += length
	}
	c.remote = trans.NewTransmitter(conn, conf.AUTHEN_BUFSIZE, buf[:conf.TOKEN_LENGTH(c.level)])
	copy(c.token, buf[:conf.TOKEN_LENGTH(c.level)])
	return true
}

func (c *Client) Authenticate(username string, passwd string) bool {
	encoded := auth.AesEncode([]byte(username+passwd), c.remote.GetBlock())
	buf := make([]byte, 24+len(encoded))
	copy(buf, auth.Int64ToBytes(int64(len(username+passwd))))
	copy(buf[8:], auth.Int64ToBytes(int64(len(encoded))))
	copy(buf[16:], auth.Int64ToBytes(int64(len(username))))
	copy(buf[24:], encoded)
	_, err := c.remote.GetConn().Write(buf)
	if err != nil {
		return false
	}
	checkInfo, err := c.remote.RecvBytes()
	if err != nil || len(checkInfo) < conf.TOKEN_LENGTH(c.level) {
		return false
	}
	for i := 0; i < conf.TOKEN_LENGTH(c.level); i++ {
		if checkInfo[i] != c.token[i] {
			return false
		}
	}
	c.token = nil
	return true
}

func main() {
	c := NewClient(1)
	if !c.Connect("127.0.0.1", 10087) {
		fmt.Println("CONNECTION WRONG")
	}
	if c.Authenticate(conf.TEST_USERNAME, conf.TEST_PASSWORD) {
		fmt.Println("AUTHENTICATION SUCCESS")
	} else {
		fmt.Println("AUTHENTICATION FAILED")
	}
}
