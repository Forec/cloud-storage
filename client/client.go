/*
author: Forec
last edit date: 2016/11/23
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

package main

import (
	"bufio"
	auth "cloud-storage/authenticate"
	conf "cloud-storage/config"
	trans "cloud-storage/transmit"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type Client struct {
	remote   trans.Transmitable
	info     trans.Transmitable
	level    uint8
	username string
	ip       string
	port     int
	worklist []trans.Transmitable
	token    []byte
}

func NewClient(level int) *Client {
	c := new(Client)
	c.level = uint8(level)
	c.worklist = make([]trans.Transmitable, conf.MAXTRANSMITTER)
	c.token = make([]byte, conf.TOKEN_LENGTH(c.level))
	c.remote = nil
	return c
}

func (c *Client) RemoveWork(t trans.Transmitable) bool {
	for i, ts := range c.worklist {
		if ts == t {
			c.worklist = append(c.worklist[:i], c.worklist[i+1:]...)
			return true
		}
	}
	return false
}

func (c *Client) MessageListening() {
	c.info = c.ThreadConnect(c.ip, c.port)
	if c.info == nil {
		return
	}
	for {
		infos, err := c.info.RecvBytes()
		if err != nil {
			return
		}
		fmt.Println("Server Note: ", string(infos))
	}
}

func (c *Client) ThreadConnect(ip string, port int) trans.Transmitable {
	if !trans.IsIpValid(ip) || !trans.IsPortValid(port) {
		return nil
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		fmt.Println("ERROR: Info dialing error", err.Error())
		return nil
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
			return nil
		}
		init += length
	}
	transmitter := trans.NewTransmitter(conn, conf.BUFLEN, buf[:conf.TOKEN_LENGTH(c.level)])
	temptoken := make([]byte, conf.TOKEN_LENGTH(c.level))
	copy(temptoken, buf[:conf.TOKEN_LENGTH(c.level)])
	encoded := auth.AesEncode([]byte(c.username+string(c.token)), transmitter.GetBlock())
	buf = make([]byte, 24+len(encoded))
	copy(buf, auth.Int64ToBytes(int64(len(c.username+string(c.token)))))
	copy(buf[8:], auth.Int64ToBytes(int64(len(encoded))))
	copy(buf[16:], auth.Int64ToBytes(int64(len(c.username))))
	copy(buf[24:], encoded)
	_, err = transmitter.GetConn().Write(buf)
	if err != nil {
		return nil
	}
	checkInfo, err := transmitter.RecvBytes()
	if err != nil || len(checkInfo) < conf.TOKEN_LENGTH(c.level) {
		return nil
	}
	fmt.Println(checkInfo, temptoken)
	for i := 0; i < conf.TOKEN_LENGTH(c.level); i++ {
		if checkInfo[i] != temptoken[i] {
			return nil
		}
	}
	return transmitter
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
	fmt.Println("token", string(c.token))
	c.ip = ip
	c.port = port
	return true
}

func (c *Client) Authenticate(username string, passwd string) bool {
	md5Ps := string(auth.MD5(passwd))
	fmt.Println("send ,", username+md5Ps)
	encoded := auth.AesEncode([]byte(username+md5Ps), c.remote.GetBlock())
	buf := make([]byte, 24+len(encoded))
	copy(buf, auth.Int64ToBytes(int64(len(username+md5Ps))))
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
	c.username = username
	return true
}

func (c *Client) Run() {
	// read Input
	inputReader := bufio.NewReader(os.Stdin)
	for {
		input, err := inputReader.ReadString('\n')
		if err != nil {
			fmt.Println("ERROR: Failed to get your command.\n")
			continue
		}
		switch {
		case len(input) > 3 && strings.ToUpper(input[:3]) == "GET":
			go c.getFile(input)
		case len(input) > 3 && strings.ToUpper(input[:3]) == "PUT":
			go c.putFile(input)
		case len(input) > 6 && strings.ToUpper(input[:6]) == "UPDATE":
			go c.updateFile(input)
		default:
			c.remote.SendBytes([]byte(input))
			recvB, err := c.remote.RecvBytes()
			if err != nil {
				fmt.Println("ERROR: Failed to receive remote reply")
				c.remote.Destroy()
				if c.info != nil {
					c.info.Destroy()
				}
				return
			}
			fmt.Println(string(recvB))
		}
	}
}

func main() {
	c := NewClient(conf.TEST_SAFELEVEL)
	if !c.Connect(conf.TEST_IP, conf.TEST_PORT) {
		fmt.Println("CONNECTION WRONG")
	} else {
		if c.Authenticate(conf.TEST_USERNAME, conf.TEST_PASSWORD) {
			fmt.Println("AUTHENTICATION SUCCESS")
		} else {
			fmt.Println("AUTHENTICATION FAILED")
			return
		}
	}
	// background message
	go c.MessageListening()
	c.Run()
}

func (c *Client) makeDir(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else {
		err := os.MkdirAll(path, 0711)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	}
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func (c *Client) getFile(input string) {
	// format: get+uid
	var filename string
	if c.info == nil {
		fmt.Println("Network is unstable now.")
		return
	}
	getThread := c.ThreadConnect(c.ip, c.port)
	if getThread == nil {
		fmt.Println("Cannot build thread")
		return
	}
	getThread.SendBytes([]byte(input))
	// verify whether the command is valid
	recv, err := getThread.RecvBytes()
	if err != nil {
		fmt.Println("connection lost1")
		return
	}
	if string(recv) != "VALID" {
		fmt.Println("remote server refuse the request")
		return
	}
	// command is valid, receive how many files will be transmitted
	c.worklist = append(c.worklist, getThread)
	recv, err = getThread.RecvBytes()
	if err != nil || len(recv) != 8 {
		fmt.Println("connection not valid")
		return
	}
	fileCount := auth.BytesToInt64(recv[:8])
	var isdir_int int64
	for i := 0; i < int(fileCount); i++ {
		// receive filename
		recv, err = getThread.RecvBytes()
		if err != nil {
			fmt.Println("connection not valid")
			break
		}
		filename = string(recv)
		fmt.Println("receive filename: ", filename)
		// receive whether the record is a folder or a file
		recv, err = getThread.RecvBytes()
		if err != nil {
			fmt.Println("connection not valid")
			break
		}
		isdir_int = auth.BytesToInt64(recv[:8])
		if isdir_int == 1 {
			// folder
			if !c.makeDir(conf.DOWNLOAD_PATH + filename) {
				fmt.Println("Cannot create folder")
			}
			continue
		}
		file, err := os.OpenFile(conf.DOWNLOAD_PATH+filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			fmt.Println("Cannot Open File ", filename)
			continue
		}
		defer file.Close()
		// transmitting files
		fileWriter := bufio.NewWriter(file)
		if getThread.RecvToWriter(fileWriter) {
			fmt.Println(filename, " has been downloaded")
		} else {
			fmt.Println(filename, " cannot be downloaded")
		}
	}
	c.RemoveWork(getThread)
}

func (c *Client) putFile(input string) {

}

func (c *Client) updateFile(input string) {

}
