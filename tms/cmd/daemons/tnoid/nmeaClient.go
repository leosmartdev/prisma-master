package main

import (
	"prisma/tms/log"
	"time"

	"bytes"
	"net"
)

type NmeaClient struct {
	address       string
	conn          net.Conn
	resultHandler chan []byte
	queue         []string
	connected     bool
}

func NewNmeaClient(add string) *NmeaClient {
	return &NmeaClient{
		address:       add,
		resultHandler: make(chan []byte),
		queue:         make([]string, 0),
		connected:     false,
	}
}

// nmeaClient connect to the server side
func (c *NmeaClient) Read(addr *net.TCPAddr, dt time.Duration) {
	timer := time.NewTimer(dt)
	for {
		if !c.isConnected() {
			if err := c.connect(addr); err != nil {
				log.Error("nmeaClient couldn't connect to the server side (%s)", err)
				timer.Reset(dt)
				<-timer.C
			}
		}
	}
}

func (nmeaClient *NmeaClient) connect(tcpAddr *net.TCPAddr) (err error) {
	nmeaClient.conn, err = net.DialTCP("tcp", nil, tcpAddr) // Dial the server
	if err == nil {
		nmeaClient.connected = true
		nmeaClient.receive()
	}
	return
}

func (nmeaClient *NmeaClient) receive() {
	pre := make([]byte, 0)

	for {
		data := make([]byte, 1024)
		n, err := nmeaClient.conn.Read(data)

		if err != nil {
			log.Error("Error when reading conn: %v", err)
			break
		} else {

			fields := bytes.Split(data[:n], []byte("\n"))
			for index := 0; index < len(fields)-1; index++ {
				var msg []byte
				if index == 0 {
					msg = append(pre, fields[index]...)
				} else {
					msg = fields[index]
				}

				nmeaClient.resultHandler <- msg[:len(msg)-1] // Push the read messages to the result handler
			}
			pre = fields[len(fields)-1]
		}
	}
	nmeaClient.close()
}

func (nmeaClient *NmeaClient) close() (err error) {
	err = nmeaClient.conn.Close()
	nmeaClient.connected = false
	return
}

// A func to send messages to the server side,  which is not used now
func (nmeaClient *NmeaClient) sendMsg(msg []byte) error {
	_, err := nmeaClient.conn.Write(msg)
	if err != nil {
		nmeaClient.queueMsg(msg)
	}
	return err
}

// Put the messages not sent to the server successfully into a queue, waiting to be sent again
func (nmeaClient *NmeaClient) queueMsg(msg []byte) {
	str := string(msg)
	nmeaClient.queue = append(nmeaClient.queue, str)
}

// flush all the messages in the queue to the server side
func (nmeaClient *NmeaClient) flush() {
	for _, str := range nmeaClient.queue {
		nmeaClient.sendMsg([]byte(str))
	}
}

func (nmeaClient *NmeaClient) GetResult() chan []byte {
	return nmeaClient.resultHandler
}

func (nmeaClient *NmeaClient) isConnected() bool {
	return nmeaClient.connected
}
