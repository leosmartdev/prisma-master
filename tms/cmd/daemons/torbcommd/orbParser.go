// orbParser is a daemon to request AIS data from Machine to Machine Data Streaming pushed by
// ORBCOMM server.
package main

import (
	"errors"
	"strconv"
	"strings"
)

type Pmwlss struct {
	CurrentTime uint64
	Status      int
	Username    string
	Password    string
	Direction   int
}

type Pmwts2 struct {
	Source           int
	LcSeqNumber      int
	LcStartTime      uint64
	LcEndTime        uint64
	ClosureReason    int
	LcNumberMessages int
	Direction        int
}

type Pmwts1 struct {
	CurrentTime      uint64
	Source           int
	Version          string
	CcSeqNumber      int
	CcNumberMessages int
	CcStartTime      uint64
	Direction        int
}

type Aivdm struct {
	TotalSentences          int
	SentenceNumber          int
	SequentialMessageNumber int
	AisChannel              string
	EncapsulatedData        string
	FillBits                int
	EntireAivdmMessage      string
}

func DecodeLogin(message string, pmwlss *Pmwlss) error {
	var err error

	words := strings.Split(message, ",")
	if len(words) != 6 {
		return errors.New("Invalid PMWLSS message: " + message)
	}

	pmwlss.CurrentTime, err = strconv.ParseUint(words[1], 10, 64)
	if err != nil {
		return err
	}

	pmwlss.Status, err = strconv.Atoi(words[2])
	if err != nil {
		return err
	}

	pmwlss.Username = words[3]

	pmwlss.Direction, err = strconv.Atoi(words[5])
	if err != nil {
		return err
	}

	return nil
}

func DecodeStart(message string, pmwts2 *Pmwts2) error {
	var err error

	words := strings.Split(message, ",")
	if len(words) != 8 {
		return errors.New("Invalid PMWTS2 message: " + message)
	}

	pmwts2.Source, err = strconv.Atoi(words[1])
	if err != nil {
		return err
	}

	pmwts2.LcSeqNumber, err = strconv.Atoi(words[2])
	if err != nil {
		return err
	}

	pmwts2.LcStartTime, err = strconv.ParseUint(words[3], 10, 64)
	if err != nil {
		return err
	}

	pmwts2.LcEndTime, err = strconv.ParseUint(words[4], 10, 64)
	if err != nil {
		return err
	}

	pmwts2.ClosureReason, err = strconv.Atoi(words[5])
	if err != nil {
		return err
	}

	pmwts2.LcNumberMessages, err = strconv.Atoi(words[6])
	if err != nil {
		return err
	}

	pmwts2.Direction, err = strconv.Atoi(words[7])
	if err != nil {
		return err
	}

	return nil
}

func DecodeSummary(message string, pmwts1 *Pmwts1) error {
	var err error

	words := strings.Split(message, ",")
	if len(words) != 8 {
		return errors.New("Invalid PMWTS1 message: " + message)
	}

	pmwts1.CurrentTime, err = strconv.ParseUint(words[1], 10, 64)
	if err != nil {
		return err
	}

	pmwts1.Source, err = strconv.Atoi(words[2])
	if err != nil {
		return err
	}

	pmwts1.Version = words[3]

	pmwts1.CcSeqNumber, err = strconv.Atoi(words[4])
	if err != nil {
		return err
	}

	pmwts1.CcNumberMessages, err = strconv.Atoi(words[5])
	if err != nil {
		return err
	}

	pmwts1.CcStartTime, err = strconv.ParseUint(words[6], 10, 64)
	if err != nil {
		return err
	}

	pmwts1.Direction, err = strconv.Atoi(words[7])
	if err != nil {
		return err
	}

	return nil
}

func DecodeInformation(message string, aivdm *Aivdm) {
	lines := strings.Split(message, "\n")

	for i := 0; i < len(lines); i++ {

		line := lines[i]

		words := strings.Split(line, "\\")

		for j := 0; j < len(words); j++ {

			word := words[j]
			if (len(word) > 6) && (word[0:6] == "!AIVDM") {
				parts := strings.Split(word, ",")
				aivdm.TotalSentences, _ = strconv.Atoi(parts[1])
				aivdm.SentenceNumber, _ = strconv.Atoi(parts[2])
				aivdm.EncapsulatedData = word[:len(word)-1]
			}

		}

	}
}
