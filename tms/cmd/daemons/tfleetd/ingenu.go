package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/devices"
	"prisma/tms/ingenu"
	"prisma/tms/log"
	"prisma/tms/omnicom"
	"prisma/tms/routing"
	"prisma/tms/tmsg"
	client "prisma/tms/tmsg/client"
	"prisma/tms/util/ident"

	"github.com/golang/protobuf/ptypes/wrappers"
)

//MaxIngenuPayloads is the maximum number of payload the ingenu platform will send in one message
const MaxIngenuPayloads = 500

type IngenuRawMessage struct {
	Uplinks []Uplink
}
type Uplink struct {
	MessageId           string
	MessageType         string
	DatagramUplinkEvent DatagramUplinkEvent
}

type DatagramUplinkEvent struct {
	NodeId        string
	ApplicationId uint16
	Timestamp     uint64
	Payload       string
}

// IngenuAPI is for communicating with an endpoint of the rest-service
type IngenuAPI struct {
	token    string // this is used by requests
	username string
	password string
	client   *http.Client
	url      string
}

// NewIngenuAPI returns an instance of IngenuAPI. url should be https?://site.com
func NewIngenuAPI(url, username, password string) *IngenuAPI {
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}
	return &IngenuAPI{
		url:      url,
		username: username,
		password: password,
		client:   &http.Client{},
	}
}

// sends username and password to the endpoint and sets the token from a response
func (ia *IngenuAPI) requestToken() error {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/config/v1/session", ia.url), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Username", ia.username)
	req.Header.Add("Password", ia.password)
	req.Header.Add("Accept", "application/json")

	response, err := ia.client.Do(req)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %d", response.StatusCode)
	}

	defer response.Body.Close()
	var data map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
		return err
	}
	if token, ok := data["token"]; !ok {
		return errors.New("bad scheme")
	} else {
		ia.token, ok = token.(string)
		if !ok {
			return errors.New("bad scheme")
		}
	}
	return nil
}

// If the path stats with https then
// it will be used in the next requests otherwise it will be concatenated with the url
func (ia *IngenuAPI) getConcatenatedURL(path string) (targetUrl string) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		targetUrl = path
	} else {
		targetUrl = ia.url + "/" + path
	}
	return
}

// GetRequest returns a request for make requests
func (ia *IngenuAPI) GetRequest(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, ia.getConcatenatedURL(path), body)
	if err != nil {
		return nil, err
	}
	if ia.token == "" {
		if err := ia.requestToken(); err != nil {
			return nil, err
		}
	}
	req.Header.Add("Authorization", ia.token)
	return req, nil
}

// Return a response from an endpoint using auth
func (ia *IngenuAPI) getResponseWithAuth(method, path string, body io.Reader) (*http.Response, error) {
	req, err := ia.GetRequest(http.MethodGet, ia.getConcatenatedURL(path), nil)
	if err != nil {
		return nil, err
	}
	resp, err := ia.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Get Rest
func (ia *IngenuAPI) Get(path string) ([]byte, error) {
	resp, err := ia.getResponseWithAuth(http.MethodGet, path, nil)
	if err == nil && resp.StatusCode == http.StatusUnauthorized {
		if err := ia.requestToken(); err != nil {
			return nil, err
		}
		resp, err = ia.getResponseWithAuth(http.MethodGet, path, nil)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("unauth")
	}
	return ioutil.ReadAll(resp.Body)
}

// Post Rest
func (ia *IngenuAPI) Post(path string, body *ingenu.DatagramDownlinkRequest) ([]byte, error) {
	datagramDownlinkRequestBuffer, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := ia.getResponseWithAuth(http.MethodPost, path, bytes.NewBuffer(datagramDownlinkRequestBuffer))
	if err == nil && resp.StatusCode == http.StatusUnauthorized {
		if err := ia.requestToken(); err != nil {
			return nil, err
		}
		resp, err = ia.getResponseWithAuth(http.MethodPost, path, bytes.NewBuffer(datagramDownlinkRequestBuffer))
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("unauth")
	}
	return ioutil.ReadAll(resp.Body)
}

//IngenuListener listens on tgwad for *Iridium messages.
type IngenuListener struct {
	tclient      client.TsiClient
	ctxt         gogroup.GoGroup
	ingenuStream <-chan *client.TMsg
}

//OmnicomSendIngenuRestClient listen to tgwad for ingenu.DatagramDownlinkRequest messages and POST them to the iridium server
func (I *IngenuListener) OmnicomSendIngenuRestClient(url string, ia *IngenuAPI, ctxt gogroup.GoGroup) {

	log.Debug("Posting messages to > %s", url)
	for {
		select {
		case <-I.ctxt.Done():
			return
		default:
			tmsg := <-I.ingenuStream
			log.Debug("Type of message received from tgwad %+v", reflect.TypeOf(tmsg.Body))
			report, ok := tmsg.Body.(*ingenu.DatagramDownlinkRequest)
			if !ok {
				log.Warn("Got a non-DatagramDownlinkRequest message in an ingenu stream. Got %v instead", reflect.TypeOf(tmsg.Body))
			} else {
				log.Debug("the POST request body: %+v", report)
				_, err := ia.Post(url, report)
				if err != nil {
					log.Error("error: %+v", err)
					continue
				}
			}
		}
	}
}

//IngenuSend this function listen to tgwad and handles MT data
func IngenuSend(ctxt gogroup.GoGroup, url, urlIngenu, username, password string) {

	waits := &sync.WaitGroup{}
	ia := NewIngenuAPI(urlIngenu, username, password)

	I := &IngenuListener{
		tclient: tmsg.GClient,
		ctxt:    ctxt,
	}
	waits.Add(1)

	ctxt.Go(func() {
		I.handle(url, ia)
		waits.Done()
	})

	waits.Wait()

}

func (I *IngenuListener) handle(url string, api *IngenuAPI) {
	ctxt := I.ctxt.Child("Ingenu listener")
	ctxt.ErrCallback(func(err error) {
		pe, ok := err.(gogroup.PanicError)
		if ok {
			log.Error("Panic in ingenu listener thread: %v\n%v", pe.Msg, pe.Stack)
		} else {
			log.Error("Error in ingenu listener thread: %v", err)
		}
	})

	I.ingenuStream = I.tclient.Listen(ctxt, routing.Listener{
		MessageType: "prisma.tms.ingenu.DatagramDownlinkRequest",
	})

	I.OmnicomSendIngenuRestClient(url, api, ctxt)

}

//OmnicomReceiveIngenuRestClient is a function that will receive omnicom message coming from the RPMA platform.
func OmnicomReceiveIngenuRestClient(ctxt gogroup.GoGroup, url, urlIngenu, username, password string, dev devices.DeviceType) {

	log.Info("Fetching messages from > %s", url)
	var messageID string
	ia := NewIngenuAPI(urlIngenu, username, password)
	lastuplink, err := GetLastMessage(url, ia)
	log.Debug(">>> %v", lastuplink)
	if err != nil {
		log.Error("error: %+v", err)
		return
	}
	//I have the last message so let start the fun of processing this stuff
	log.Debug("Json uplink object is: %+v\n", lastuplink)
	log.Debug("the payload is %+v", lastuplink.DatagramUplinkEvent)
	if lastuplink.MessageType != "DatagramUplinkEvent" {
		processIngenu(ctxt, []Uplink{lastuplink}, dev)
	}
	//get next collection of data and process it
	messageID = lastuplink.MessageId
	log.Debug("the last message id received is %s", messageID)

	for {
		log.Debug("RPMA: trying to get the next messages")
		nextuplinks, err := GetNextIngenuRawMessage(url, messageID, ia)
		if err != nil {
			log.Error("error: %+v", err)
		} else {
			//set the messageID to the last Ingenue MessageId received
			messageID = nextuplinks.Uplinks[len(nextuplinks.Uplinks)-1].MessageId
			go processIngenu(ctxt.Child("process ingenu messages"), nextuplinks.Uplinks, dev)
		}
		//sleep for a while before trying to fetch the next batch
		time.Sleep(10 * time.Second)
	}

}

//GetLastMessage is called when tfleetd starts to retreive the last message receive by the Ingenu servers
func GetLastMessage(url string, api *IngenuAPI) (Uplink, error) {

	link := url

	for {
		var Uplinks IngenuRawMessage
		body, err := api.Get(link)
		if err != nil {
			return Uplink{}, err
		}
		err = json.Unmarshal(body, &Uplinks)
		if err != nil {
			return Uplink{}, err
		}
		// 500 is the maximum number of payload the ingenu platform will send in one message
		// 0 is to make sure that the ingenu message contains at least one payload
		if len(Uplinks.Uplinks) < MaxIngenuPayloads && len(Uplinks.Uplinks) != 0 {
			return Uplinks.Uplinks[len(Uplinks.Uplinks)-1], nil
		} else if len(Uplinks.Uplinks) == 0 {
			time.Sleep(10 * time.Second)
		} else if len(Uplinks.Uplinks) == MaxIngenuPayloads {
			link = url + Uplinks.Uplinks[len(Uplinks.Uplinks)-1].MessageId
		}
	}
}

//GetNextIngenuRawMessage pulls everything after the messageID given to it
func GetNextIngenuRawMessage(url, messageID string, api *IngenuAPI) (IngenuRawMessage, error) {

	var NextIngenuRawMessage IngenuRawMessage
	for {
		var Uplinks IngenuRawMessage
		body, err := api.Get(url + messageID)
		if err != nil {
			return IngenuRawMessage{}, err
		}
		err = json.Unmarshal(body, &Uplinks)
		if err != nil {
			return IngenuRawMessage{}, err
		}
		// 500 is the maximum number of payload the ingenu platform will send in one message
		// 0 is to make sure that the ingenu message contains at least one payload
		if len(Uplinks.Uplinks) < MaxIngenuPayloads && len(Uplinks.Uplinks) != 0 {
			NextIngenuRawMessage.Uplinks = append(NextIngenuRawMessage.Uplinks, Uplinks.Uplinks...)
			return NextIngenuRawMessage, nil
		} else if len(Uplinks.Uplinks) == 0 {
			log.Debug("RPMA: No new messages received, going to retry in a while...")
			time.Sleep(10 * time.Second)
		} else if len(Uplinks.Uplinks) == MaxIngenuPayloads {
			messageID = Uplinks.Uplinks[len(Uplinks.Uplinks)-1].MessageId
			NextIngenuRawMessage.Uplinks = append(NextIngenuRawMessage.Uplinks, Uplinks.Uplinks...)
		}
	}
}

func pullMessages(req *http.Request) ([]byte, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%+v", err)
	}
	defer resp.Body.Close()
	log.Debug("http request status: %+v", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func processIngenu(ctxt gogroup.GoGroup, uplinks []Uplink, dev devices.DeviceType) {

	for _, uplink := range uplinks {

		if uplink.MessageType != "DatagramUplinkEvent" || len(uplink.DatagramUplinkEvent.Payload) == 0 {
			log.Warn("uplink message does not have a DatagramUplinkEvent object or the payload is empty")
		} else {
			data, _ := base64.StdEncoding.DecodeString(uplink.DatagramUplinkEvent.Payload)
			log.Debug("the raw data is to be processed is %+v", data)
			omn, err := omnicom.Parse(data)
			if err != nil {
				log.Error("error: %+v", err)
			} else {

				pbOmni, err := omnicom.PopulateProtobuf(omn)
				if err != nil {
					log.Error("error: %+v", err)
				} else {
					log.Debug("this is the output of it: %+v", pbOmni)
					var Msgs []*tms.TsiMessage
					id := createIDforIngenuMessage(uplink.DatagramUplinkEvent.NodeId)
					registryID := createRegistryIDforIngenuMessage(uplink.DatagramUplinkEvent.NodeId)
					tracks, activity, err := ProcessOmnicom(pbOmni, dev)
					if err != nil {
						log.Error("error: %+v", err)
					} else {
						for _, track := range tracks {
							track.Id = id
							track.RegistryId = registryID
							if len(track.Targets) != 0 {
								track.Targets[0].Nodeid = &wrappers.StringValue{
									Value:                uplink.DatagramUplinkEvent.NodeId,
									XXX_NoUnkeyedLiteral: struct{}{},
									XXX_unrecognized:     nil,
									XXX_sizecache:        0,
								}
							}
						}

						if activity != nil {
							activity.ActivityId = id
							activity.RegistryId = registryID
							activitybody, errpack := tmsg.PackFrom(activity)
							if errpack != nil {
								log.Error("Error when packing activity message body: %v", errpack)
							} else {
								infoMsg := &tms.TsiMessage{
									Destination: []*tms.EndPoint{
										{
											Site: tmsg.GClient.ResolveSite(""),
										},
									},
									WriteTime: tms.Now(),
									SendTime:  tms.Now(),
									Body:      activitybody,
								}
								Msgs = append(Msgs, infoMsg)
							}
						}

						log.Debug("This is the tracks we are sending %+v", tracks)
						Msgs = append(Msgs, PackfromOmnicom(tracks)...)
						for _, Msg := range Msgs {
							log.Debug("track %+v", Msg)
							tmsg.GClient.Send(ctxt, Msg)
						}
					}
				}
				// process the message here
			}
		}
	}
}

//PackfromOmnicom adds tsimessage envelop to omnicom tracks and alerts.
func PackfromOmnicom(tracks []*tms.Track) []*tms.TsiMessage {

	trackMsgs := []*tms.TsiMessage{}

	for _, track := range tracks {

		body, errpack := tmsg.PackFrom(track)

		if errpack != nil {
			log.Error("Error when packing omnicom message body: %v", errpack)
		} else {
			infoMsg := &tms.TsiMessage{
				Destination: []*tms.EndPoint{
					{
						Site: tmsg.GClient.ResolveSite(""),
					},
				},
				WriteTime: tms.Now(),
				SendTime:  tms.Now(),
				Body:      body,
			}
			trackMsgs = append(trackMsgs, infoMsg)
		}

	}

	return trackMsgs

}

func createIDforIngenuMessage(nodeid string) string {
	return ident.
		With("ingenuNodeId", nodeid).
		With("site", tmsg.GClient.Local().Site).
		With("eid", tmsg.GClient.Local().Eid).
		Hash()
}

func createRegistryIDforIngenuMessage(nodeid string) string {
	return ident.With("ingenuNodeId", nodeid).Hash()
}

func createTrackFromOmnicom(datePosition *omnicom.DatePosition, move *omnicom.MV, targetID *tms.TargetID, dev devices.DeviceType) *tms.Track {

	target := &tms.Target{
		Id:         targetID,
		IngestTime: tms.Now(),
		Type:       dev,
	}

	if move != nil {
		target.Heading = &wrappers.DoubleValue{
			Value: float64(move.Heading),
		}
		target.Speed = &wrappers.DoubleValue{
			Value: float64(move.Speed),
		}
	}
	target.Position = &tms.Point{
		Latitude:  float64(datePosition.Latitude),
		Longitude: float64(datePosition.Longitude),
	}

	sensedTime, err := calculateTime(datePosition)
	if err != nil {
		sensedTime = time.Now()
	}
	target.Time = tms.ToTimestamp(sensedTime)

	target.UpdateTime = target.Time

	targets := []*tms.Target{target}

	track := &tms.Track{
		Targets: targets,
	}
	return track
}
