package main

import (
	"errors"
	"fmt"
	"time"

	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/devices"
	"prisma/tms/log"
	"prisma/tms/spidertracks"
	"prisma/tms/tmsg"
	"prisma/tms/util/ident"
	"prisma/tms/util/units"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
)

var (
	ErrTrackEmpty  = errors.New("empty track")
	ErrSpiderIndex = errors.New("spider index out of range")
)

type SpiderSimple struct {
	IMEI      string
	Time      time.Time
	Latitude  float64
	Longitude float64
	Altitude  int
	Speed     int
	Heading   int
}

func dataToSpider(rawSpiderList spidertracks.Spider, indx int) (SpiderSimple, error) {
	var newSpider SpiderSimple
	if len(rawSpiderList.PosList) <= indx {
		return newSpider, ErrSpiderIndex
	}
	formatTime, err := time.Parse("2006-01-02T15:04:05Z", rawSpiderList.PosList[indx].DateTime)
	if err != nil {
		return newSpider, err
	}
	newSpider = SpiderSimple{
		IMEI:      rawSpiderList.PosList[indx].Esn,
		Time:      formatTime,
		Latitude:  rawSpiderList.PosList[indx].Lat,
		Longitude: rawSpiderList.PosList[indx].Long,
		Altitude:  rawSpiderList.PosList[indx].Altitude,
		Speed:     rawSpiderList.PosList[indx].Speed,
		Heading:   rawSpiderList.PosList[indx].Heading,
	}
	return newSpider, nil
}

func ReturnSimplifiedSpiderList(sysId string, user string, password string, url string, dataCtrDateTime *string) ([]SpiderSimple, error) {
	data, err := spidertracks.BuildRequest(sysId, user, password, url, dataCtrDateTime)
	if err != nil {
		return nil, fmt.Errorf("error when building spider request due to %s", err)
	}
	log.Debug("Spider track response: %+v", data)
	parsedSpiders, err := spidertracks.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("Error when parsing spider data due to %s \n Data: %v", err, data)
	}
	var rawNextList spidertracks.Spider
	for i := len(parsedSpiders.PosList); i == spidertracks.MaxPoints; i = len(rawNextList.PosList) {
		time.Sleep(sleepBetweenRequests)
		nextList, err := spidertracks.CheckCount(sysId, user, password, url, parsedSpiders)
		if err != nil {
			return nil, fmt.Errorf("error when building spider request due to %s", err)
		}
		rawNextList, err := spidertracks.Parse(nextList)
		if err != nil {
			return nil, fmt.Errorf("error when parsing spider data due to %s", err)
		}
		parsedSpiders.PosList = append(parsedSpiders.PosList, rawNextList.PosList...)
	}

	var simplifiedSpiderList []SpiderSimple
	for i := range parsedSpiders.PosList {
		nextSpider, err := dataToSpider(parsedSpiders, i)
		if err != nil {
			return nil, fmt.Errorf("error when simplifying spider list due to %s", err)
		}
		simplifiedSpiderList = append(simplifiedSpiderList, nextSpider)
	}

	if len(parsedSpiders.PosList) > 0 {
		*dataCtrDateTime = parsedSpiders.PosList[len(parsedSpiders.PosList)-1].DataCtrDateTime
	}

	return simplifiedSpiderList, nil
}

func PopulateTrack(spider SpiderSimple) (*tms.Track, error) {
	ingestTime := tms.Now()
	track := &tms.Track{
		Targets:  make([]*tms.Target, 0),
		Metadata: make([]*tms.TrackMetadata, 0),
	}

	point := &tms.Point{
		Latitude:  spider.Latitude,
		Longitude: spider.Longitude,
		Altitude:  float64(spider.Altitude),
	}

	stime, err := ptypes.TimestampProto(spider.Time)
	if err != nil {
		return nil, err
	}
	heading := wrappers.DoubleValue{
		Value: float64(spider.Heading),
	}

	speed := wrappers.DoubleValue{
		Value: units.FromMetersSecondToKnots(float64(spider.Speed)),
	}

	imei := wrappers.StringValue{
		Value: spider.IMEI,
	}

	metadata := tms.TrackMetadata{
		Time:       stime,
		IngestTime: ingestTime,
		Type:       devices.DeviceType_Spidertracks,
		Name:       spider.IMEI,
	}

	target := tms.Target{
		Id:         generateTargetID(),
		Type:       devices.DeviceType_Spidertracks,
		Time:       stime,
		IngestTime: ingestTime,
		Position:   point,
		Heading:    &heading,
		Speed:      &speed,
		Imei:       &imei,
	}

	track.Targets = append(track.Targets, &target)
	track.Metadata = append(track.Metadata, &metadata)
	track.RegistryId = spidertracks.RegistryID(spider.IMEI)
	track.Id = spidertracks.TrackID(spider.IMEI)

	return track, nil
}

func SendTrackToTGWAD(ctxt gogroup.GoGroup, track *tms.Track) error {
	msg, err := tmsg.PackFrom(track)
	if err != nil {
		return err
	}
	if len(msg.Value) == 0 {
		return ErrTrackEmpty
	}

	tmsg.GClient.Send(ctxt, &tms.TsiMessage{
		Destination: []*tms.EndPoint{
			{
				Site: tmsg.GClient.ResolveSite(""),
			},
		},
		WriteTime: tms.Now(),
		SendTime:  tms.Now(),
		Body:      msg,
	})

	return nil
}

func generateTargetID() *tms.TargetID {
	sn := ident.TimeSerialNumber()
	return &tms.TargetID{
		Producer: &tms.SensorID{
			Site: tmsg.GClient.Local().Site,
			Eid:  tmsg.GClient.Local().Eid,
		},
		SerialNumber: &tms.TargetID_TimeSerial{&sn},
	}
}
