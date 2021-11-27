//+build integration

package main

import (
	"prisma/tms/ingenu"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIngenuApi_RequestToken(t *testing.T) {
	ia := NewIngenuAPI("https://glds.ingenu.com", "orolia@orolia.com", "0r011a_McMurd0%")
	if assert.NoError(t, ia.requestToken()) {
		assert.NotEmpty(t, ia.token)
	}

}

func TestIngenuApi_Get(t *testing.T) {
	ia := NewIngenuAPI("https://glds.ingenu.com/", "orolia@orolia.com", "0r011a_McMurd0%")
	data, err := ia.Get("https://glds.ingenu.com/data/v1/receive/")
	if assert.NoError(t, err) {
		assert.NotEmpty(t, data)
	}

}

//This is not really a good unit test for posting to ingenu, needs to be refactored.
func TestIngenuApi_Post(t *testing.T) {
	ia := NewIngenuAPI("https://glds.ingenu.com/", "orolia@orolia.com", "0r011a_McMurd0%")
	body := &ingenu.DatagramDownlinkRequest{
		Nodeid:  "testNodeId",
		Payload: "testpayload",
		Tag:     "testtag",
	}
	data, err := ia.Post("https://glds.ingenu.com/data/v1/receive/", body)
	if assert.NoError(t, err) {
		assert.NotEmpty(t, data)
	}

}
