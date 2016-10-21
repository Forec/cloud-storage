package main

import (
	auth "Cloud/authenticate"
	"Cloud/cstruct"
	//"bufio"
	"fmt"
	//"os"
)

func main() {
	u := cstruct.NewCUser(1, "Forec", ".")
	f := cstruct.NewCFile(1, "test.txt", ".", 100)
	uf := cstruct.NewUFile(f, u, "update.txt", "C:\\")
	fmt.Println(uf.GetFilename())
	filename := "testshit.txt"
	encoded := auth.Base64Encode([]byte(filename))
	fmt.Println(string(encoded))
	decoded, _ := auth.Base64Decode(encoded)
	fmt.Println(string(decoded))
	u.AddUFile(uf)
	fmt.Println(u.GetFilelist()[0].GetFilename())
	c := auth.NewAesBlock([]byte("AABCDEFGHIJKLMNOPBCDEFGHIJKLMNOP"))
	ciphertext := string(auth.AesEncode([]byte("Keep Doctor Away  ."), c))
	plaintext, _ := auth.AesDecode([]byte(ciphertext), int64(19), c)
	fmt.Println(ciphertext)
	fmt.Println(string(plaintext))
}
