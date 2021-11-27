package moc

import (
	"testing"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/json-iterator/go/assert"
)

func TestHasIncidentLogEntityy(t *testing.T) {
	incLog := []*IncidentLogEntry{
		{
			Type: "FAKE",
			Entity: &EntityRelationship{
				Id: "test_id",
			},
			Timestamp: &timestamp.Timestamp{
				Seconds: 1,
			},
		},
		{
			Type: "TRACK",
			Entity: &EntityRelationship{
				Id: "test_id",
			},
			Timestamp: &timestamp.Timestamp{
				Seconds: 2,
			},
		},
		{
			Type: "TRACK",
			Entity: &EntityRelationship{
				Id: "test_id",
			},
			Timestamp: &timestamp.Timestamp{
				Seconds: 3,
			},
			Deleted: true,
		},
	}
	assert.False(t, HasIncidentLogEntity(incLog, "test_id"))
	incLog[2].Deleted = false
	incLog[1].Deleted = true
	assert.True(t, HasIncidentLogEntity(incLog, "test_id"))
	incLog[0].Entity.Id = "test_fake"
	assert.False(t, HasIncidentLogEntity(incLog, "test_fake"))
}
