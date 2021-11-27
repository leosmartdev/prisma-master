package main

import (
	"net"
	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/log"
	"prisma/tms/tmsg"
	"time"

	pb "github.com/golang/protobuf/proto"
)

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

// Reporter will be send report messages to tgwad.
func Reporter(after time.Duration, ctx gogroup.GoGroup, env *Env) {
	for range time.NewTicker(after).C {
		report(ctx, env)
	}
}

func report(ctx gogroup.GoGroup, env *Env) {
	m := createMsgBody(env)
	body, err := tmsg.PackFrom(m)
	if err != nil {
		log.Error("Error packing pb.message: %v", err)
		return
	}

	tmsg.GClient.Send(ctx, &tms.TsiMessage{
		Destination: []*tms.EndPoint{
			&tms.EndPoint{
				Site: tmsg.TMSG_HQ_SITE,
			},
		},
		WriteTime: tms.Now(),
		SendTime:  tms.Now(),
		Status:    tms.TsiMessage_BROADCAST,
		Body:      body,
	})
}

func createMsgBody(env *Env) pb.Message {
	tmsInfo := &tms.TnoidInfo{
		TnoidConfiguration: env.TnoidConfiguration,
		HostInfo: &tms.HostInfo{
			Ip:   env.HostIP,
			Name: env.HostName,
		},
		Uptime:       env.Uptime(),
		SentRequests: env.Requests(),
		GotErrors:    env.Errors(),
		LastMsg:      env.GetLastMessage(),
	}

	return tmsInfo
}
