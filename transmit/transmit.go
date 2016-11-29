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

package transmit

import (
	"bufio"
	auth "cloud-storage/authenticate"
	conf "cloud-storage/config"
	"crypto/cipher"
	"fmt"
	"net"
	"os"
	"regexp"
	"time"
)

type Transmitable interface {
	SendFromReader(*bufio.Reader, int64) bool
	SendBytes([]byte) bool
	RecvToWriter(*bufio.Writer) bool
	RecvBytes() ([]byte, error)
	RecvUntil(int64, int64, <-chan time.Time) (int64, error)
	Destroy()
	SetBuflen(int64) bool
	GetConn() net.Conn
	GetBuf() []byte
	GetBuflen() int64
	GetBlock() cipher.Block
}

type transmitter struct {
	conn    net.Conn
	block   cipher.Block
	buf     []byte
	recvLen int64
	buflen  int64
}

func NewTransmitter(tconn net.Conn, tbuflen int64, token []byte) *transmitter {
	t := new(transmitter)
	t.conn = tconn
	if tbuflen < 1 {
		t.buflen = 1
	} else {
		t.buflen = tbuflen
	}
	t.buf = make([]byte, t.buflen)
	t.block = auth.NewAesBlock(token)
	t.recvLen = 0
	return t
}

func (t *transmitter) SetBuflen(buflen int64) bool {
	if buflen <= t.buflen {
		t.buflen = buflen
		t.buf = t.buf[:t.buflen]
		return true
	} else {
		t.buf = make([]byte, buflen)
		t.buflen = buflen
		return true
	}
}

func GetFileSize(path string) (size int64, err error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return -1, err
	}
	fileSize := fileInfo.Size()
	return fileSize, nil
}

func (t *transmitter) SendBytes(toSend []byte) bool {
	if t.buf == nil || t.conn == nil {
		return false
	}
	totalLength := len(toSend)
	_, err := t.conn.Write(auth.Int64ToBytes(int64(totalLength)))
	if err != nil {
		return false
	}
	chRate := time.Tick(2e3)
	alSend := 0
	var length int
	for {
		<-chRate
		if totalLength == alSend {
			break
		}
		if totalLength-alSend < int(t.buflen/3) {
			length = totalLength - alSend
		} else {
			length = int(t.buflen / 3)
		}
		copy(t.buf[16:], toSend[alSend:alSend+length])
		copy(t.buf, auth.Int64ToBytes(int64(length)))
		encoded := auth.AesEncode(t.buf[16:length+16], t.block)
		copy(t.buf[8:], auth.Int64ToBytes(int64(len(encoded)+16)))
		copy(t.buf[16:], encoded)
		_, err = t.conn.Write(t.buf[:len(encoded)+16])
		if err != nil {
			return false
		}
		alSend += length
	}
	return true
}

func (t *transmitter) SendFromReader(reader *bufio.Reader, totalLength int64) bool {
	if t.buf == nil || t.conn == nil {
		return false
	}
	_, err := t.conn.Write(auth.Int64ToBytes(totalLength))
	if err != nil {
		return false
	}
	sendLength := totalLength
	chRate := time.Tick(2e3)
	var encodeBufLen int64 = t.buflen/3 - 16
	var length int
	for {
		<-chRate
		if sendLength <= 0 {
			return true
		}
		if sendLength >= encodeBufLen {
			length, err = reader.Read(t.buf[16 : 16+encodeBufLen])
		} else {
			length, err = reader.Read(t.buf[16 : 16+sendLength])
		}
		if err != nil {
			return false
		}
		copy(t.buf, auth.Int64ToBytes(int64(length)))
		encoded := auth.AesEncode(t.buf[16:length+16], t.block)
		copy(t.buf[8:], auth.Int64ToBytes(int64(len(encoded)+16)))
		copy(t.buf[16:], encoded)
		_, err = t.conn.Write(t.buf[:len(encoded)+16])
		if err != nil {
			return false
		}
		sendLength -= int64(length)
		if length == 0 {
			return true
		}
	}
}

func (t *transmitter) RecvUntil(until int64, init int64, chR <-chan time.Time) (int64, error) {
	for {
		if init >= until {
			break
		}
		<-chR
		length, err := t.conn.Read(t.buf[init:])
		if err != nil {
			fmt.Println("ERROR: Transmission Error.")
			return -1, err
		}
		init += int64(length)
	}
	return init, nil
}

func (t *transmitter) RecvBytes() ([]byte, error) {
	var err error
	if t.buf == nil || t.conn == nil {
		return nil, err
	}
	chRate := time.Tick(1e3)
	length, err := t.RecvUntil(8, t.recvLen, chRate)
	if err != nil {
		fmt.Println("ERROR: Connection Error.")
		return nil, err
	}
	totalLength := auth.BytesToInt64(t.buf[:8])
	//percent := 0
	var toRecvLength int64 = totalLength
	var plength int64 = 0
	var elength int64 = 0
	var pRecv int64 = length - 8
	copy(t.buf, t.buf[8:length])
	returnBytes := make([]byte, 0, conf.AUTHEN_BUFSIZE)
	for {
		if toRecvLength == int64(0) {
			t.recvLen = pRecv
			return returnBytes, nil
		}
		pRecv, err = t.RecvUntil(int64(16), pRecv, chRate)
		if err != nil {
			return nil, err
		}
		plength = auth.BytesToInt64(t.buf[:8])
		elength = auth.BytesToInt64(t.buf[8:16])
		pRecv, err = t.RecvUntil(elength, pRecv, chRate)
		if err != nil {
			return nil, err
		}
		receive, err := auth.AesDecode(t.buf[16:elength], plength, t.block)
		if err != nil {
			fmt.Println("ERROR: Token Error.")
			return nil, err
		}
		returnBytes = append(returnBytes, receive...)
		toRecvLength -= plength
		copy(t.buf, t.buf[elength:pRecv])
		pRecv -= elength
	}
}

func (t *transmitter) RecvToWriter(writer *bufio.Writer) bool {
	var err error
	if t.buf == nil || t.conn == nil {
		return false
	}
	chRate := time.Tick(1e3)
	length, err := t.RecvUntil(8, t.recvLen, chRate)
	if err != nil {
		fmt.Println("ERROR: Connection Error.")
		return false
	}
	totalLength := auth.BytesToInt64(t.buf[:8])
	//percent := 0
	var recvLength int64 = 0
	var plength int64 = 0
	var elength int64 = 0
	var pRecv int64 = length - 8
	copy(t.buf, t.buf[8:length])
	for {
		if recvLength == int64(totalLength) {
			t.recvLen = pRecv
			writer.Flush()
			fmt.Println("File Transimission Complete.")
			return true
		}
		pRecv, err = t.RecvUntil(int64(16), pRecv, chRate)
		if err != nil {
			fmt.Println("receive head:", err.Error())
			return false
		}
		plength = auth.BytesToInt64(t.buf[:8])
		elength = auth.BytesToInt64(t.buf[8:16])
		pRecv, err = t.RecvUntil(elength, pRecv, chRate)
		if err != nil {
			fmt.Println("receive body:", err.Error())
			return false
		}
		receive, err := auth.AesDecode(t.buf[16:elength], plength, t.block)
		if err != nil {
			fmt.Println("ERROR: Token Error.")
			return false
		}
		outputLength, outputError := writer.Write(receive)
		if outputError != nil || outputLength != int(plength) {
			fmt.Println("ERROR: File Write Error.")
			return false
		}
		recvLength = recvLength + plength
		copy(t.buf, t.buf[elength:pRecv])
		pRecv -= elength
	}
}

func (t *transmitter) Destroy() {
	t.conn.Close()
}

func (t *transmitter) GetBuf() []byte {
	return t.buf
}

func (t *transmitter) GetBuflen() int64 {
	return t.buflen
}

func (t *transmitter) GetBlock() cipher.Block {
	return t.block
}

func (t *transmitter) GetConn() net.Conn {
	return t.conn
}

func IsIpValid(ip string) bool {
	ip_ok, _ := regexp.MatchString(
		"^(25[0-5]|2[0-4]\\d|[0-1]?\\d?\\d)(\\.(25[0-5]|2[0-4]\\d|[0-1]?\\d?\\d)){3}$", ip)
	if !ip_ok {
		return false
	}
	return true
}

func IsPortValid(port int) bool {
	return 0 <= port && port <= 65535
}
