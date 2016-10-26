package transmit

import (
	auth "Cloud/authenticate"
	"bufio"
	"crypto/cipher"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type Transmitable interface {
	SendFromReader(*bufio.Reader, int64) bool
	RecvToWriter(*bufio.Writer) bool
	RecvUntil(int64, int64, <-chan time.Time) (int64, error)
	Destroy()
	GetBuf() []byte
	GetBuflen() int64
	GetBlock() cipher.Block
}

type transmitter struct {
	conn   net.Conn
	block  cipher.Block
	buf    []byte
	buflen int64
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
	return t
}

func GetFileSize(path string) (size int64, err error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return -1, err
	}
	fileSize := fileInfo.Size()
	return fileSize, nil
}

func (t *transmitter) SendFromReader(reader *bufio.Reader, totalLength int64) bool {
	if t.buf == nil || t.conn == nil {
		return false
	}
	_, err := t.conn.Write(auth.Base64Encode([]byte(fmt.Sprintf("%d", totalLength))))
	if err != nil {
		return false
	}
	chRate := time.Tick(2e3)
	for {
		<-chRate
		length, err := reader.Read(t.buf[16 : t.buflen/3])
		if err != nil {
			return false
		}
		copy(t.buf, auth.Int64ToBytes(int64(length)))
		encoded := auth.AesEncode(t.buf[16:length+16], t.block)
		copy(t.buf[8:], auth.Int64ToBytes(int64(len(encoded)+16)))
		copy(t.buf[16:], encoded)
		_, err = t.conn.Write(t.buf[:length+16])
		if err != nil {
			return false
		}
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

func (t *transmitter) RecvToWriter(writer *bufio.Writer) bool {
	var err error
	if t.buf == nil || t.conn == nil {
		return false
	}
	length, err := t.conn.Read(t.buf)
	if err != nil {
		fmt.Println("ERROR: Connection Error.")
		return false
	}
	decodeLength, err := auth.Base64Decode(t.buf[:length])
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
	//percent := 0
	var recvLength int64 = 0
	var plength int64 = 0
	var elength int64 = 0
	var pRecv int64 = 0
	for {
		pRecv, err = t.RecvUntil(int64(16), pRecv, chRate)
		if err != nil {
			return false
		}
		plength = auth.BytesToInt64(t.buf[:8])
		elength = auth.BytesToInt64(t.buf[8:16])
		pRecv, err = t.RecvUntil(elength, pRecv, chRate)
		if err != nil {
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
		//if 100*fileLength/totalFileLength > percent {
		//	percent = 100 * fileLength / totalFileLength
		//	fmt.Printf("Received: %v%%...\n", percent)
		//}
		if recvLength == int64(totalLength) {
			writer.Flush()
			fmt.Println("File Transimission Complete.")
			return true
		}
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
