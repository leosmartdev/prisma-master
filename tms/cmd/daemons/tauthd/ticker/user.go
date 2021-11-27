// Package Ticker manages the expiration of sessions.
package ticker

import (
	"context"
	"prisma/tms/security"
	"prisma/tms/security/database"
	"prisma/tms/security/database/audit"
	"prisma/tms/security/message"
	"time"
)

func UserTicker(ctxt context.Context) *time.Ticker {
	tickChan := time.NewTicker(time.Minute * 5).C
	for {
		select {
		case <-tickChan:
			processUsers(ctxt)
		}
	}
}

func processUsers(ctxt context.Context) {
	users, err := database.FindAllNotDisabled(ctxt)
	if err == nil {
		auditor := audit.NewAuditor(ctxt)
		searchMap := make(map[string]string, 0)
		searchMap["action"] = "AUTHENTICATE"
		searchMap["classId"] = "User"
		for _, user := range users {
			searchMap["userId"] = string(user.UserId)
			records, err := auditor.GetRecordsByMapByTimeQuery(ctxt, searchMap, audit.TimeQuery{Limit: 1}, "")
			if err == nil && len(records) > 0 {
				inactiveDuration := time.Since(records[0].Created)
				// check policy
				enforce, userState := security.EnforceInactiveDuration(ctxt, inactiveDuration)
				if enforce {
					if user.State != userState {
						user.State = userState
						action := message.User_LOCK.String()
						if message.User_deactivated == user.State {
							action = message.User_DEACTIVATE.String()
						}
						_, err = database.Update(ctxt, &user)
						if err == nil {
							security.AuditUserObject(ctxt, "User", string(user.UserId), "SYSTEM", action, "SUCCESS")
						} else {
							security.AuditUserObject(ctxt, "User", string(user.UserId), "SYSTEM", action, "FAIL_ERROR", err)
						}
					}
				}
			}
		}
	}
}
