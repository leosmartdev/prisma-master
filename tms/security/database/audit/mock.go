package audit

import (
	"context"
	"sync"
)

var (
	once     sync.Once
	instance Auditor
)

func mockAuditorInstance() Auditor {
	once.Do(func() {
		instance = &mockAuditor{
			records: make([]Record, 0),
		}
	})
	return instance
}

type mockAuditor struct {
	Auditor
	records []Record
}

func (auditor *mockAuditor) Record(context context.Context, record Record) error {
	auditor.records = append(auditor.records, record)
	return nil
}

func (auditor *mockAuditor) GetRecords(context context.Context, searchQuery string) ([]Record, error) {
	return auditor.records, nil
}

func (auditor *mockAuditor) GetRecordsBySessionId(context context.Context, sessionId string) ([]Record, error) {
	sessionRecords := make([]Record, 0)
	for _, record := range auditor.records {
		if record.SessionId == sessionId {
			sessionRecords = append(auditor.records, record)
		}
	}
	return sessionRecords, nil
}

func (auditor *mockAuditor) GetRecordsByUserId(context context.Context, userId string) ([]Record, error) {
	userRecords := make([]Record, 0)
	for _, record := range auditor.records {
		if record.UserId == userId {
			userRecords = append(auditor.records, record)
		}
	}
	return userRecords, nil
}
