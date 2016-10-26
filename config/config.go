package config

const USER_FOLDER = "/home/cloud/store/"

const AUTHEN_BUFSIZE = 1024

const BUFLEN = 4096 * 1024

const MAXTRANSMITTER = 20

func TOKEN_LENGTH(level uint8) int {
	if level <= 1 {
		return 16
	} else if level == 2 {
		return 24
	} else {
		return 32
	}
}
