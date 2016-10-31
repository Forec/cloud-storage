package main

import (
	cloud "Cloud/server"
)

func main() {
	s := new(cloud.Server)
	s.Run("127.0.0.1", 10087, 1)
}
