// tmccd handles messages of sarsat beacons.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"prisma/gogroup"
	. "prisma/tms"
	mcc "prisma/tms/cmd/daemons/tmccd/lib"
	"prisma/tms/libmain"
	"prisma/tms/log"
	"prisma/tms/sar"
	"prisma/tms/sit185"
	"prisma/tms/tmsg"
	"reflect"
	"strconv"
	"time"

	"strings"

	"regexp"

	"github.com/fsnotify/fsnotify"
	pb "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/jinzhu/copier"
)

type jtemplate struct {
	MsgNum     string `json:"msg_num"`
	Date       string `json:"date"`
	DateFormat string `json:"date_format"`
	HexaID     string `json:"hex_id"`
	Encoded    string `json:"encoded"`
	DopplerA   string `json:"doppler_a"`
	DopplerB   string `json:"doppler_b"`
	Confirmed  string `json:"confirmed"`
	Doa        string `json:"doa"`
}

var (
	address        string
	protocol       string
	ftpDir         string
	ftpCapture     bool
	capDir         string
	exp            *regexp.Regexp // This exp helps us get rid of mcc headers and footers that are not part of the xml message
	sit185Template string
	trace          = log.GetTracer("server")
	startTime      = time.Now().Unix()
	fields         map[string]string
	regularpaterns []sit185.RegularExpPattern
)

func init() {
	flag.StringVar(&address, "address", ":9999", "-address = IPaddress:Port The address and port to request for mcc messages")
	flag.StringVar(&protocol, "protocol", "ftp", "-protocol = Transport protocol (eg: TCP,FTP,AMHS...)")
	flag.StringVar(&ftpDir, "ftp-dir", "/srv/ftp", "watch this ftp directory for new messages")
	flag.BoolVar(&ftpCapture, "ftp-capture", false, "capture raw ftp data coming from mccs")
	flag.StringVar(&capDir, "capture-dir", "/srv/capture", "captured file are copied to this directory")
	flag.StringVar(&sit185Template, "sit185-template", "/etc/trident/sit185-template.json", "file path to the sit template")

	sit185temp, err := jsonParse(sit185Template)
	if err != nil {
		log.Error("unable to parse sit template: %v", err)
	} else {
		fields = map[string]string{
			"date":        sit185temp.Date,
			"date_format": sit185temp.DateFormat,
			"confirmed":   sit185temp.Confirmed,
			"encoded":     sit185temp.Encoded,
			"doa":         sit185temp.Doa,
			"doppler_a":   sit185temp.DopplerA,
			"doppler_b":   sit185temp.DopplerB,
			"hex_id":      sit185temp.HexaID,
			"msg_num":     sit185temp.MsgNum,
		}
		for key, field := range fields {
			if key != "date_format" {
				rexp, err := regexp.Compile(field)
				if err != nil {
					log.Fatal("unable to compile regular exp failed because of: %v", err)
				}
				regularpaterns = append(regularpaterns, sit185.RegularExpPattern{key, rexp})
			}
		}
	}

}

func main() {
	flag.Parse()

	var fn func(gogroup.GoGroup)

	switch strings.ToUpper(protocol) {
	case "TCP":
		fn = serviceTcp
	case "FTP":
		fn = serviceFtp
	default:
		fmt.Printf("unsupported protocol: %v\n", protocol)
		os.Exit(1)
	}
	libmain.Main(tmsg.APP_ID_TMCCD, fn)
}

// ---------------------------------------------------------------------------
// TCP
// ---------------------------------------------------------------------------

func serviceTcp(ctxt gogroup.GoGroup) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("unable to listen on port: %s", err)
	}
	log.Info("mccd listening on %s\n", address)
	ctxt.Go(func() {
		watchCancel(ctxt)
	})
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("unable to accept connection: %+v", err)
		}
		ctxt.Go(func() {
			handleTcp(ctxt.Child("client"), conn)
		})
	}
}

func handleTcp(ctxt gogroup.GoGroup, conn net.Conn) {
	trace.Logf("Connection accepted from %v", conn.RemoteAddr())
	defer func() {
		conn.Close()
		trace.Log("Connection closed")
	}()
	data, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Error("unable to read from connection: %v", err)
		return
	}
	conn.Write([]byte(fmt.Sprintf("Message received on: %+v", conn.RemoteAddr())))
	err = ingest(ctxt, data, regularpaterns, fields["date_format"])
	if err != nil {
		log.Error("%v", err)
	}
}

func watchCancel(ctxt gogroup.GoGroup) {
	<-ctxt.Done()
	os.Exit(0)
}

// ---------------------------------------------------------------------------
// FTP
// ---------------------------------------------------------------------------

func serviceFtp(ctxt gogroup.GoGroup) {
	// Watch for file system changes in the ftp directory and ingest when
	// a new file appears.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("unable to watch: %v", err)
	}
	defer watcher.Close()

	//create capture directory in the file system in case it does not exist
	if ftpCapture {
		err := os.MkdirAll(capDir, 0755)
		if err != nil {
			log.Error("was not able to create the capturing directory %v: %v", capDir, err)
		}
	}
	// Required to be in a separate go routine according to the FAQ
	ctxt.Go(func() {
		watchEvents(ctxt, watcher)
	})
	// There is a separate username and password for each site that may
	// upload to the system. In the FTP root, there is also a separate
	// directory for each user. Watch all of those directories.
	files, err := ioutil.ReadDir(ftpDir)
	if err != nil {
		log.Fatal("unable to watch %v: %v", ftpDir, err)
	}

	watching := false
	for _, file := range files {
		if file.IsDir() {
			userDir := path.Join(ftpDir, file.Name())
			err = watcher.Add(userDir)
			if err != nil {
				log.Fatal("unable to watch %v: %v", userDir, err)
			}
			log.Info("watching directory: %v", userDir)
			watching = true

			// preprocess existed files
			if err = handleFilesInSubFolders(ctxt, userDir); err != nil {
				log.Error(err.Error())
			}
		}
	}

	// It is an error if there were no user directories found in the
	// FTP root
	if !watching {
		log.Fatal("no directories to watch")
	}
	ctxt.Wait()
}

func handleFilesInSubFolders(ctx gogroup.GoGroup, folder string) error {
	subFiles, err := ioutil.ReadDir(folder)
	if err != nil {
		return fmt.Errorf("unable to watch %v: %v", subFiles, err)
	}
	for _, subFile := range subFiles {
		if subFile.IsDir() {
			continue
		}
		handleTextFile(ctx, path.Join(folder, subFile.Name()))
	}
	return nil
}

func handleTextFile(ctx gogroup.GoGroup, fname string) {
	if strings.HasSuffix(strings.ToLower(fname), ".txt") {
		ingestFromFile(ctx, fname)
	}
}

func watchEvents(ctxt gogroup.GoGroup, watcher *fsnotify.Watcher) {

	for {
		select {
		case event := <-watcher.Events:
			// Remote site will upload the file with a "*.tmp" extension.
			// Once the upload is complete, it will rename it to a ".txt"
			// extension (even though it is XML)// it can be also SIT185 not only XML
			if event.Op&fsnotify.Create == fsnotify.Create {
				handleTextFile(ctxt, event.Name)
			}
		case err := <-watcher.Errors:
			log.Error("error: %v", err)
		}
	}
}

func ingestFromFile(ctxt gogroup.GoGroup, filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Error("unable to read file %v: %v", filename, err)
		return
	}
	if err := ingest(ctxt, data, regularpaterns, fields["date_format"]); err != nil {
		log.Error("unable to ingest file %v: %v", filename, err)
		return
	}

	if ftpCapture {
		delta := time.Now().Unix() - startTime
		capfile := capDir + "/" + strconv.FormatInt(delta, 10)

		if err := copyfile(filename, capfile); err != nil {
			log.Warn("unable to copy ingested file %v to %v: %v", filename, capfile, err)
		}

		if err := os.Remove(filename); err != nil {
			log.Warn("unable to remove ingested file %v: %v", filename, err)
			return
		}

	} else {

		if err := os.Remove(filename); err != nil {
			log.Warn("unable to remove ingested file %v: %v", filename, err)
			return
		}
	}

}

// -----------------------------------------------------------
// Copy file from src to dst
// -----------------------------------------------------------

func copyfile(src, dst string) error {

	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = io.Copy(output, input)
	if err != nil {
		return err
	}

	return nil
}

// ---------------------------------------------------------------------------
// Ingest
// ---------------------------------------------------------------------------

func ingest(ctxt gogroup.GoGroup, data []byte, regularpaterns []sit185.RegularExpPattern, dateformat string) error {

	var msg *sar.SarsatMessage
	var err error
	var track *Track

	if len(data) == 0 {
		return fmt.Errorf("Unable to parse mcc data stream if size %d", len(data))
	}
	log.Info("received message:\n%v", string(data))

	//extract the xml data from the mcc file if any
	xmldata := mcc.XMLExp.Find(data)

	if len(xmldata) != 0 {
		//try to parse xml mcc messages
		//TODO: should this be more flexible ?
		msg, err = mcc.MccxmlParser(xmldata, protocol)
		if err != nil {
			log.Warn("Unable to parse xml mcc message, trying next type: %s", err)
		} else {
			log.Info("Message parsed as XML document")
		}
	}
	if msg == nil && len(regularpaterns) != 0 {
		msg, err = mcc.Sit185Parser(data, protocol, dateformat, regularpaterns)
		if err != nil {
			log.Warn("Unable to parse mcc message as Sit185, trying next type: %s", err)
		} else {
			log.Info("Message parsed as Sit185")
		}
	}
	if msg == nil {
		msg, err = mcc.Sit915Parser(data, protocol)
		if err != nil {
			log.Warn("Unable to parse mcc message as Sit915, trying next type: %s", err)
		} else {
			log.Info("Message parsed as Sit915")
		}
	}
	if msg == nil {
		msg = mcc.DefaultParser(data, protocol)
		log.Warn("Message parsed using default parser")
	}

	trace.Logf("received message: %v", msg)
	me := &SensorID{
		Site: tmsg.GClient.Local().Site,
		Eid:  tmsg.GClient.Local().Eid,
	}

	if msg.GetMessageType() == sar.SarsatMessage_SIT_915 {
		activity, err := mcc.ProcessActivity(msg, me)
		if err != nil {
			return err
		}
		SendActivityToTgwad(ctxt, activity)
	} else {
		track, err = mcc.Process(msg, me)
		if err != nil {
			return err
		}
		SendToTgwad(ctxt, track)
	}
	return nil
}

func SendToTgwad(ctxt gogroup.GoGroup, track *Track) {

	var body *any.Any

	body, err := tmsg.PackFrom(track)
	if err != nil {
		log.Error("tmccd: error in packing track")
	} else {
		infoMsg := TsiMessage{
			Source: tmsg.GClient.Local(),
			Destination: []*EndPoint{
				&EndPoint{
					Site: tmsg.GClient.ResolveSite(""),
				},
			},
			WriteTime: Now(),
			SendTime:  Now(),
			Body:      body,
		}
		trace.Logf("Sending to tgwad: %+v", infoMsg)
		tmsg.GClient.Send(ctxt, &infoMsg)
		trace.Log("Message sent to tgwad")
	}
}

func SendActivityToTgwad(ctxt gogroup.GoGroup, act *MessageActivity) {
	var body *any.Any
	body, err := tmsg.PackFrom(act)
	if err != nil {
		log.Error("error in packing activity: %v", err)
		return
	}
	msg := TsiMessage{
		Source: tmsg.GClient.Local(),
		Destination: []*EndPoint{
			&EndPoint{
				Site: tmsg.GClient.ResolveSite(""),
			},
		},
		WriteTime: Now(),
		SendTime:  Now(),
		Body:      body,
	}
	trace.Logf("Sending to tgwad: %+v", msg)
	tmsg.GClient.Send(ctxt, &msg)
	trace.Log("Message sent to tgwad")
}

func Reflect(i interface{}, msg pb.Message) (*any.Any, error) {

	err := copier.Copy(msg, i)
	if err != nil {
		return nil, fmt.Errorf("could not reflect %v: %v \n", reflect.TypeOf(msg), err)
	}
	body, err := tmsg.PackFrom(msg)
	if err != nil {
		return nil, fmt.Errorf("Error when packing to body: %v \n", err)
	}

	return body, nil
}

func jsonParse(filename string) (jtemplate, error) {

	var jsontype jtemplate

	file, e := ioutil.ReadFile(filename)
	if e != nil {
		return jsontype, e
	}
	err := json.Unmarshal(file, &jsontype)
	if err != nil {
		return jsontype, err
	}

	return jsontype, nil
}
