package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"sync"
	"time"

	"prisma/gogroup"
	"prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/devices"
	"prisma/tms/envelope"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/omnicom"
	"prisma/tms/tmsg"
	"prisma/tms/util/ident"
	omni "prisma/tms/util/omnicom"
	"prisma/tms/ws"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	pb "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

var root = rand.New(rand.NewSource(time.Now().UnixNano()))

type omnicomStage struct {
	n         Notifier
	mutex     sync.RWMutex
	ctxt      gogroup.GoGroup
	miscDb    db.MiscDB
	clt       *mongo.MongoClient
	publisher *ws.Publisher
	// initilized is bool variable that gets set to true when the init function is done
	// only when initilized is true, we can analyze the stage
	initialized bool
}

func newOmnicomStage(n Notifier, ctxt gogroup.GoGroup, client *mongo.MongoClient, publisher *ws.Publisher) *omnicomStage {
	return &omnicomStage{
		n:           n,
		ctxt:        ctxt,
		miscDb:      mongo.NewMongoMiscData(ctxt, client),
		clt:         client,
		publisher:   publisher,
		initialized: false,
	}
}

func (s *omnicomStage) init(ctxt gogroup.GoGroup, client *mongo.MongoClient) error {
	log.Info("initial activity loading")
	activitydb := s.miscDb
	activityupdates := activitydb.GetPersistentStream(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: "prisma.tms.MessageActivity",
		},
		Ctxt: s.ctxt,
	}, nil, nil)
	go func() {
		for {
			select {
			case update, ok := <-activityupdates:
				if !ok {
					log.Error("channel was closed")
					return
				}
				if update.Contents != nil {
					err := s.analyzeActivity(update)
					if err != nil {
						log.Error("error analyzing activity update: %v", err)
					}
				}
			case <-ctxt.Done():
				return
			}
		}
	}()
	s.initialized = true
	return nil
}

func (s *omnicomStage) analyzeActivity(update db.GoGetResponse) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	devicedb := mongo.NewMongoDeviceDb(s.ctxt, s.clt)
	tranDb := mongo.NewTransmissionDb(s.ctxt, s.clt)
	mcDb := mongo.NewMulticastDb(s.ctxt)
	activity, ok := update.Contents.Data.(*tms.MessageActivity)
	if !ok {
		return fmt.Errorf("Not an activity stream %+v", update.Contents)
	}
	//Check if the activity has an iridium Mobile Terminated Confirmation payload
	if activity.GetMtc() != nil {
		mtc := activity.GetMtc()
		if activity.RequestId == "" {
			return fmt.Errorf("received an %+v with no request id", mtc)
		}
		log.Debug("Got an Mtc activity request Id %+v", activity.RequestId)
		err := tranDb.Status(activity.RequestId, tms.Transmission_Partial, 102)
		if err != nil {
			return err
		}
		stat := mtc.MT_Message_Status
		//Packet status update
		// if status is in [0,50) then iridium gateway received payload so packet status is partial
		// default packets did not reach gateway so status is failed
		if stat >= 0 && stat < 50 {
			tr, err := tranDb.PacketStatusSingle(activity.RequestId, tms.Transmission_Partial.String(), 102)
			if err != nil {
				return err
			}
			err = mcDb.UpdateTransmission(s.ctxt, tr)
			if err != nil {
				return err
			}
		} else {
			tr, err := tranDb.PacketStatusSingle(activity.RequestId, tms.Transmission_Failure.String(), 400)
			if err != nil {
				return err
			}
			tranDb.ClearMessageId(activity.RequestId)
			err = mcDb.UpdateTransmission(s.ctxt, tr)
			if err != nil {
				return err
			}
			notifyTransmissionFailure(s.n, tr)
		}
	}
	//Check if the activity has an Omnicom payload
	if activity.GetOmni() != nil {

		omni := activity.GetOmni()
		log.Debug("Got an Omni activity %+v", omni)
		if activity.RequestId != "" {
			tr, err := tranDb.PacketStatusSingle(activity.RequestId, tms.Transmission_Success.String(), 200)
			if err != nil {
				log.Error(err.Error()+"%v", tr)
			}
			// since one packet update transmission
			tranDb.Status(activity.RequestId, tms.Transmission_Success, 200)
			tranDb.ClearMessageId(activity.RequestId)
			// since one transmission update multicast
			err = mcDb.UpdateTransmission(s.ctxt, tr)
			if err != nil {
				log.Error(err.Error()+"%v", tr)
			}
		} else {
			log.Debug("no RequestId for %v", activity)
		}
		if omni.GetAr() != nil {

			report := omni.GetAr()

			if report.Assistance_Alert != nil && report.Assistance_Alert != (&omnicom.AssistanceAlert{}) {
				log.Info("Got an assistance alert %v", report.Assistance_Alert)
				hasAssistance := report.Assistance_Alert.Alert_Status == 1
				isAssistance := report.Assistance_Alert.Current_Assistance_Alert_Status == 1
				if hasAssistance {
					notice := omnicomNotice(moc.Notice_OmnicomAssistance, activity, moc.Notice_Alert)
					s.n.Notify(notice, isAssistance)
				}
			}
			if report.Test_Mode != nil && report.Test_Mode != (&omnicom.TestMode{}) {
				hasTest := report.Test_Mode.Alert_Status == 1
				isTest := report.Test_Mode.Current_Test_Mode_Status == 1
				if hasTest {
					log.Info("Got test alert %d %v", isTest, report)

					//Send test ack mode message
					tran, err := SendTestAck(s.ctxt, activity.Imei.Value)
					if err != nil {
						return nil
					}

					//Upsert Transmission created by SendTestAck
					tran.Packets = append(tran.Packets, &tms.Packet{Name: omnicom.OmnicomConfiguration_Action_name[int32(omnicom.OmnicomConfiguration_TestModeAck)], State: 1, Status: &tms.ResponseStatus{Code: 102}})
					tran.Status = &tms.ResponseStatus{Code: 102}

					//check if device exists
					dev, err := devicedb.FindNet(activity.Imei.Value)
					if mgo.ErrNotFound != err {
						// Change destination from network to device
						tran.Destination = &tms.EntityRelationship{Id: dev.Id, Type: dev.GetType()}
						//Upsert transmission
						err = tranDb.Update(&tran)
						if err != nil {
							return err
						}
						return nil
					}

					//Insert transmission for an unregistred device
					err = tranDb.Update(&tran)
					if err != nil {
						return err
					}

					//Registration request
					tran, err = RequestSpecificMessage(0x03, s.ctxt, activity.Imei.Value, dev)
					if err != nil {
						return err
					}

					tran.Packets = append(tran.Packets, &tms.Packet{Name: omnicom.OmnicomConfiguration_Action_name[int32(omnicom.OmnicomConfiguration_RequestGlobalParameters)], State: 1, Status: &tms.ResponseStatus{Code: 102}})
					tran.Status = &tms.ResponseStatus{Code: 102}
					err = tranDb.Update(&tran)
					if err != nil {
						return err
					}
				}
			}

		}
		if omni.GetGp() != nil {
			report := omni.GetGp()
			log.Debug("Got a global parameter %+v", report)
			// OmnicomConfiguration
			omAny, err := ptypes.MarshalAny(&omnicom.OmnicomConfiguration{
				PositionReportingInterval: report.Position_Reporting_Interval.ValueInMn,
			})
			if err != nil {
				return err
			}
			// device configuration
			gpAny, err := ptypes.MarshalAny(report)
			if err != nil {
				return err
			}
			//This would be either omnicom-vms or omnicome-solar
			deviceType := devices.DeviceType_name[int32(activity.Type)]
			device := &moc.Device{
				DeviceId: strconv.FormatUint(uint64(report.Beacon_ID), 10),
				Type:     deviceType,
				Networks: []*moc.Device_Network{
					{
						SubscriberId: string(report.IRI_IMEI),
						Type:         "satellite-data",
						ProviderId:   "iridium",
						RegistryId:   ident.With("imei", string(report.IRI_IMEI)).Hash(),
					},
					{
						SubscriberId: string(report.G3_IMEI),
						Type:         "cellular-data",
						ProviderId:   "3g",
						RegistryId:   ident.With("imei", report.G3_IMEI).Hash(),
					},
				},
				Configuration: &moc.DeviceConfiguration{
					Configuration: omAny,
					Original:      gpAny,
					LastUpdate:    tms.Now(),
				},
			}
			// find device
			foundDevice, err := devicedb.FindByDevice(device)
			if err == nil {
				//device found, update network and configuration
				foundDevice.RegistryId = device.RegistryId
				foundDevice.Networks = device.Networks
				foundDevice.Configuration = device.Configuration
				err = devicedb.Update(s.ctxt, foundDevice)
				if err != nil {
					log.Error(err.Error())
				}
			} else {
				// if no device found then create one
				err = devicedb.Insert(device)
				if err != nil {
					log.Error(err.Error())
				}
			}
			// check vessel
			vesseldb := mongo.NewMongoVesselDb(s.ctxt)
			device.Id = "" // workaround for device and vessel.device being out-of-sync
			vessel, err := vesseldb.FindByDevice(s.ctxt, device)
			if err != nil {
				log.Debug("vessel %v", err.Error())
				return nil // vessel is not required and should not return error
			}
			s.updateDevice(vessel, device, vesseldb)
		}
		if omni.GetHpr() != nil {
			//if omnicom history position report encapsulates only one data report
			//it mean that the HPR is a confirmation for reporting unit interval change of the beacon
			// when that happens we need to update device configuration with the new reporting interval
			report := omni.GetHpr()
			log.Debug("Got a history position report: %+v", report)

			if report.GetCount_Total_Data_Reports() > 0 {
				//check if device exists
				dev, err := devicedb.FindNet(activity.Imei.Value)
				if mgo.ErrNotFound != err {
					//update configuration by including the new reporting period
					if dev.Configuration == nil {
						dev.Configuration = &moc.DeviceConfiguration{}
					}
					// OmnicomConfiguration
					omAny, err := ptypes.MarshalAny(&omnicom.OmnicomConfiguration{
						PositionReportingInterval: report.Data_Report[0].Period,
					})
					if err != nil {
						return err
					}
					dev.Configuration.Configuration = omAny
					dev.Configuration.LastUpdate = tms.Now()
					// update device
					err = devicedb.UpsertDeviceConfig(dev.DeviceId, dev.Type, dev.Configuration)
					if err != nil {
						log.Error(err.Error())
					}
					// check vessel
					vesseldb := mongo.NewMongoVesselDb(s.ctxt)
					dev.Id = "" // workaround for device and vessel.device being out-of-sync
					vessel, err := vesseldb.FindByDevice(s.ctxt, dev)
					if err != nil {
						log.Debug("vessel %v", err.Error())
						return nil // vessel is not required and should not return error
					}
					s.updateDevice(vessel, dev, vesseldb)
				}
			}
		}
		if omni.GetGa() != nil {
			zoneDb := mongo.NewMongoZoneMiscData(s.miscDb)
			report := omni.GetGa()
			//retrieve device
			dev, err := devicedb.FindNet(activity.Imei.Value)
			if err != nil {
				if err == mgo.ErrNotFound {
					//Registration request
					tran, err := RequestSpecificMessage(0x03, s.ctxt, activity.Imei.Value, dev)
					if err != nil {
						return err
					}
					tran.Packets = append(tran.Packets, &tms.Packet{Name: omnicom.OmnicomConfiguration_Action_name[int32(omnicom.OmnicomConfiguration_RequestGlobalParameters)], State: 1, Status: &tms.ResponseStatus{Code: 102}})
					tran.Status = &tms.ResponseStatus{Code: 102}
					err = tranDb.Update(&tran)
					if err != nil {
						return err
					}
				}
				return fmt.Errorf("Geofence upload ack could not be processed because of: %+v", err)
			}
			//retrieve zone
			mz, err := zoneDb.GetOne(report.GetGEO_ID())
			if err != nil {
				return err
			}
			//evaluate error type code to determine actions
			switch report.Error_Type {
			case 0:
				if report.Msg_ID%2 == 0 {
					// append entity relationship for zones
					mz.Entities = append(mz.Entities, &moc.EntityRelationship{
						Type:         pb.MessageName(dev),
						Id:           dev.DeviceId,
						UpdateTime:   tms.Now(),
						Relationship: moc.EntityRelationship_UPLOAD,
					})
					// append entity relationship for device
					if dev.GetConfiguration() != nil {
						dev.Configuration.Entities = append(dev.Configuration.Entities, &moc.EntityRelationship{
							Type:         pb.MessageName(mz),
							Id:           strconv.FormatInt(int64(mz.ZoneId), 10),
							UpdateTime:   tms.Now(),
							Relationship: moc.EntityRelationship_UPLOAD,
						})
					} else {
						dev.Configuration = &moc.DeviceConfiguration{
							Entities: []*moc.EntityRelationship{{
								Type:         pb.MessageName(mz),
								Id:           strconv.FormatInt(int64(mz.ZoneId), 10),
								UpdateTime:   tms.Now(),
								Relationship: moc.EntityRelationship_UPLOAD,
							},
							},
						}
					}
				} else {
					// delete entity relationship with device from zone
					for key, entity := range mz.Entities {
						if entity.Id == dev.DeviceId {
							mz.Entities = append(mz.Entities[:key], mz.Entities[key+1:]...)
							break
						}
					}
					// delete entity relationship with zone from device
					if dev.GetConfiguration() != nil {
						for key, entity := range dev.Configuration.Entities {
							if entity.Id == strconv.FormatInt(int64(mz.ZoneId), 10) {
								dev.Configuration.Entities = append(dev.Configuration.Entities[:key], dev.Configuration.Entities[key+1:]...)
								break
							}
						}
					}
				}
				// upsert zone
				goreq := db.GoMiscRequest{
					Req: &db.GoRequest{
						ObjectType: mongo.ZoneObjectType,
						Obj: &db.GoObject{
							Data: mz,
							ID:   mz.DatabaseId,
						},
					},
					Ctxt: s.ctxt,
					Time: &db.TimeKeeper{},
				}
				_, err = s.miscDb.Upsert(goreq)
				if err != nil {
					return err
				}
				// upsert device configuration
				err = devicedb.UpsertDeviceConfig(dev.DeviceId, dev.Type, dev.Configuration)
				if err != nil {
					return err
				}
				notice := omnicomNotice(moc.Notice_OmnicomUploadGeoFence, activity, moc.Notice_Info)
				s.n.Notify(notice, true)
			case 1:
				notice := omnicomNotice(moc.Notice_OmnicomFullBuffer, activity, moc.Notice_Info)
				s.n.Notify(notice, true)
			case 2:
				notice := omnicomNotice(moc.Notice_OmnicomGeoIdDoesNotExist, activity, moc.Notice_Info)
				s.n.Notify(notice, true)
			case 3:
				notice := omnicomNotice(moc.Notice_OmnicomUploadGeoFence, activity, moc.Notice_Info)
				s.n.Notify(notice, true)
			default:
				return fmt.Errorf("Omnicom Geo Ack with error type %+v is not supported", report.Error_Type)
			}
		}
	}
	return nil
}

func (s *omnicomStage) updateDevice(vessel *moc.Vessel, dev *moc.Device, vesseldb db.VesselDB) {
	// update device
	for i, oldDevice := range vessel.Devices {
		if oldDevice.Type == dev.Type && oldDevice.DeviceId == dev.DeviceId {
			dev.Id = oldDevice.Id
			// remove for vessel
			dev.Configuration.Original = nil
			vessel.Devices[i] = dev
			break
		}
	}
	// update vessel
	_, err := vesseldb.Update(s.ctxt, vessel)
	if err != nil {
		log.Error(err.Error())
	} else {
		// publish vessel
		const TOPIC = "Vessel"
		s.publisher.Publish(TOPIC, envelope.Envelope{
			Type: TOPIC + "/UPDATE",
			Contents: &envelope.Envelope_Vessel{
				Vessel: vessel,
			},
		})
	}
}

func notifyTransmissionFailure(n Notifier, tr *tms.Transmission) error {
	name := ""
	if len(tr.Packets) > 0 {
		name = tr.Packets[0].Name
	}
	return n.Notify(&moc.Notice{
		NoticeId: ident.With("transmissionId", tr.Id).Hash(),
		Event:    moc.Notice_OmnicomTransmission,
		Priority: moc.Notice_Info,
		Source: &moc.SourceInfo{
			Name:        tr.Id,
			MulticastId: tr.ParentId,
		},
		Target: &moc.TargetInfo{
			Name:       name,
			Type:       tr.Destination.Type,
			DatabaseId: tr.Id,
			ActivityId: tr.MessageId,
			Imei:       tr.Destination.Id,
			RegistryId: ident.
				With("imei", tr.Destination.Id).
				Hash(),
			TrackId: ident.
				With("imei", tr.Destination.Id).
				With("site", tmsg.GClient.Local().Site).
				With("eid", tmsg.APP_ID_TFLEETD*1000).
				Hash(),
		},
	}, true)
}

//SendTestAck function will send the an ack message to the beacon after receiving
//and alert report with test mode == 1
func SendTestAck(ctxt gogroup.GoGroup, netID string) (tms.Transmission, error) {

	date := omni.CreateOmnicomDate(time.Now().UTC())

	var tran tms.Transmission

	tma := &omnicom.Omni{
		Omnicom: &omnicom.Omni_Tma{
			Tma: &omnicom.Tma{
				Header: []byte{0x30},
				Date:   &omnicom.Dt{Year: date.Year, Month: date.Month, Day: date.Day, Minute: date.Minute},
			},
		},
	}

	cmd, errpack := tmsg.PackFrom(tma)
	if errpack != nil {
		return tran, errpack
	}

	id := bson.NewObjectId()
	//destination is always set over iridium.
	dest := tms.EntityRelationship{Id: netID, Type: "iridium"}

	tran = tms.Transmission{
		Id:          id.Hex(),
		MessageId:   id.Hex(),
		Destination: &dest,
		State:       tms.Transmission_Pending,
	}

	request := &tms.Multicast{
		Id:            bson.NewObjectId().Hex(),
		Payload:       cmd,
		Destinations:  []*tms.EntityRelationship{&dest},
		Transmissions: []*tms.Transmission{&tran},
	}

	body, errpack := tmsg.PackFrom(request)
	if errpack != nil {
		return tran, errpack
	}

	infoMsg := &tms.TsiMessage{
		Source: tmsg.GClient.Local(),
		Destination: []*tms.EndPoint{
			{
				Site: tmsg.GClient.ResolveSite(""),
			},
		},
		WriteTime: tms.Now(),
		SendTime:  tms.Now(),
		Body:      body,
	}

	log.Debug("sending out test mode ack %v", infoMsg)
	tmsg.GClient.Send(ctxt, infoMsg)

	return tran, nil
}

//RequestSpecificMessage is a conviniance function to be able to request a specific message from the omnicom beacon
func RequestSpecificMessage(MsgToAsk uint32, ctxt gogroup.GoGroup, netID string, dev *moc.Device) (tms.Transmission, error) {
	date := omni.CreateOmnicomDate(time.Now().UTC())
	msgid := uint32(root.Intn(4095))
	var trans tms.Transmission
	rsm := &omnicom.Omni{
		Omnicom: &omnicom.Omni_Rsm{
			Rsm: &omnicom.Rsm{
				Header: []byte{0x33},
				//FIXME: Have to determine a way to avoir collision
				ID_Msg: msgid,
				Date:   &omnicom.Dt{Year: date.Year, Month: date.Month, Day: date.Day, Minute: date.Minute},
				// 0x00: send an alert report: message 'Alert report". Response with “Alert Report(0x02)”
				// 0x01: send the last position recorded with the message "History position report(0x01)".
				// 0x02: make a new position acquisition and send it with the message "History position report(0x01)".
				// 0x03: send the global parameters setting.Response with Global parameters(0x03)
				// 0x04: send the parameters URL. Response with API url parameters(0x08)
				// 0x10: test the 3G. The Dome send a “Single position report(0x06) and a History report(0x01) in a 3g message.
				MsgTo_Ask: MsgToAsk,
			},
		},
	}
	cmd, errpack := tmsg.PackFrom(rsm)
	if errpack != nil {
		return trans, errpack
	}
	//Default initialize destination using network
	dest := tms.EntityRelationship{Id: netID, Type: "iridium"}
	transID := bson.NewObjectId()
	trans = tms.Transmission{
		Id:          transID.Hex(),
		MessageId:   strconv.Itoa(int(msgid)),
		Destination: &dest,
		State:       tms.Transmission_Pending,
	}
	multicast := &tms.Multicast{
		Payload:       cmd,
		Destinations:  []*tms.EntityRelationship{&dest},
		Transmissions: []*tms.Transmission{&trans},
	}
	body, errpack := tmsg.PackFrom(multicast)
	if errpack != nil {
		return trans, errpack
	}
	infoMsg := &tms.TsiMessage{
		Source: tmsg.GClient.Local(),
		Destination: []*tms.EndPoint{
			{
				Site: tmsg.GClient.ResolveSite(""),
			},
		},
		WriteTime: tms.Now(),
		SendTime:  tms.Now(),
		Body:      body,
	}
	log.Debug("sending out message %v registration request %v", trans.MessageId, infoMsg)
	tmsg.GClient.Send(ctxt, infoMsg)
	return trans, nil
}

func (s *omnicomStage) start() {}

func (s *omnicomStage) analyze(update api.TrackUpdate) error {
	if len(update.Track.Targets) == 0 || s.initialized == false {
		return nil
	}
	t := update.Track
	tgt := t.Targets[0]
	if tgt.Type != devices.DeviceType_OmnicomVMS && tgt.Type != devices.DeviceType_OmnicomSolar {
		return nil
	}
	deviceDb := mongo.NewMongoDeviceDb(s.ctxt, s.clt)
	dev, err := deviceDb.FindNet(tgt.Imei.Value)
	if mgo.ErrNotFound != err {
		log.Debug("Beacon with Imei: %v is already registered", tgt.Imei.Value)
		return nil
	}
	tran, err := RequestSpecificMessage(0x03, s.ctxt, tgt.Imei.Value, dev)
	if err != nil {
		return err
	}
	//Upsert Transmission created by SendTestAck
	tranDb := mongo.NewTransmissionDb(s.ctxt, s.clt)
	tran.Packets = append(tran.Packets, &tms.Packet{Name: omnicom.OmnicomConfiguration_Action_name[int32(omnicom.OmnicomConfiguration_RequestGlobalParameters)], State: 1, Status: &tms.ResponseStatus{Code: 102}})
	tran.Status = &tms.ResponseStatus{Code: 102}
	if dev.Id != "" {
		tran.Destination = &tms.EntityRelationship{Id: dev.Id, Type: dev.GetType()}
	}
	err = tranDb.Update(&tran)
	if err != nil {
		return err
	}
	log.Info("Requesting global parameters for beacon with Imei %v", tgt.Imei.Value)
	return nil
}

func omnicomNotice(event moc.Notice_Event, update interface{}, pt moc.Notice_Priority) *moc.Notice {

	if reflect.TypeOf(update).String() == reflect.TypeOf((*tms.Track)(nil)).String() {

		track := update.(*tms.Track)
		imei := ""
		nodeID := ""

		tgt := track.Targets[0]
		id := ident.With("event", event)
		switch {
		case tgt.Imei != nil:
			imei = tgt.Imei.Value
			id = id.With("imei", imei)
		case tgt.Nodeid != nil:
			nodeID = tgt.Nodeid.Value
			id = id.With("nodeId", nodeID)
		default:
			log.Warn("cannot find source of the activity message: %+v", tgt)
			return nil
		}

		return &moc.Notice{
			NoticeId: id.Hash(),
			Event:    event,
			Priority: pt,
			Target:   TargetInfoFromTrack(track),
		}
	}

	if reflect.TypeOf(update).String() == reflect.TypeOf((*tms.MessageActivity)(nil)).String() {

		activity := update.(*tms.MessageActivity)
		imei := ""
		nodeID := ""

		id := ident.With("event", event)

		switch {
		case activity.Imei != nil:
			imei = activity.Imei.Value
			id = id.With("imei", imei)
		case activity.NodeId != nil:
			nodeID = activity.NodeId.Value
			id = id.With("nodeId", nodeID)
		default:
			log.Warn("Cannot find network id of the activity message %v", activity)
			return nil
		}

		return &moc.Notice{
			NoticeId: id.Hash(),
			Event:    event,
			Priority: pt,
			Target:   TargetInfoFromActivity(activity),
		}
	}
	return nil
}
