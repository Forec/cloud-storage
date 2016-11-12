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

package config

const USER_FOLDER = "/home/cloud/store/"

const STORE_PATH = "G:\\Cloud\\"

const DOWNLOAD_PATH = "G:\\Cloud\\"

const CLIENT_VERSION = "Windows"

const AUTHEN_BUFSIZE = 1024

const BUFLEN = 4096 * 1024

const MAXTRANSMITTER = 20

const DATABASE_TYPE = "sqlite3"

const DATABASE_PATH = "work.db"

const START_USER_LIST = 10

const TEST_USERNAME = "forec"

const TEST_PASSWORD = "TESTTHISPASSWORD"

const TEST_SAFELEVEL = 1

const TEST_IP = "127.0.0.1"

const TEST_PORT = 10087

const SEPERATER = "+"

const CHECK_MESSAGE_SEPERATE = 5

func TOKEN_LENGTH(level uint8) int {
	if level <= 1 {
		return 16
	} else if level == 2 {
		return 24
	} else {
		return 32
	}
}
