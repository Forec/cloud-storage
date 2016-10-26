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

func client_test() {
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
	t := NewTransmitter(cconn, BUFSIZE, []byte("1234567890123456"))
	t.RecvToWriter(fileWriter)
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
	go client_test()
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
