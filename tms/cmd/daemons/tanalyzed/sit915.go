package main

import (
	"bytes"
	"fmt"
	"os"
	"prisma/gogroup"
	"prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/db"
	"prisma/tms/db/mongo"
	"prisma/tms/envelope"
	"prisma/tms/log"
	"prisma/tms/moc"
	client "prisma/tms/tmsg/client"
	"prisma/tms/ws"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"github.com/secsy/goftp"
	"golang.org/x/crypto/ssh"
)

const (
	Sit915ObjectType = "prisma.tms.moc.Sit915"
)

type sit915Stage struct {
	tsiClient client.TsiClient
	n         Notifier
	publisher *ws.Publisher
}

func newSit915Stage(tsiClient client.TsiClient, notifier Notifier, publisher *ws.Publisher) *sit915Stage {
	return &sit915Stage{
		tsiClient: tsiClient,
		n:         notifier,
		publisher: publisher,
	}
}

func (stage *sit915Stage) init(ctx gogroup.GoGroup, dbClient *mongo.MongoClient) error {
	log.Info("Sit915 init")
	miscDB := mongo.NewMongoMiscData(ctx, dbClient)
	sit915DB := mongo.NewSit915Db(miscDB)
	remoteSiteDB := mongo.NewMongoRemoteSiteMiscData(miscDB)
	configDB := mongo.ConfigDb{}

	stream := miscDB.GetPersistentStream(db.GoMiscRequest{
		Req: &db.GoRequest{
			ObjectType: Sit915ObjectType,
		},
		Ctxt: ctx,
		Time: &db.TimeKeeper{},
	}, nil, nil)

	ctx.Go(func() {
		for {
			select {
			case update, ok := <-stream:
				if !ok {
					continue // channel was closed
				}
				if update.Contents == nil || update.Contents.Data == nil {
					log.Error("no content: %v", update)
					continue
				}
				sit915, ok := update.Contents.Data.(*moc.Sit915)
				if !ok {
					log.Error("bad content: %v", update)
					continue
				}

				if sit915.Status == moc.Sit915_SENT.String() ||
					sit915.Status == moc.Sit915_FAILED.String() {
					continue
				}

				go stage.sendMessage(ctx, sit915DB, remoteSiteDB, configDB, sit915)

			case <-ctx.Done():
				return
			}
		}
	})
	log.Info("Sit915 init done")
	return nil
}

// send message
func (stage *sit915Stage) sendMessage(ctx gogroup.GoGroup, sit915DB db.Sit915DB, remoteSiteDB db.RemoteSiteDB, configDB mongo.ConfigDb, sit915 *moc.Sit915) {
	remoteSiteConfig, err := remoteSiteDB.FindOneRemoteSite(sit915.RemotesiteId, false)
	if err != nil {
		errMessage := fmt.Sprintf("Unable to read remote site configuration: %v", err)
		log.Error(errMessage)

		sit915.Timestamp = tms.Now()
		sit915.Status = moc.Sit915_FAILED.String()
		sit915.ErrorDetail = errMessage
		err = sit915DB.UpsertSit915(sit915)
		if err != nil {
			log.Error("Unable to set message status")
		}

		stage.sendStatus(sit915)

		return
	}

	logMsgHeader := fmt.Sprintf("[SIT 915 to %s (%s)]", remoteSiteConfig.Csname, remoteSiteConfig.Cscode)

	// Message Field # 1 - Transmission number and retransmission number
	transmissionNum := sit915.TransmissionNum
	retransmissionNum := sit915.RetransmissionNum
	mf1 := fmt.Sprintf("/%.5d %.5d", transmissionNum, retransmissionNum)

	// Message Field # 2 - Code of the reporting facility
	localSiteConfig, err := configDB.Read(ctx)
	if err != nil {
		errMessage := fmt.Sprintf("Unable to read local site configuration: %v", err)
		log.Error("%s %s", logMsgHeader, errMessage)

		sit915.Timestamp = tms.Now()
		sit915.Status = moc.Sit915_FAILED.String()
		sit915.ErrorDetail = errMessage
		err = sit915DB.UpsertSit915(sit915)
		if err != nil {
			log.Error("%s Unable to set message status", logMsgHeader)
		}

		stage.sendStatus(sit915)

		return
	}

	localCode := localSiteConfig.Site.Cscode
	mf2 := fmt.Sprintf("/%s", localCode)

	// Message Field # 3 - The date and time of the transmission
	t := time.Now()
	year := strconv.Itoa(t.Year())[2:4]
	dayOfYear := t.YearDay()
	hour := t.Hour()
	min := t.Minute()
	mf3 := fmt.Sprintf("/%s %.3d %.2d%.2d", year, dayOfYear, hour, min)

	// Message Field # 4 - SIT number
	mf4 := fmt.Sprintf("/%d", 915)

	// Message Field # 5 - Code of the final destination
	remoteSiteCode := remoteSiteConfig.Cscode
	mf5 := fmt.Sprintf("/%s", remoteSiteCode)

	// Message Field # 41 - Narrative text
	mf41 := fmt.Sprintf("/%s\nQQQQ", sit915.Narrative)

	// Message Field # 42 - Always /LASSIT
	mf42 := "/LASSIT"

	// Message Field # 43 - Always /ENDMSG
	mf43 := "/ENDMSG"

	message := fmt.Sprintf("%s%s%s\n%s%s\n%s\n%s\n%s\n", mf1, mf2, mf3, mf4, mf5, mf41, mf42, mf43)

	if sit915.CommLinkType == moc.Sit915_FTP.String() {
		ipAddress := remoteSiteConfig.FtpCommunication.IpAddress
		username := remoteSiteConfig.FtpCommunication.Username
		password := remoteSiteConfig.FtpCommunication.Password
		startingDirectory := remoteSiteConfig.FtpCommunication.StartingDirectory
		if startingDirectory == "" {
			startingDirectory = "~"
		}

		filename := fmt.Sprintf("%s_%s_%.5d", localSiteConfig.Site.Csname, remoteSiteConfig.Csname, sit915.TransmissionNum)
		fallback_to_ftp := remoteSiteConfig.FtpCommunication.FallbackToFtp

		sit915.CommLinkType = moc.Sit915_SFTP.String()
		err := stage.sendMessageOverSftp(ipAddress, username, password, startingDirectory, filename, message, logMsgHeader)
		if err != nil && fallback_to_ftp == true {
			log.Info("%s Trying to send SIT 915 message over FTP", logMsgHeader)

			sit915.CommLinkType = moc.Sit915_FTP.String()
			err = stage.sendMessageOverFtp(ipAddress, username, password, startingDirectory, filename, message, logMsgHeader, false)
		}

		if err != nil {
			log.Error("%s Failed to send SIT 915 message", logMsgHeader)

			errMessageSplit := strings.Split(err.Error(), "]")
			errMessage := errMessageSplit[len(errMessageSplit)-1]

			// Set status as "FAILED"
			sit915.Timestamp = tms.Now()
			sit915.Status = moc.Sit915_FAILED.String()
			sit915.ErrorDetail = errMessage
			err = sit915DB.UpsertSit915(sit915)
			if err != nil {
				log.Error("%s Unable to set message status", logMsgHeader)
			}

			stage.sendStatus(sit915)

			return
		}

		// Set status as "SENT"
		sit915.MessageBody = message
		sit915.Timestamp = tms.ToTimestamp(t)
		sit915.Status = moc.Sit915_SENT.String()
		sit915.ErrorDetail = ""
		err = sit915DB.UpsertSit915(sit915)
		if err != nil {
			log.Error("%s Unable to set message status", logMsgHeader)
			return
		}

		stage.sendStatus(sit915)

		// Increase transmission number for the remote site
		remoteSiteConfig.CurrentMessageNum += 1
		if remoteSiteConfig.CurrentMessageNum == 100000 {
			remoteSiteConfig.CurrentMessageNum = 1
		}

		err = remoteSiteDB.UpsertRemoteSite(remoteSiteConfig)
		if err != nil {
			log.Error("%s Unable to increase the transmission number for the remote site", logMsgHeader)
			return
		}
	}
}

// Send SIT915 message over SFTP
func (stage *sit915Stage) sendMessageOverSftp(ipAddress string, username string, password string, startingDirectory string, filename string, message string, logMsgHeader string) error {
	sshConfig := &ssh.ClientConfig{
		User:            username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}

	sshAddress := fmt.Sprintf("%s:22", ipAddress)

	sshConn, err := ssh.Dial("tcp", sshAddress, sshConfig)
	if err != nil {
		log.Error("%s Unable to connect to SFTP server: %v", logMsgHeader, err)

		return err
	}

	log.Info("%s Successfully connected to SFTP server", logMsgHeader)

	// Open an SFTP session over a ssh connection.
	client, err := sftp.NewClient(sshConn)
	if err != nil {
		log.Error("%s Unable to open an SFTP session: %v", logMsgHeader, err)

		return err
	}
	defer func() {
		err = client.Close()
		if err != nil {
			log.Error("%s Unable to disconnect from SFTP server: %v", logMsgHeader, err)

			return
		}

		log.Info("%s Successfully disonnected from SFTP server", logMsgHeader)
	}()

	tmpFilePath := fmt.Sprintf("%s/%s.tmp", startingDirectory, filename)
	dstFile, err := client.OpenFile(tmpFilePath, (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		log.Error("%s Unable to open/create file %s on SFTP server: %v\n", logMsgHeader, tmpFilePath, err)

		// Delete the temp file
		client.Remove(tmpFilePath)

		return err
	}
	defer dstFile.Close()

	log.Info("%s Successfully opened/created file %s on SFTP server", logMsgHeader, tmpFilePath)

	if _, err := dstFile.Write([]byte(message)); err != nil {
		log.Error("%s Unable to write file %s on SFTP server: %v\n", logMsgHeader, tmpFilePath, err)

		return err
	}

	log.Info("%s Successfully wrote file %s on SFTP server", logMsgHeader, tmpFilePath)

	filePath := fmt.Sprintf("%s/%s.txt", startingDirectory, filename)

	// Check if there is a duplicated filename
	if _, err = client.Stat(filePath); err == nil {
		// Delete the file
		err = client.Remove(filePath)
		if err != nil {
			log.Error("%s Unable to delete duplicated file on SFTP server: %v\n", logMsgHeader, err)

			// Delete the temp file
			client.Remove(tmpFilePath)

			return err
		}
	}

	// Replace .tmp extension with .txt
	err = client.Rename(tmpFilePath, filePath)
	if err != nil {
		log.Error("%s Unable to rename file %s to %s on SFTP server: %v\n", logMsgHeader, tmpFilePath, filePath, err)

		// Delete the temp file
		client.Remove(tmpFilePath)

		return err
	}

	log.Info("%s Successfully renamed file %s to %s on SFTP server", logMsgHeader, tmpFilePath, filePath)

	log.Info("%s Successfully sent SIT 915 message over SFTP", logMsgHeader)

	return nil
}

// Send SIT915 message over FTP
func (stage *sit915Stage) sendMessageOverFtp(ipAddress string, username string, password string, startingDirectory string, filename string, message string, logMsgHeader string, activeMode bool) error {
	address := fmt.Sprintf("%s:21", ipAddress)

	config := goftp.Config{}
	config.User = username
	config.Password = password

	config.ActiveTransfers = activeMode

	client, err := goftp.DialConfig(config, address)
	if err != nil {
		log.Error("%s Unable to connect to FTP server: %v", logMsgHeader, err)

		return err
	}

	var ftpMode string
	if activeMode == true {
		ftpMode = "active"
	} else {
		ftpMode = "passive"
	}

	log.Info("%s Successfully connected to FTP server", logMsgHeader)
	log.Info("%s Trying with %s mode", logMsgHeader, ftpMode)

	data := bytes.NewBufferString(message)
	tmpFilePath := fmt.Sprintf("%s/%s.tmp", startingDirectory, filename)
	err = client.Store(tmpFilePath, data)
	if err != nil {
		if activeMode == false {
			log.Error("%s Unable to write file %s on FTP server with passive mode: %v", logMsgHeader, tmpFilePath, err)

			// Delete the temp file
			client.Delete(tmpFilePath)

			// Try with active mode
			err = stage.sendMessageOverFtp(ipAddress, username, password, startingDirectory, filename, message, logMsgHeader, true)

			return err
		} else {
			log.Error("%s Unable to write file %s on FTP server with active mode: %v", logMsgHeader, tmpFilePath, err)

			// Delete the temp file
			client.Delete(tmpFilePath)

			return err
		}
	}

	log.Info("%s Successfully wrote file %s on FTP server", logMsgHeader, tmpFilePath)

	filePath := fmt.Sprintf("%s/%s.txt", startingDirectory, filename)

	// Check if there is a duplicated filename
	if _, err = client.Stat(filePath); err == nil {
		// Delete the file
		err = client.Delete(filePath)
		if err != nil {
			log.Error("%s Unable to delete duplicated file on FTP server: %v\n", logMsgHeader, err)

			// Delete the temp file
			client.Delete(tmpFilePath)

			return err
		}
	}

	// Replace .tmp extension with .txt
	err = client.Rename(tmpFilePath, filePath)
	if err != nil {
		log.Error("%s Unable to rename file %s to %s on FTP server: %v", logMsgHeader, tmpFilePath, filePath, err)

		// Delete the temp file
		client.Delete(tmpFilePath)

		return err
	}

	log.Info("%s Successfully renamed file %s to %s on FTP server", logMsgHeader, tmpFilePath, filePath)

	if err := client.Close(); err != nil {
		log.Error("%s Unable to disconnect from FTP server: %v", logMsgHeader, err)

		return err
	}

	log.Info("%s Successfully disconnected from FTP server", logMsgHeader)

	log.Info("%s Successfully sent SIT 915 message over FTP", logMsgHeader)

	return nil
}

// Send status over WebSocket
func (stage *sit915Stage) sendStatus(sit915 *moc.Sit915) {
	stage.publisher.Publish("Sit915", envelope.Envelope{
		Type: "Sit915/STATUS",
		Contents: &envelope.Envelope_Sit915{
			Sit915: sit915,
		},
	})
}

// start not used
func (stage *sit915Stage) start() {}

// analyze not used
func (stage *sit915Stage) analyze(update api.TrackUpdate) error {
	return nil
}
