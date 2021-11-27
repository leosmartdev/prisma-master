package main

import (
	"net"
	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/devices"
	"prisma/tms/libnaf"
	"prisma/tms/log"
	"prisma/tms/routing"
	client "prisma/tms/tmsg/client"
	"prisma/tms/util/ident"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
)

//NafListener is responsible for listening on tgwad for *track messages of "omnicom solar" device type and convvert them to naf format
//TODO: "omnicom solar" device type has to be changed to something more generic like omnicom type
//or we have to add a new type called "omnicom vms"wa
type NafListener struct {
	tclient     client.TsiClient
	ctxt        gogroup.GoGroup
	trackStream <-chan *client.TMsg
	tcpconn     net.Conn
	sshconn     *ssh.Client
}

func (I *NafListener) handle(address string) error {
	ctxt := I.ctxt.Child("Iridium listener")
	ctxt.ErrCallback(func(err error) {
		pe, ok := err.(gogroup.PanicError)
		if ok {
			log.Error("Panic in iridium listener thread: %v\n%v", pe.Msg, pe.Stack)
		} else {
			log.Error("Error in iridium listener thread: %v", err)
		}
	})

	I.trackStream = I.tclient.Listen(ctxt, routing.Listener{
		MessageType: "prisma.tms.MessageActivity",
	})

	log.Debug("Listening for tracks on tgwad")

	err := I.readThread(address, *Conn)
	if err != nil {
		return err
	}

	return nil
}

func (I *NafListener) readThread(address, conn string) error {

	for {
		select {
		case <-I.ctxt.Done():
			switch conn {
			case "tcp":
				I.tcpconn.Close()
			case "sftp":
				I.sshconn.Conn.Close()
			}
			return nil
		default:
			tmsg := <-I.trackStream
			report, ok := tmsg.Body.(*tms.MessageActivity)
			if !ok {
				log.Warn("Got non-track message in track stream. Got %v instead", reflect.TypeOf(tmsg.Body))
				continue
			}

			// dType determines if the device type is an omnicom solar or vms beacon
			dType := (report.Type == devices.DeviceType_OmnicomSolar || report.Type == devices.DeviceType_OmnicomVMS)

			if dType {
				log.Debug("Iridium structure to encode %+v", report)
				str, err := libnaf.EncodeNaf(report)
				if err != nil {
					log.Info("%+v", err)
					continue
				}
				log.Debug("data to forward: %+v", str)
				switch conn {
				case "tcp":
					err = sendtcp([]byte(str), I.tcpconn)
					if err != nil {
						I.tcpconn.Close()
						log.Warn("Will try to open a new conn after 1 second...")
						time.Sleep(time.Second * 1)
						I.tcpconn = opentcpconn(address)
					}
				case "sftp":
					fname := filename(report.Imei.Value)
					err := sendsftp(str, *Dir, fname, I.sshconn)
					if err != nil {
						log.Error("sftp conn error: %+v", err)
					sftpsend:
						for {
							I.sshconn.Conn.Close()
							log.Debug(" Will try to open a new conn after 1 second...")
							time.Sleep(time.Second * 1)

							config := &ssh.ClientConfig{
								User: *Username,
								Auth: []ssh.AuthMethod{
									ssh.Password(*Password),
								},
								HostKeyCallback: ssh.InsecureIgnoreHostKey(),
							}
							I.sshconn = opensshconn(address, config)
							err := sendsftp(str, *Dir, fname, I.sshconn)
							if err == nil {
								break sftpsend
							} else {
								log.Error("%+v", err)
							}
						}
					}
				default:
					log.Error("connection of type %+v is not supported", conn)
					I.ctxt.Done()
				}
			}

		}
	}
}

//Newlistner listens on tgwad and converts mo iridium messages to naf
func Newlistner(ctxt gogroup.GoGroup, client client.TsiClient, waits *sync.WaitGroup, address string) (*NafListener, error) {

	I := &NafListener{
		tclient: client,
		ctxt:    ctxt,
	}

	var config *ssh.ClientConfig

	if *Conn == "sftp" {
		config = &ssh.ClientConfig{
			User: *Username,
			Auth: []ssh.AuthMethod{
				ssh.Password(*Password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		I.sshconn = opensshconn(address, config)
	} else if *Conn == "tcp" {
		I.tcpconn = opentcpconn(address)
	}

	waits.Add(1)

	ctxt.Go(func() {
		if *Conn == "sftp" {
			defer I.sshconn.Conn.Close()
		} else if *Conn == "tcp" {
			defer I.tcpconn.Close()
		}

		for {
			err := I.handle(address)
			if err != nil {
				log.Error("%+v", err)
			}
		}
		waits.Done()
	})

	return I, nil
}

func sendtcp(raw []byte, conn net.Conn) error {

	_, err := conn.Write(raw)
	if err != nil {
		conn.Close()
		return err
	}
	return nil
}

func sendsftp(raw, directory, filename string, conn *ssh.Client) error {

	// open an SFTP sesison over an existing ssh connection.
	sftp, err := sftp.NewClient(conn)
	if err != nil {
		conn.Close()
		return err
	}
	defer sftp.Close()

	log.Debug("Creating file to in remote server")
	// leave your mark
	f, err := sftp.Create(directory + "/" + filename)
	if err != nil {
		return err
	}

	log.Debug("Data forwarda to %+v", conn)
	if _, err := f.Write([]byte(raw)); err != nil {
		return err
	}
	return nil
}

func opentcpconn(address string) net.Conn {

	var err error
	var conn net.Conn

	for {
		conn, err = net.Dial("tcp", address)
		if err != nil {
			log.Warn("Could not dial %+v because of %+v", address, err)
			log.Info("Will try to dial %+v in 1 second ...", address)
			time.Sleep(time.Second * 1)
		} else {
			return conn
		}
	}
}

func opensshconn(address string, config *ssh.ClientConfig) *ssh.Client {

	var err error
	var conn *ssh.Client

	for {
		conn, err = ssh.Dial("tcp", address, config)
		if err != nil {
			log.Warn("Could not dial %+v because of %+v", address, err)
			log.Info("Will try to dial %+v in 1 second...", address)
			time.Sleep(time.Second * 1)
		} else {
			return conn
		}
	}
}

func filename(imei string) string {

	tsn := ident.TimeSerialNumber()
	return imei + strconv.FormatInt(tsn.Seconds, 10) + strconv.FormatInt(int64(tsn.Counter), 10)
}
