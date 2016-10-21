package transmit

import (
	auth "Cloud/authenticate"
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type transmitter struct {
	conn   net.Conn
	block  cipher.Block
	buf    []byte
	buflen int64
}

func NewTransmitter(tconn net.Conn, tbuflen int64, token []byte) *transmitter {
	t := new(transmitter)
	if tbuflen < 1 {
		t.buflen = 1
	} else {
		t.buflen = tbuflen
	}
	t.buf = make([]byte, t.buflen)
	t.block = auth.NewAesBlock(token)
	return t
}

const buflen = 4096 * 1024

func GetFileSize(path string) (size int64, err error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return -1, err
	}
	fileSize := fileInfo.Size()
	return fileSize, nil
}

func (t *transmitter) SendFromReader(reader bufio.Reader, totalLength int64) bool {
	_, err := conn.Write(TokenEncode([]byte(fmt.Sprintf("%d", totalLength)), ""))
	if err != nil {
		t.Errorf("Transmit: Send Length failed")
		return
	}
	chRate := time.Tick(2e3)
	for {
		<-chRate
		length, err := reader.Read(t.buf)
		if err != nil {
			return false
		}
		if length == 0 {
			return true
		}
		_, err = t.conn.Write([]byte(auth.TokenEncode(t.buf[:length], "")))
		if err != nil {
			return false
		}
	}
}

func (t *transmitter) RecvToWriter(writer bufio.Writer) bool {
	length, err := t.conn.Read(t.buf)
	if err != nil {
		fmt.Println("ERROR: Connection Error.")
		return false
	}
	decodeLength, err := auth.TokenDecode(t.buf[:length], "")
	if err != nil {
		fmt.Println("ERROR: Token Error.")
		return false
	}
	totalLength, err := strconv.Atoi(string(decodeLength))
	if err != nil {
		fmt.Println("ERROR: Header Transmission Error.")
		return false
	}
	chRate := time.Tick(1e3)
	chTimeout := make(chan bool, 1)
	defer close(chTimeout)
	percent := 0
	recvLength := 0
	for {
		<-chRate
		go func() {
			for {
				select {
				case <-chTimeout:
					return
				case <-time.After(2e9):
					return
				}
			}
		}()
		length, err = t.conn.Read(t.buf)
		chTimeout <- true
		if err != nil {
			fmt.Println("ERROR: Transmission Error.")
			return false
		}
		receive, err := auth.TokenDecode(t.buf[:length], "")
		if err != nil {
			fmt.Println("ERROR: Token Error.")
			return false
		}
		outputLength, outputError := writer.Write(receive)
		if outputError != nil || outputLength != length {
			fmt.Println("ERROR: File Write Error.")
			return false
		}
		recvLength = recvLength + length
		//if 100*fileLength/totalFileLength > percent {
		//	percent = 100 * fileLength / totalFileLength
		//	fmt.Printf("Received: %v%%...\n", percent)
		//}
		if recvLength == totalLength {
			writer.Flush()
			fmt.Println("File Transimission Complete.")
			return true
		}
	}
}
