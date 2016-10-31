# Cloud Storage（云存储）

> A simple package written in Golang 1.7.3 for building cloud storage server or client, also a task for my 2016 course design. you can use the code under limit of the license in this repository.

## Theory
### File Store
* To capture how this simple cloud storage works, considering the following text graph.
```
--cfile is stored in disk, any reference to the cfile
   is recorded as a ufile for different cusers.
 -------            -------          -------------
| cfile |  <-----  | ufile |  <---  | cuser (id=1)|
 -------            -------          -------------
   /|               -------          -------------
    |<-----------  | ufile |  <---  | cuser (id=2)|
                    -------          -------------
```
* In the text graph above, there are three structs (same to class in Java/C++), the `cuser` is the object of users, each user points to several `ufile`, and each `ufile` can only belongs to one `cuser`. The `ufile` builds a relationship between user and real files. 
* The real files are recorded as `cfile`, each `cfile` may be pointed by many `ufile`, because many users may store a same file in their "different" user spaces. They may name the same file different names, but the content of the file is same exactly, so the server should only store one copy of the real file, and use `ufile` to link the user and the real file. **This is similar to the table used to build many-many relationships in databases**.

### Authentication
The authentication process between server and client is list below:
* Client ask for connection
* Server sends a `token` for client
* Client enciphers the username and password (the password is the md5 or sha1 of plain password exactly, we don't suggest storing plain password in database) by `token` received
* Server verifies the username and password, if valid, return the client that token again, otherwise return another token.
* Client verify whether the two tokens are same, same means success

### Transmission
The transmission protocal in normal data (bytes) is :
* Send 8 bytes without ciphering, specify the total length of stream.
* For each part of stream (since the buf has a limit length, we need to split the stream into several parts), we use the following format to transmit. The first 8 bits are the length of data in plain text form, the second 8 bits is the  length of encoded data + 16, and follows by the encoded data.
```
 --------------------------------------
|  8 bits  |  8 bits  |  encoded data  |
 --------------------------------------
```

* The transmission protocal in authentication is is the following format. The first 8 bits are the length of username+password in plain text form, the second 8 bits are the length of encoded username+password, and the final 8 bits are the length of username.
```
 ------------------------------------------------------------------
|  8 bits  |  8 bits  |  8 bits  |  encoded username and password  |
 ------------------------------------------------------------------
```

### MultiProgramming
Each user has a main transmitter (the object for transmitting data, declared in our package `transmit`) to send/receive basic orders and text messages. When the client asking for transmission of big files, a new transmitter will be created, and it will run in an independent go routine. The `cuser` object stores the transmitters in `worklist`, once transmission is over, they are destroyed.


## Packages
### config
The file `config.go` is in folder `config`, stores the settings and constants for the package.
* `USER_FOLDER` : The prefix for user. You can think it as the folder you store users' files. Maybe we will use database to store those files in the future version.
* `AUTHEN_BUFSIZE` : The bufsize for main transmitters.
* `MAXTRANSMITTER` : The maximum number of transmitters can be created at same time by a user.
* `DATABASE_TYPE` : The type of database you want to use.
* `DATABASE_PATH` : The path of your database.
* `TOKEN_LENGTH(level uint8) int` : Returns the number of bytes in the condition of `level`. The `level` decides how many bits will be used in AES.

### authenticate
`authenticate` : The file `authenticate.go` is in folder `authenticate`, used to encipher/decipher, encode/decode, etc.
* `Base64Encode(plaintext []byte) []byte` : receive the `plaintext` and return it with base64 encoded.
* `Base64Decode(ciphertext []byte) ([]byte, error)` : receive the `ciphertext` encoded by base64, return its plain text and `nil`. If `ciphertext` cannot be decoded by base64, `error` will not be `nil`.
* `TokenEncode(plaintext []byte, token string) []byte` : Use `token` to encipher `plaintext`, this method is not implemented yet.
* `TokenDecode(ciphertext []byte, token string) ([]byte, error)` : Use `token` to decipher `ciphertext`, this method is not implemented yet.
* `NewAesBlock(key []byte) cipher.Block` : The `cipher.Block` is imported from `crypto/cipher`, this method will return a `cipher.Block` built by `key`.
* `AesEncode(plaintext []byte, block cipher.Block) []byte` : This method use `block` to encipher `plaintext`, the enciphering method is CFB. It will return the bytes after enciphered.
* `AesDecode(ciphertext []byte, plainLen int64, block cipher.Block) ([]byte, error)` : This method use `block` to decipher `ciphertext`. To decipher, the block needs the length of the original plain text, which is the `plainLen`. If deciphering is successful, it will return the palin text and `nil`. Otherwise, `error` will not be `nil`.
* `Int64ToBytes(i int64) []byte` : Return 8 bytes expressing an `int64` type, using big endian. For example, `0x00000000 00000055` is the expression of 85.
* `BytesToInt64(buf []byte) int64` : Return a number in base 10, whose value is same to the value expressed by `buf[:8]`.
* `GetRandomString(leng int) string` : Return a random string with length `leng`.
* `MD5(text string) []byte` : Generate and return the MD5 value for `text` in form of `[]byte`.
* `GenerateToken(level uint8) []byte` : Generate a token according to `level`.  When `level` is 1 or lower, the token is 16 bytes (128bits); when `level` is 2, the token is 24 bytes (192 bits); else the token is 32 bytes (256 bits).

### cstruct

### transmit

### server

## TODO
* ~~Authorisation check~~
* ~~File transmission~~
* ~~File system management~~
 * ~~File list items~~
 * ~~Create~~
 * ~~Upload/Download~~
 * Delete/Move/Copy
 * Edit online
 * Share link
* Fork
* ~~Protocal~~
 * Instructions
 * ~~Head length~~
 * ~~Authorisation~~

## Update-Logs
* 2016-10-17: Build repository.
* 2016-10-21: Add basic packages and test cases. Only packages.
* 2016-10-27: Finish basic packages and test cases. Left `Login()` and `Communicate()` to test, next is finish parsing instructions.
* 2016-10-31: Finish authentication part and `Login()`, client can recieve token and pass the test. Write a basic struct of `Communicate()`. Add some functions in `transmit` and `server` these days. Next step is finish `Communicate()` by the protocal.
* 2016-11-1: Add document for current version.

# License
All codes in this repository are licensed under the terms you may find in the file named "LICENSE" in this directory.