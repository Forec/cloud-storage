package transmit

import (
	"bufio"
	"os"
	"testing"
)

func TestTransmission(t *testing.T) {
	filename := "test.txt"
	file, err := os.Open(filename)
	if err != nil {
		t.Errorf("Transmit: Cannot Open Testfile")
		return
	}
	fileReader := bufio.NewReader(file)
	totalFileLength, err := getFileSize(filename)
	if err != nil {
		t.Errorf("Transmit: GetFileSize function failed")
		return
	}
	defer file.Close()
	// 128bits aes
	t := NewTransmitter(conn, int64(4096*1024), []byte("1234567890123456"))
	t.SendFromReader(fileReader, int64(totalFileLength))
}
