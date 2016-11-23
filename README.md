# Cloud Storage（云存储）

> A simple package written in Golang 1.7.3 for building cloud storage server or client, also a task for my 2016 course design. you can use the code under limit of the license in this repository.

[**查看中文版项目介绍**](https://blog.forec.cn/2016/11/12/cloud-storage-system/)  

与项目同步更新的中文指南发布在我的博客 [**专栏**](http://blog.forec.cn/columns/cloud-storage.html) **《云存储系统从入门到放弃》**中。此 README 将在项目完成后统一更新，项目完成前 README 中的内容可能与项目不统一。

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
Once the client login, the authentication process between server and client is list below:
* Client ask for connection
* Server sends a `token` for client
* Client enciphers the username and password (the password is the md5 or sha1 of plain password exactly, we don't suggest storing plain password in database) by `token` received
* Server verifies the username and password, if valid, return the client that token again, otherwise disconnect.
* Client verifis whether the two tokens are same, same means success, then starts sending commands.

If the client has already been online, the new connection built between client and server should deal with transmission. The authentication process is list below:
* Client ask for connection
* Server sends a `token1` for client
* Client enciphers the username and token (the token is the one when logged in) by the token1 received
* Server finds the username is already online, and then verifies token. If valid, the server will return token1 again, otherwise disconnect.
* Client verifis whether the two tokens are same, same means success, then starts transmission.

### Transmission
The transmission protocal in normal data (bytes) is :
* Send 8 bytes without ciphering, specify the total length of stream.
* For each part of stream (since the buf has a limit length, we need to split the stream into several parts), we use the following format to transmit. The first 8 bytes are the length of data in plain text form, the second 8 bytes is the  length of encoded data + 16, and follows by the encoded data.
```
 ----------------------------------------
|  8 bytes  |  8 bytes  |  encoded data  |
 ----------------------------------------
```

* The transmission protocal in authentication is is the following format. The first 8 bytes are the length of username+password in plain text form, the second 8 bytes are the length of encoded username+password, and the final 8 bytes are the length of username.
```
 ---------------------------------------------------------------------
|  8 bytes  |  8 bytes  |  8 bytes  |  encoded username and password  |
 ---------------------------------------------------------------------
```

### MultiProgramming
Each user has a main transmitter (the object for transmitting data, declared in our package `transmit`) to send/receive basic orders and text messages. When the client asking for transmission of big files, a new transmitter will be created, and it will run in an independent go routine. The `cuser` object stores the transmitters in `worklist`, once transmission is over, they are destroyed.


## Packages
### config
The file `config.go` is in folder `config`, stores the settings and constants for the package.
* `USER_FOLDER` : The prefix for user. You can think it as the folder you store users' files. Maybe we will use database to store those files in the future version.
* `STORE_PATH` : The path for saving files.
* `AUTHEN_BUFSIZE` : The bufsize for main transmitters.
* `BUFLEN`: The bufsize for transmission.
* `MAXTRANSMITTER` : The maximum number of transmitters can be created at same time by a user.
* `DATABASE_TYPE` : The type of database you want to use.
* `DATABASE_PATH` : The path of your database.
* `START_USER_LIST`: The server needs to keep an online user list. This is the capacity of the list when the server starts.
* `SEPERATER`: The seperate character for commands.
* `CHECK_MESSAGE_SEPERATE`: The server will check whether there are some messages should be passed from one user to another at a same rate. This value is the duration time for server checking new messages.
* `TOKEN_LENGTH(level uint8) int` : Returns the number of bytes in the condition of `level`. The `level` decides how many bits will be used in AES.

### authenticate
`authenticate` : The file `authenticate.go` is in folder `authenticate`, used to encipher/decipher, encode/decode, etc.
* `Base64Encode(plaintext []byte) []byte` : receive the `plaintext` and return it with base64 encoded.
* `Base64Decode(ciphertext []byte) ([]byte, error)` : receive the `ciphertext` encoded by base64, return its plain text and `nil`. If `ciphertext` cannot be decoded by base64, `error` will not be `nil`.
* ~~`TokenEncode(plaintext []byte, token string) []byte` : Use `token` to encipher `plaintext`, this method is not implemented yet.~~
* ~~`TokenDecode(ciphertext []byte, token string) ([]byte, error)` : Use `token` to decipher `ciphertext`, this method is not implemented yet.~~
* `NewAesBlock(key []byte) cipher.Block` : The `cipher.Block` is imported from `crypto/cipher`, this method will return a `cipher.Block` built by `key`.
* `AesEncode(plaintext []byte, block cipher.Block) []byte` : This method use `block` to encipher `plaintext`, the enciphering method is CFB. It will return the bytes after enciphered.
* `AesDecode(ciphertext []byte, plainLen int64, block cipher.Block) ([]byte, error)` : This method use `block` to decipher `ciphertext`. To decipher, the block needs the length of the original plain text, which is the `plainLen`. If deciphering is successful, it will return the palin text and `nil`. Otherwise, `error` will not be `nil`.
* `Int64ToBytes(i int64) []byte` : Return 8 bytes expressing an `int64` type, using big endian. For example, `0x00000000 00000055` is the expression of 85.
* `BytesToInt64(buf []byte) int64` : Return a number in base 10, whose value is same to the value expressed by `buf[:8]`.
* `GetRandomString(leng int) string` : Return a random string with length `leng`.
* `MD5(text string) []byte` : Generate and return the MD5 value for `text` in form of `[]byte`.
* `GenerateToken(level uint8) []byte` : Generate a token according to `level`.  When `level` is 1 or lower, the token is 16 bytes (128bits); when `level` is 2, the token is 24 bytes (192 bits); else the token is 32 bytes (256 bits).

### cstruct
#### User
You can create a `cuser` struct by factory method `NewCUser(username string, curpath string) *cuser`, however, `cuser` is not exported. Exported interface `User` can be used as the type of `cuser`.
* `GetUsername() string` : return the username.
* `GetId() int64` : return user's id.
* `GetWorkList() []transmit.Transmitable` : return active transmitters currently, not include the main listener.
* ~~`GetFilelist() []UFile` : return the files this user owns.~~
* ~~`GetAbsPath() string` : return the current absolutely path this user in.~~
* `GetToken() string` : return the token used by this user, this method is not recommended, it will be removed in the future versions.
* ~~`GoToUpper()` : change the user's current path.~~
* ~~`GoToPath(path string) bool` : return `true` if the user change the current path to `path` successfully.~~
* ~~`SetPath(path string) bool` : just set the user's current path as `path` without check, always return true.~~
* `SetToken(string) bool` : set the token this user use, this method is not recommended, it will be removed in the future versions.
* `SetListener(transmit.Transmitable) bool` :  set the main listener for this user, the main listener is the listener only used for transmitting messages and commands.
* ~~`Verify(pass string) bool` : check whether  `pass` is this user's password, it will be removed in future versions.~~
* ~~`AddUFile(UFile) bool` : add a `UFile` value to this user's filelist.~~
* ~~`RemoveUFile(UFile) bool` : remove a `UFile` value from this user's filelist.~~
* `AddTransmit(transmit.Transmitable) bool` : add a transmitter to this user's worklist, return `true` if succeed since it can be at most `config.MAXTRANSMITTER` transmitters at one time.
* `RemoveTransmit(transmit.Transmitable) bool` : remove a transmitter from the user's worklist.
* `DealWithRequests()` : the method for this user to listen and accept requests/commands. Rewrite this method to implement your own logical actions.
* `DealWithTransmission(transmit.Transmitable)`: not implement yet, a method for transmit data concurrently. Rewrite it as your wish.
* `Logout()` : logout this user, destroy all the linked connections and clear its memory.

#### CFile
`CFile` is the interface for real file `cfile`. You can use `NewCFile(fid int64, fsize int64) *cfile` to create a new `cfile` instance.
* `GetId() int64` : get the cfile's id.
* `GetTimestamp() time.Time` : get the created time for this cfile.
* `GetSize() int64` : get the size of this cfile.
* `GetRef() int32` : get the reference time of this cfile.
* `SetId(int64) bool` : set this cfile's id.
* `SetSize(int64) bool` : set this cfile's size.
* `AddRef(int32) bool` : add an `int32` value for the reference of this cfile.

#### UFile
`UFile` is the interface for virtual file `ufile`. You can use `NewUFile(upointer *cfile, uowner *cuser, uname string, upath string) *ufile` to create a new `ufile` instance.
* `GetFilename() string` : get this ufile's filename.
* `GetShared() int32` : get the reference time of this ufile.
* `GetDownloaded() int32` : get the downloaded time of this ufile.
* `GetPath() string` : get the path of this ufile.
* `GetPerlink() string` : get a permenant link for this ufile.
* `GetTime() time.Time` : get the created time for this ufile.
* `GetPointer() *cfile` : get the real `cfile` this ufile points to.
* `GetOwner() *cuser` : get the owner of this ufile.
* `IncShared() bool` : inc this ufile's reference time.
* `IncDowned() bool` : inc this ufile's downloaded time.
* `SetPath(string) bool` : set the path of this ufile.
* `SetPerlink(string) bool` : set a permenant  link for this ufile.
* `SetPointer(*cfile) bool` : set a `cfile` as the reality of this ufile.
* `SetOwner(*cuser) bool` : set a `cuser` as the user of this ufile.

#### clist operation
* `func AppendUser(slice []User, data ...User) []User` : append a `User` array to  `slice`, it has better performance than `append` one by one.
* `func AppendUFile(slice []UFile, data ...UFile) []UFile` : same to the function above, appends `ufile`.
* `func AppendTransmitable(slice []trans.Transmitable, data ...trans.Transmitable) []trans.Transmitable ` : same to the function above, appends `transmit.Transmitable`.
* `func UFileIndexByPath(slice []UFile, path string) []UFile` : return a list of `UFile` under `path`.
* `func UFileIndexById(slice []UFile, id int64) []UFile` : return a list of `UFile` points to a same `CFile` whose id is `id`.
* `func UserIndexByName(slice []User, name string) User` : return the `User` named `name`.

### transmit
`Transmitable` is the interface for `transmitter`. You can use `NewTransmitter(tconn net.Conn, tbuflen int64, token []byte) *transmitter` to create a new `transmitter` instance.
* The struct of `transmitter` is :
```golang
type transmitter struct {
	conn   net.Conn
	block  cipher.Block
	buf    []byte
	buflen int64
}
```
* `SendFromReader(*bufio.Reader, length int64) bool` : sends data with length `length` in the provided `bufio.Reader` .
* `SendBytes([]byte) bool` : sends the given bytes.
* `RecvToWriter(*bufio.Writer) bool` : receive data and write them to the given `bufio.Writer`. This method will get the total length automaticly.
* `RecvBytes() ([]byte, error)` : receive bytes and return them as `[]bytes`, if failed, return the error too.
* `RecvUntil(until int64, init int64, <-chan time.Time) (int64, error)` : the current number of bytes received is `init`, and this method will receive until the number arrive `until`. `chan` is a channel for controlling transmitting speed. The method will return the number of bytes at last, if failed, return error too. The buffer used in this method is the buffer in `transmitter` struct.
* `Destroy()` : destroy this transmitter.
* `GetConn() net.Conn` : return the socket of this transmitter.
* `SetBuf(int64) bool`: set a new bufsize for this transmitter.
* `GetBuf() []byte` : return the buffer of this transmitter.
* `GetBuflen() int64` : return the buffer length of this transmitter.
* `GetBlock() cipher.Block` : return the cipher block of this transmitter.

### server
The `Server` is an exported struct, its definition is below. Database will be added in the future versions.
```golang
type Server struct {
	listener      net.Listener
	loginUserList []cs.User
	//db            *sql.DB
}
```
* `InitDB() bool` : initial the database for the server.
* `CheckBroadCast()`: the background function for sending messages between users.
* `AddUser(u cstruct.User)` : add a user to the online list.
* `RemoveUser(u cstruct.User) bool` : remove user `u` from the online list, return `true` if succeed.
* `Login(t transmit.Transmitable) (cstruct.User, int)` : login a user from the transmitter `t`.
* `Communicate(conn net.Conn, level uint8)` : after accept the client's connect request, this method will authenticate the client and start providing service.
* `Run(ip string, port int, level int)` : runs your server in `ip:port`, level means the safety level, discussed before.

### client
The `Client` is an exported struct, its definition is below.
```golang
type Client struct {
	remote   trans.Transmitable
	level    uint8
	worklist []trans.Transmitable
	token    []byte
}
```
* `NewClient(level int) *Client` : returns a new `Client`, level assigns the safety level.
* `Connect(ip string, port int) bool` : let your client connects `ip:port`, if succeed, it will return `true.
* `Authenticate(username string, passwd string) bool` : authenticate the client. If succeed, it will return `true`.
* `ThreadConnect(ip string, port int) trans.Transmitable`: return a new transmitter for this client, used for file transmission.

## TODO
* ~~Authorisation check~~
* ~~File transmission~~
* ~~File system management~~
 * ~~File list items~~
 * ~~Create~~
 * ~~Upload/Download~~
 * ~~Delete/Move/Copy~~
 * Edit online
 * Share link
* Fork
* ~~Protocal~~
 * ~~Instructions~~
 * ~~Head length~~
 * ~~Authorisation~~
* Project
 * logic
 * ~~database~~

## Update-Logs
* 2016-10-17: Build repository.
* 2016-10-21: Add basic packages and test cases. Only packages.
* 2016-10-27: Finish basic packages and test cases. Left `Login()` and `Communicate()` to test, next is finish parsing instructions.
* 2016-10-31: Finish authentication part and `Login()`, client can recieve token and pass the test. Write a basic struct of `Communicate()`. Add some functions in `transmit` and `server` these days. Next step is finish `Communicate()` by the protocal.
* 2016-11-1: Add document for part of the current version.
* 2016-11-2: Add document for current version.
* 2016-11-9: Migrate the data source from memory to database, fix errors in `transmit`, add two logic actions in `DealWithRequests()`. The basic frame is  already built. To make a simple complete cloud storage server, the only thing need to be done is finish `DealWithRequests()`. Some problems left: the struct is not wrapped well, the inner implement is opened to users. Part of the `server` and `cstruct` should be reconstructed  after the demo finished.
* 2016-11-12: Fix many errors found in `transmit` again. Finish basic commands for `cuser`, including `touch`, `cp`, `mv`, `fork`, `rm` and `send`, which used for sending messages to other users. The client can download files from server now. There are many small changes in  `client`, `server`. I also seperate some functions from the file `cuser.go`. After finishing all the logical operations, the struction should be reconstructed to make `user` and `server` independant.
* 2016-11-23: Now the server and client can contact and use basic commands. The basic functions including transmission for client will be finished before December.

# License
All codes in this repository are licensed under the terms you may find in the file named "LICENSE" in this directory.