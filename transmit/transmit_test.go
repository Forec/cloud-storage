/*
author: Forec
last edit date: 2016/11/09
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
	"fmt"
	"net"
	"os"
	"testing"
	"time"
)

const BUFSIZE int64 = 4096 * 1024
const test_in_filename string = "test_in.exe"
const test_out_filename string = "test_out.exe"

func client_test(t *testing.T) {
	time.Sleep(time.Second)
	cconn, err := net.Dial("tcp", "127.0.0.1:10086")
	if err != nil {
		fmt.Println("ERROR: Error dialing", err.Error())
		return
	}
	defer cconn.Close()
	file, err := os.OpenFile(test_out_filename,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println("ERROR: Cannot Open TestOutFile")
		return
	}
	defer file.Close()
	fileWriter := bufio.NewWriter(file)
	ts := NewTransmitter(cconn, BUFSIZE, []byte("1234567890123456"))
	ts.RecvToWriter(fileWriter)
	recvB, err := ts.RecvBytes()
	if err != nil {
		t.Errorf("ERROR: Cannot receive bytes")
		return
	}
	if string(recvB) != "helloworld" {
		t.Errorf("ERROR: Receive bytes error")
		return
	}
	return
}

func TestTransmission(t *testing.T) {
	file, err := os.Open(test_in_filename)
	if err != nil {
		t.Errorf("Transmit: Cannot Open TestInFile")
		return
	}
	fileReader := bufio.NewReader(file)
	totalFileLength, err := GetFileSize(test_in_filename)
	if err != nil {
		t.Errorf("Transmit: GetFileSize function failed")
		return
	}

	// test server
	listener, err := net.Listen("tcp", "127.0.0.1:10086")
	if err != nil {
		fmt.Println("test server starting with an error, break down...")
		return
	}
	defer listener.Close()
	go client_test(t)
	sconn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting", err.Error())
		return
	}
	fmt.Println("Rececive connection request from",
		sconn.RemoteAddr().String())
	// 128bits aes
	tr := NewTransmitter(sconn, BUFSIZE, []byte("1234567890123456"))
	tr.SendFromReader(fileReader, int64(totalFileLength))
	tr.SendBytes([]byte("helloworld"))
	time.Sleep(time.Second * 2)
	file.Close()

	// verify received
	vfile, err := os.Open(test_out_filename)
	if err != nil {
		t.Errorf("Transmit: Cannot Open TestOutFile")
		return
	}
	defer vfile.Close()
	rfile, err := os.Open(test_in_filename)
	if err != nil {
		t.Errorf("Transmit: Cannot Open TestOutFile")
		return
	}
	defer rfile.Close()

	vfileReader := bufio.NewReader(vfile)
	rfileReader := bufio.NewReader(rfile)

	for {
		rbyte, err1 := rfileReader.ReadByte()
		vbyte, err2 := vfileReader.ReadByte()
		if err1 != nil && err2 != nil {
			break
		} else if err != nil || err2 != nil || rbyte != vbyte {
			t.Errorf("Transmit: Received File Is Not Same With Origin File")
			break
		}
	}

}
