package omngen

import "net"
import "prisma/tms/omnicom"
import . "github.com/grafov/bcast"

type UpAlt struct {
	Spr  omnicom.SPR
	Imei string
}

//send raw message over tcp sockets
func Send(raw []byte, IP string) error {

	tcpAddr, err := net.ResolveTCPAddr("tcp4", IP)
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}

	_, err = conn.Write(raw)
	if err != nil {
		return err
	}

	conn.Close()

	return nil
}

func UpdateAlert(group *Group, spr omnicom.SPR, imei string) {

	member := group.Join()
	defer member.Close()

	update := UpAlt{spr, imei}

	member.Send(update)
}
