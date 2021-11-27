package omngen

import "net"
import "prisma/tms/log"
import "io/ioutil"
import . "github.com/grafov/bcast"

func ReiceiveAlert(addr string, group *Group) {

	member1 := group.Join()

	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		log.Debug("%+v\n", err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Debug("%+v", err)
	}

	conn, err := listener.Accept()
	if err != nil {
		log.Fatal("%+v", err)
	}
	defer func() {
		log.Warn("omn-gen: connection closed to: ", addr)
		conn.Close()
	}()
	for {
		result, err := ioutil.ReadAll(conn)
		if err != nil {
			log.Error("error: %s", err.Error())
		}
		member1.Send(result)
	}

}
