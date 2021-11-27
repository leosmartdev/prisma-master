package spidertracks

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"prisma/tms/log"
	"strings"
	"time"

	"prisma/tms/tmsg"
	"prisma/tms/util/ident"
)

var ErrUnauthorized = errors.New("unauthorized")

const (
	MaxPoints         = 1000
	version           = "2.23"
	TimeLayout        = "2006-01-02T15:04:05Z"
	RequestForTimeAgo = 15
)

type SpiderRequest struct {
	XMLName    xml.Name   `xml:"data"`
	Xmlns      string     `xml:"xmlns,attr"`
	SysId      string     `xml:"sysId,attr"`
	RptTime    string     `xml:"rptTime,attr"`
	Version    string     `xml:"version,attr"`
	MsgRequest MsgRequest `xml:"msgRequest"`
}

type MsgRequest struct {
	MsgTo    string `xml:"to,attr"`
	MsgFrom  string `xml:"from,attr"`
	MsgType  string `xml:"msgType,attr"`
	Subject  string `xml:"subject,attr"`
	DateTime string `xml:"dateTime,attr"`
	Body     string `xml:"body"`
}

func BuildRequest(sysId string, user string, password string, url string, body *string) (string, error) {
	var request SpiderRequest
	tn := time.Now()

	request.Xmlns = "https://aff.gov/affSchema"
	request.SysId = sysId
	request.RptTime = tn.Format(TimeLayout)
	request.Version = version
	request.MsgRequest.MsgTo = "spidertracks"
	request.MsgRequest.MsgFrom = sysId
	request.MsgRequest.MsgType = "Data Request"
	request.MsgRequest.Subject = "Async"
	request.MsgRequest.DateTime = tn.Format(TimeLayout)
	request.MsgRequest.Body = *body

	requestxml, err := xml.Marshal(&request)
	if err != nil {
		return "", err
	}
	newstring := string(requestxml)
	log.Debug("This is the xml request sent to spider tracks: %s", newstring)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(newstring))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/xml; charset=utf-8")
	req.SetBasicAuth(user, password)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return "", ErrUnauthorized
		}
		return "", fmt.Errorf("status error: %v", resp.StatusCode)
	}
	data, err := ioutil.ReadAll(resp.Body)
	return string(data), err
}

func NextRequest(sysId string, user string, password string, url string, timeLast string) (string, error) { //if points returned > 1000
	var request SpiderRequest
	timeForm := time.Now().Format(TimeLayout)

	request.Xmlns = "https://aff.gov/affSchema"
	request.SysId = sysId
	request.RptTime = timeForm
	request.Version = "2.23"
	request.MsgRequest.MsgTo = "spidertracks"
	request.MsgRequest.MsgFrom = sysId
	request.MsgRequest.MsgType = "Data Request"
	request.MsgRequest.Subject = "Async"
	request.MsgRequest.DateTime = timeForm
	request.MsgRequest.Body = timeLast

	requestxml, err := xml.Marshal(&request)
	newstring := string(requestxml)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(newstring))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/xml; charset=utf-8")
	req.SetBasicAuth(user, password)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status error: %v", resp.StatusCode)
	}
	data, err := ioutil.ReadAll(resp.Body)
	return string(data), err
}

func CheckCount(sysId string, user string, password string, url string, List Spider) (string, error) {
	var SecondSpiderList string
	count := len(List.PosList)
	if count == MaxPoints {
		return NextRequest(sysId, user, password, url, List.PosList[count-1].DateTime)
	}
	return SecondSpiderList, nil
}

/*
xml post (for reference):::

<?xml version="1.0" encoding="utf-8"?>
<data xmlns="https://aff.gov/affSchema" sysId="Orolia" rptTime="2018-05-18T11:12:44.103Z" version="2.23">
<msgRequest to="spidertracks" from="Orolia" msgType="Data Request" subject="Async" dateTime="2018-05-18T10:06:26Z"> <body>2018-05-18T10:01:00Z</body>
</msgRequest> </data>

*/

func TrackID(imei string) string {
	return ident.
		With("imei", imei).
		With("site", tmsg.GClient.Local().Site).
		With("eid", tmsg.GClient.Local().Eid).
		Hash()
}

func RegistryID(imei string) string {
	return ident.With("imei", imei).Hash()
}
