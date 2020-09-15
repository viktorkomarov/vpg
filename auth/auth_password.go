package auth

import "net"

type authCleartexPassword struct {
	password string
}

func (a authCleartexPassword) Authorize(conn net.Conn) error {

}
