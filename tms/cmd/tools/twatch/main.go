// twatch watches for tmsg messages
package main

import (
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/routing"
	"prisma/tms/tmsg"
	client "prisma/tms/tmsg/client"

	"fmt"
	"prisma/gogroup"

	"github.com/golang/protobuf/jsonpb"
)

var (
	json = jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: false,
		Indent:       "  ",
	}
)

func printMsg(msg *client.TMsg) {
	msgStr, err := json.MarshalToString(&msg.TsiMessage)
	if err != nil {
		log.Error("Unable to marshal message: %v", err)
		return
	}

	bodyStr, err := json.MarshalToString(msg.Body)
	if err != nil {
		log.Error("Unable to marshal message: %v", err)
		return
	}

	fmt.Printf("{\"Type\": \"%v\",\n\"Header\": %v,\n\"Body\":%v\n}\n",
		msg.Type(), msgStr, bodyStr)
}

func main() {
	libmain.Main(tmsg.APP_ID_TWATCH, func(ctxt gogroup.GoGroup) {
		log.Debug("Registring listener for everything...")
		msgChan := tmsg.GClient.Listen(ctxt, routing.Listener{})
		log.Debug("Waiting for messages...")
		for {
			select {
			case <-ctxt.Done():
				return
			case msg := <-msgChan:
				printMsg(msg)
			}
		}
	})
}
