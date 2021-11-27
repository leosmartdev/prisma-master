package omngen

import "prisma/tms/log"
import "prisma/tms/iridium"
import "prisma/tms/omnicom"
import "time"
import "encoding/json"
import . "github.com/grafov/bcast"

//EventTime current event time for alert reports
func EventTime() omnicom.Date_Event {

	var date omnicom.Date_Event

	date.Year = uint32(time.Now().Year() - 2000)
	date.Month = uint32(time.Now().Month())
	date.Day = uint32(time.Now().Day())
	date.Minute = uint32((uint32(time.Now().Hour()) * 60) + uint32(time.Now().Minute()))

	return date
}

/*func ReceiveAlert(bool chan timeout, group1 *Group, member2 *Member, spr UpAlt, beacon Jbeacon) {

}*/

//SimAltBeacon get the last position generated by the beacon and generates and alert when it receives an alert request from the Alert CLI
func SimAltBeacon(beacon Jbeacon, group1 *Group, group2 *Group, Address string) {
	// create the message header for the beacon
	// 1- initialize MOH structure using iridium library
	// 2- use EncodeMO() function to encode the MOH structure into []byte
	MOH := iridium.MOHeader{0x01, 28, 2578512475, beacon.Imei, "0", 15661, 380, 1475582020}

	Hraw, err := MOH.EncodeMO()
	if err != nil {
		log.Error(" %+v\n", err)
	}

	AR := omnicom.AR{0x02,
		0,
		omnicom.Date_Position{},
		omnicom.Date_Event{},
		omnicom.Power_Up{0, 0},
		omnicom.Power_Down{0, 0},
		omnicom.Battery_Alert{0, 0},
		omnicom.Intrusion_Alert{0, 0},
		omnicom.No_Position_Fix{0, 0, 5},
		omnicom.JB_Dome_Alert{0, 0},
		omnicom.Loss_Mobile_Com{0, 0},
		omnicom.Daylight_Alert{0, 0},
		omnicom.Assistance_Alert{0, 0},
		omnicom.Test_Mode{0, 0},
		0, omnicom.Move{0, 0}, //extention bit and move
		0, 0, // extention bit and beacon id
		0, 18}

loop:
	for {

		timeout := make(chan bool, 1)

		go func() {
			time.Sleep(time.Duration(beacon.Reportperiod) * time.Minute)
			timeout <- true
		}()

		member2 := group2.Join()

	ListenForSpr:
		for {
			var spr UpAlt = UpAlt{}

			select {
			case val2 := <-member2.In:
				spr = val2.(UpAlt)
			case <-timeout:
				member2.Close()
				continue loop
			}

			if spr.Imei != beacon.Imei {
				continue ListenForSpr
			}
		ListenForAlert:
			for {

				var bytes []byte
				member1 := group1.Join()

				select {
				case val := <-member1.In:
					bytes = val.([]byte)
				case <-timeout:
					member2.Close()
					member1.Close()
					continue loop
				}

				if len(bytes) != 0 {

					var Alt struct {
						Imei   string
						PU     bool
						PD     bool
						BA     bool
						IA     bool
						NPF    bool
						JBDA   bool
						LMC    bool
						DA     bool
						AA     bool
						TM     bool
						Status string
					}

					err := json.Unmarshal(bytes, &Alt)
					if err != nil {
						log.Error("Unable to parse MT JSON message from Alert: %v\n", string(bytes))
					}

					if beacon.Imei != Alt.Imei {
						member1.Close()
						continue ListenForAlert
					} else if beacon.Imei == Alt.Imei {

						if Alt.Status == "start" {
							if Alt.PU {

								AR.Power_Up = omnicom.Power_Up{1, 1}
								AR.Date_Position = spr.Spr.Date_Position
								AR.Date_Event = EventTime()

								Payload := iridium.MPayload{0x02, 0, &AR}

								Praw, err := Payload.Encode()
								if err != nil {
									log.Error(" %+v\n", err)
								}
								raw := []byte{}
								raw = append(raw, 0x01, byte(uint16(len(Praw)+len(Hraw)))>>8, byte(uint16(len(Praw)+len(Hraw))))
								raw = append(raw, Hraw...)
								raw = append(raw, Praw...)

								log.Debug("Generated Iridium AR %+v and sent to %+v", raw, Address)

								err = Send(raw, Address)
								if err != nil {
									log.Error(" %+v\n", err)
								}

								AR.Power_Up = omnicom.Power_Up{0, 1}
							}

							if Alt.PD {

								AR.Power_Down = omnicom.Power_Down{1, 1}
								AR.Date_Position = spr.Spr.Date_Position
								AR.Date_Event = EventTime()

								Payload := iridium.MPayload{0x02, 0, &AR}

								Praw, err := Payload.Encode()
								if err != nil {
									log.Error(" %+v\n", err)
								}
								raw := []byte{}
								raw = append(raw, 0x01, byte(uint16(len(Praw)+len(Hraw)))>>8, byte(uint16(len(Praw)+len(Hraw))))
								raw = append(raw, Hraw...)
								raw = append(raw, Praw...)

								log.Debug("Generated Iridium AR %+v and sent to %+v", raw, Address)

								err = Send(raw, Address)
								if err != nil {
									log.Error(" %+v\n", err)
								}

								AR.Power_Down = omnicom.Power_Down{0, 1}
							}

							if Alt.IA {

								AR.Intrusion_Alert = omnicom.Intrusion_Alert{1, 1}
								AR.Date_Position = spr.Spr.Date_Position
								AR.Date_Event = EventTime()

								Payload := iridium.MPayload{0x02, 0, &AR}

								Praw, err := Payload.Encode()
								if err != nil {
									log.Error(" %+v\n", err)
								}
								raw := []byte{}
								raw = append(raw, 0x01, byte(uint16(len(Praw)+len(Hraw)))>>8, byte(uint16(len(Praw)+len(Hraw))))
								raw = append(raw, Hraw...)
								raw = append(raw, Praw...)

								log.Debug("Generated Iridium AR %+v and sent to %+v", raw, Address)

								err = Send(raw, Address)
								if err != nil {
									log.Error(" %+v\n", err)
								}
								AR.Intrusion_Alert = omnicom.Intrusion_Alert{0, 1}
							}

							if Alt.NPF { // end of IA condition

								AR.No_Position_Fix = omnicom.No_Position_Fix{1, 1, 5}
								AR.Date_Position = spr.Spr.Date_Position
								AR.Date_Event = EventTime()

								Payload := iridium.MPayload{0x02, 0, &AR}

								Praw, err := Payload.Encode()
								if err != nil {
									log.Error(" %+v\n", err)
								}
								raw := []byte{}
								raw = append(raw, 0x01, byte(uint16(len(Praw)+len(Hraw)))>>8, byte(uint16(len(Praw)+len(Hraw))))
								raw = append(raw, Hraw...)
								raw = append(raw, Praw...)

								log.Debug("Generated Iridium AR %+v and sent to %+v", raw, Address)

								err = Send(raw, Address)
								if err != nil {
									log.Error(" %+v\n", err)
								}
								AR.No_Position_Fix = omnicom.No_Position_Fix{0, 1, 5}
							}
							if Alt.JBDA { //end of NPF condition

								AR.JB_Dome_Alert = omnicom.JB_Dome_Alert{1, 1}
								AR.Date_Position = spr.Spr.Date_Position
								AR.Date_Event = EventTime()

								Payload := iridium.MPayload{0x02, 0, &AR}

								Praw, err := Payload.Encode()
								if err != nil {
									log.Error(" %+v\n", err)
								}
								raw := []byte{}
								raw = append(raw, 0x01, byte(uint16(len(Praw)+len(Hraw)))>>8, byte(uint16(len(Praw)+len(Hraw))))
								raw = append(raw, Hraw...)
								raw = append(raw, Praw...)

								log.Debug("Generated Iridium AR %+v and sent to %+v", raw, Address)

								err = Send(raw, Address)
								if err != nil {
									log.Error(" %+v\n", err)
								}
								AR.JB_Dome_Alert = omnicom.JB_Dome_Alert{0, 1}

							}

							if Alt.LMC { // end of JBDA condition

								AR.Loss_Mobile_Com = omnicom.Loss_Mobile_Com{1, 1}
								AR.Date_Position = spr.Spr.Date_Position
								AR.Date_Event = EventTime()

								Payload := iridium.MPayload{0x02, 0, &AR}

								Praw, err := Payload.Encode()
								if err != nil {
									log.Error(" %+v\n", err)
								}
								raw := []byte{}
								raw = append(raw, 0x01, byte(uint16(len(Praw)+len(Hraw)))>>8, byte(uint16(len(Praw)+len(Hraw))))
								raw = append(raw, Hraw...)
								raw = append(raw, Praw...)

								log.Debug("Generated Iridium AR %+v and sent to %+v", raw, Address)

								err = Send(raw, Address)
								if err != nil {
									log.Error(" %+v\n", err)
								}
								AR.Loss_Mobile_Com = omnicom.Loss_Mobile_Com{0, 1}
							}

							if Alt.DA { // end of LMC condition

								AR.Daylight_Alert = omnicom.Daylight_Alert{1, 1}
								AR.Date_Position = spr.Spr.Date_Position
								AR.Date_Event = EventTime()

								Payload := iridium.MPayload{0x02, 0, &AR}

								Praw, err := Payload.Encode()
								if err != nil {
									log.Error(" %+v\n", err)
								}
								raw := []byte{}
								raw = append(raw, 0x01, byte(uint16(len(Praw)+len(Hraw)))>>8, byte(uint16(len(Praw)+len(Hraw))))
								raw = append(raw, Hraw...)
								raw = append(raw, Praw...)

								log.Debug("Generated Iridium AR %+v and sent to %+v", raw, Address)

								err = Send(raw, Address)
								if err != nil {
									log.Error(" %+v\n", err)
								}
								AR.Daylight_Alert = omnicom.Daylight_Alert{0, 1}
							}

							if Alt.AA { // end of DA condition

								AR.Assistance_Alert = omnicom.Assistance_Alert{1, 1}
								AR.Date_Position = spr.Spr.Date_Position
								AR.Date_Event = EventTime()

								Payload := iridium.MPayload{0x02, 0, &AR}

								Praw, err := Payload.Encode()
								if err != nil {
									log.Error(" %+v\n", err)
								}
								raw := []byte{}
								raw = append(raw, 0x01, byte(uint16(len(Praw)+len(Hraw)))>>8, byte(uint16(len(Praw)+len(Hraw))))
								raw = append(raw, Hraw...)
								raw = append(raw, Praw...)

								log.Debug("Generated Iridium AR %+v and sent to %+v", raw, Address)

								err = Send(raw, Address)
								if err != nil {
									log.Error(" %+v\n", err)
								}
								AR.Assistance_Alert = omnicom.Assistance_Alert{0, 1}
							}

							if Alt.TM { // end of AA condition

								AR.Test_Mode = omnicom.Test_Mode{1, 1}
								AR.Date_Position = spr.Spr.Date_Position
								AR.Date_Event = EventTime()

								Payload := iridium.MPayload{0x02, 0, &AR}

								Praw, err := Payload.Encode()
								if err != nil {
									log.Error(" %+v\n", err)
								}
								raw := []byte{}
								raw = append(raw, 0x01, byte(uint16(len(Praw)+len(Hraw)))>>8, byte(uint16(len(Praw)+len(Hraw))))
								raw = append(raw, Hraw...)
								raw = append(raw, Praw...)

								log.Debug("Generated Iridium AR %+v and sent to %+v", raw, Address)

								err = Send(raw, Address)
								if err != nil {
									log.Error(" %+v\n", err)
								}
								AR.Test_Mode = omnicom.Test_Mode{0, 1}
							}
							if !Alt.TM && !Alt.PU && !Alt.PD && !Alt.JBDA && !Alt.AA && !Alt.DA && !Alt.LMC && !Alt.NPF && !Alt.IA {
								log.Error("There is not alert to start")
							}

						} else if Alt.Status == "stop" {

							if Alt.PU {

								AR.Power_Up = omnicom.Power_Up{1, 0}
								AR.Date_Position = spr.Spr.Date_Position
								AR.Date_Event = EventTime()

								Payload := iridium.MPayload{0x02, 0, &AR}

								Praw, err := Payload.Encode()
								if err != nil {
									log.Error(" %+v\n", err)
								}
								raw := []byte{}
								raw = append(raw, 0x01, byte(uint16(len(Praw)+len(Hraw)))>>8, byte(uint16(len(Praw)+len(Hraw))))
								raw = append(raw, Hraw...)
								raw = append(raw, Praw...)

								log.Debug("Generated Iridium AR %+v and sent to %+v", raw, Address)

								err = Send(raw, Address)
								if err != nil {
									log.Error(" %+v\n", err)
								}

								AR.Power_Up = omnicom.Power_Up{0, 0}
							}

							if Alt.PD {

								AR.Power_Down = omnicom.Power_Down{1, 0}
								AR.Date_Position = spr.Spr.Date_Position
								AR.Date_Event = EventTime()

								Payload := iridium.MPayload{0x02, 0, &AR}

								Praw, err := Payload.Encode()
								if err != nil {
									log.Error(" %+v\n", err)
								}
								raw := []byte{}
								raw = append(raw, 0x01, byte(uint16(len(Praw)+len(Hraw)))>>8, byte(uint16(len(Praw)+len(Hraw))))
								raw = append(raw, Hraw...)
								raw = append(raw, Praw...)

								log.Debug("Generated Iridium AR %+v and sent to %+v", raw, Address)

								err = Send(raw, Address)
								if err != nil {
									log.Error(" %+v\n", err)
								}

								AR.Power_Down = omnicom.Power_Down{0, 0}
							}

							if Alt.IA {

								AR.Intrusion_Alert = omnicom.Intrusion_Alert{1, 0}
								AR.Date_Position = spr.Spr.Date_Position
								AR.Date_Event = EventTime()

								Payload := iridium.MPayload{0x02, 0, &AR}

								Praw, err := Payload.Encode()
								if err != nil {
									log.Error(" %+v\n", err)
								}
								raw := []byte{}
								raw = append(raw, 0x01, byte(uint16(len(Praw)+len(Hraw)))>>8, byte(uint16(len(Praw)+len(Hraw))))
								raw = append(raw, Hraw...)
								raw = append(raw, Praw...)

								log.Debug("Generated Iridium AR %+v and sent to %+v", raw, Address)

								err = Send(raw, Address)
								if err != nil {
									log.Error(" %+v\n", err)
								}
								AR.Intrusion_Alert = omnicom.Intrusion_Alert{0, 0}
							}

							if Alt.NPF { // end of IA condition

								AR.No_Position_Fix = omnicom.No_Position_Fix{1, 0, 7}
								AR.Date_Position = spr.Spr.Date_Position
								AR.Date_Event = EventTime()

								Payload := iridium.MPayload{0x02, 0, &AR}

								Praw, err := Payload.Encode()
								if err != nil {
									log.Error(" %+v\n", err)
								}
								raw := []byte{}
								raw = append(raw, 0x01, byte(uint16(len(Praw)+len(Hraw)))>>8, byte(uint16(len(Praw)+len(Hraw))))
								raw = append(raw, Hraw...)
								raw = append(raw, Praw...)

								log.Debug("Generated Iridium AR %+v and sent to %+v", raw, Address)

								err = Send(raw, Address)
								if err != nil {
									log.Error(" %+v\n", err)
								}
								AR.No_Position_Fix = omnicom.No_Position_Fix{0, 0, 7}
							}

							if Alt.JBDA { //end of NPF condition

								AR.JB_Dome_Alert = omnicom.JB_Dome_Alert{1, 0}
								AR.Date_Position = spr.Spr.Date_Position
								AR.Date_Event = EventTime()

								Payload := iridium.MPayload{0x02, 0, &AR}

								Praw, err := Payload.Encode()
								if err != nil {
									log.Error(" %+v\n", err)
								}
								raw := []byte{}
								raw = append(raw, 0x01, byte(uint16(len(Praw)+len(Hraw)))>>8, byte(uint16(len(Praw)+len(Hraw))))
								raw = append(raw, Hraw...)
								raw = append(raw, Praw...)

								log.Debug("Generated Iridium AR %+v and sent to %+v", raw, Address)

								err = Send(raw, Address)
								if err != nil {
									log.Error(" %+v\n", err)
								}
								AR.JB_Dome_Alert = omnicom.JB_Dome_Alert{0, 0}

							}

							if Alt.LMC { // end of JBDA condition

								AR.Loss_Mobile_Com = omnicom.Loss_Mobile_Com{1, 0}
								AR.Date_Position = spr.Spr.Date_Position
								AR.Date_Event = EventTime()

								Payload := iridium.MPayload{0x02, 0, &AR}

								Praw, err := Payload.Encode()
								if err != nil {
									log.Error(" %+v\n", err)
								}
								raw := []byte{}
								raw = append(raw, 0x01, byte(uint16(len(Praw)+len(Hraw)))>>8, byte(uint16(len(Praw)+len(Hraw))))
								raw = append(raw, Hraw...)
								raw = append(raw, Praw...)

								log.Debug("Generated Iridium AR %+v and sent to %+v", raw, Address)

								err = Send(raw, Address)
								if err != nil {
									log.Error(" %+v\n", err)
								}
								AR.Loss_Mobile_Com = omnicom.Loss_Mobile_Com{0, 0}
							}

							if Alt.DA { // end of LMC condition

								AR.Daylight_Alert = omnicom.Daylight_Alert{1, 0}
								AR.Date_Position = spr.Spr.Date_Position
								AR.Date_Event = EventTime()

								Payload := iridium.MPayload{0x02, 0, &AR}

								Praw, err := Payload.Encode()
								if err != nil {
									log.Error(" %+v\n", err)
								}
								raw := []byte{}
								raw = append(raw, 0x01, byte(uint16(len(Praw)+len(Hraw)))>>8, byte(uint16(len(Praw)+len(Hraw))))
								raw = append(raw, Hraw...)
								raw = append(raw, Praw...)

								log.Debug("Generated Iridium AR %+v and sent to %+v", raw, Address)

								err = Send(raw, Address)
								if err != nil {
									log.Error(" %+v\n", err)
								}
								AR.Daylight_Alert = omnicom.Daylight_Alert{0, 0}
							}

							if Alt.AA { // end of DA condition

								AR.Assistance_Alert = omnicom.Assistance_Alert{1, 0}
								AR.Date_Position = spr.Spr.Date_Position
								AR.Date_Event = EventTime()

								Payload := iridium.MPayload{0x02, 0, &AR}

								Praw, err := Payload.Encode()
								if err != nil {
									log.Error(" %+v\n", err)
								}
								raw := []byte{}
								raw = append(raw, 0x01, byte(uint16(len(Praw)+len(Hraw)))>>8, byte(uint16(len(Praw)+len(Hraw))))
								raw = append(raw, Hraw...)
								raw = append(raw, Praw...)

								log.Debug("Generated Iridium AR %+v and sent to %+v", raw, Address)

								err = Send(raw, Address)
								if err != nil {
									log.Error(" %+v\n", err)
								}
								AR.Assistance_Alert = omnicom.Assistance_Alert{0, 0}
							}

							if Alt.TM { // end of AA condition

								AR.Test_Mode = omnicom.Test_Mode{1, 0}
								AR.Date_Position = spr.Spr.Date_Position
								AR.Date_Event = EventTime()

								Payload := iridium.MPayload{0x02, 0, &AR}

								Praw, err := Payload.Encode()
								if err != nil {
									log.Error(" %+v\n", err)
								}
								raw := []byte{}
								raw = append(raw, 0x01, byte(uint16(len(Praw)+len(Hraw)))>>8, byte(uint16(len(Praw)+len(Hraw))))
								raw = append(raw, Hraw...)
								raw = append(raw, Praw...)

								log.Debug("Generated Iridium AR %+v and sent to %+v", raw, Address)

								err = Send(raw, Address)
								if err != nil {
									log.Error(" %+v\n", err)
								}
								AR.Test_Mode = omnicom.Test_Mode{0, 0}
							}
							if !Alt.TM && !Alt.PU && !Alt.PD && !Alt.JBDA && !Alt.AA && !Alt.DA && !Alt.LMC && !Alt.NPF && !Alt.IA {
								log.Error("There is not alert to stop")
							}

						} // end of else if Status == start Status == stop statements
					}
				} // enf of if byte != 0 condition
				member1.Close()
			} // end of ListenForAlert: for loop of inner loop
		} // end of ListenForSpr: spr.Imei != beacon.Imei
	} // end of loop
}