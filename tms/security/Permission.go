// Package security provides functions and structures to manage accesses for Rest endpoints.
package security

import (
	"context"
	"errors"

	"prisma/tms"
	"prisma/tms/marker"
	"prisma/tms/moc"
	"prisma/tms/sar"
	secContext "prisma/tms/security/context"
	"prisma/tms/security/message"
	"prisma/tms/security/policy"
	"prisma/tms/security/session"
)

const (
	RULE_CLASS_ID     = "Rule"
	POLICY_CLASS_ID   = "Policy"
	INCIDENT_CLASS_ID = "Incident"
	NOTE_CLASS_ID     = "IncidentLogEntry"
	MARKER_CLASS_ID   = "Marker"
	ICON_CLASS_ID     = "Icon"
	CLASSIDFleet      = "Fleet"
	CLASSIDVessel     = "Vessel"
	CLASSIDDevice     = "Device"
	CLASSIDSite       = "Site"
	CLASSIDConfig     = "Config"
	CLASSIDMulticast  = "Multicast"
	CLASSIDRemoteSite = "RemoteSite"
	CLASSIDSit915     = "Sit915"
	CLASSIDMessage    = "Message"
)

var (
	ErrorForbidden = errors.New("forbidden")
)

func HasPermissionForClass(ctxt context.Context, classId string) bool {
	allowed := false
	sessionId := secContext.SessionIdFromContext(ctxt)
	storedSession, err := session.GetStore(ctxt).Get(sessionId)
	if nil == err {
		permissions := PermissionsFromSession(ctxt, storedSession)
		for _, permission := range permissions {
			if classId == permission.ClassId {
				allowed = true
				break
			}
		}
	}
	return allowed
}

func HasPermissionForAction(ctxt context.Context, classId string, action string) bool {
	allowed := false
	sessionId := secContext.SessionIdFromContext(ctxt)
	storedSession, err := session.GetStore(ctxt).Get(sessionId)
	if nil == err {
		permissions := PermissionsFromSession(ctxt, storedSession)
		for _, permission := range permissions {
			if classId == permission.ClassId {
				for _, objectAction := range permission.Actions {
					if action == objectAction {
						allowed = true
						break
					}
				}
			}
		}
	}
	return allowed
}

func HasPermissionForActionOnObject(ctxt context.Context, classId string, action string, objectId string) bool {
	allowed := false
	sessionId := secContext.SessionIdFromContext(ctxt)
	storedSession, err := session.GetStore(ctxt).Get(sessionId)
	if nil == err {
		permissions := PermissionsFromSession(ctxt, storedSession)
		for _, permission := range permissions {
			if classId == permission.ClassId {
				for _, objectAction := range permission.Actions {
					if action == objectAction {
						allowed = true
						break
					}
				}
			}
		}
		if allowed {
			if "Profile" == classId {
				// access rule: only self can update profile
				allowed = storedSession.Id() == objectId
			}
		}
	}
	return allowed
}

// permissions change with the state of the session
func PermissionsFromSession(ctxt context.Context, s session.InternalSession) []*message.Permission {
	permissions := make([]*message.Permission, 0)
	if s.GetState() == message.Session_idled {
		rolesToRemove := ConsequenceSessionIdle(ctxt)
		idledRoles := make([]string, 0)
		for _, roleStringId := range s.GetRoles() {
			ok := true
			for _, roleToRemove := range rolesToRemove {
				if roleStringId == roleToRemove {
					ok = false
					break
				}
			}
			if ok {
				idledRoles = append(idledRoles, roleStringId)
			}
		}
		permissions = PermissionsFromRolesString(idledRoles)
	} else {
		permissions = PermissionsFromRolesString(s.GetRoles())
	}
	return permissions
}

// permissions for roles.  do not use for session permission check
func PermissionsFromRolesString(roleStringIds []string) []*message.Permission {
	roleIds := make([]message.RoleId, 0)
	for _, roleStringId := range roleStringIds {
		// check if valid roleId
		if roleId, ok := message.RoleId_value[roleStringId]; ok {
			// add role
			roleIds = append(roleIds, message.RoleId(roleId))
		}
	}
	return PermissionsFromRoles(roleIds)
}

func PermissionsFromRoles(roleIds []message.RoleId) []*message.Permission {
	permissions := make([]*message.Permission, 0)
loop:
	for _, roleId := range roleIds {
		switch roleId {
		case message.RoleId_StandardUser:
			permissions = append(permissions, Permissions_StandardUser...)
		case message.RoleId_RuleCreator:
			permissions = append(permissions, Permissions_RuleCreator...)
		case message.RoleId_RuleManager:
			permissions = append(permissions, Permissions_RuleManager...)
		case message.RoleId_FleetManager:
			permissions = append(permissions, Permissions_FleetManager...)
		case message.RoleId_UserManager:
			permissions = append(permissions, Permissions_UserManager...)
		case message.RoleId_IncidentManager:
			permissions = append(permissions, Permissions_IncidentManager...)
		case message.RoleId_RestrictedUser:
			permissions = append(permissions, Permissions_RestrictedUser...)
		case message.RoleId_Administrator:
			permissions = append(make([]*message.Permission, 0), Permissions_Administrator...)
			break loop
		}
	}
	return permissions
}

// role definitions
var Permissions_RestrictedUser = []*message.Permission{
	{
		ClassId: "Profile",
		Actions: []string{"READ", "UPDATE", "UPDATE_PASSWORD"},
	},
}
var Permissions_StandardUser = []*message.Permission{
	{
		ClassId: "Profile",
		Actions: []string{"READ", "UPDATE", "UPDATE_PASSWORD"},
	},
	{
		ClassId: "Zone",
		Actions: []string{"CREATE", "READ", "UPDATE", "DELETE"},
	},
	{
		ClassId: INCIDENT_CLASS_ID,
		Actions: []string{
			moc.Incident_READ.String(),
		},
	},
	{
		ClassId: NOTE_CLASS_ID,
		Actions: []string{
			moc.IncidentLogEntry_READ.String(),
		},
	},
	{
		ClassId: MARKER_CLASS_ID,
		Actions: []string{
			marker.Marker_CREATE.String(),
			marker.Marker_READ.String(),
			marker.Marker_UPDATE.String(),
			marker.Marker_DELETE.String(),
			marker.MarkerImage_CREATE.String(),
			marker.MarkerImage_READ.String(),
		},
	},
	{
		ClassId: ICON_CLASS_ID,
		Actions: []string{
			moc.Icon_CREATE.String(),
			moc.Icon_READ.String(),
			moc.Icon_UPDATE.String(),
			moc.Icon_DELETE.String(),
			moc.IconImage_CREATE.String(),
			moc.IconImage_READ.String(),
		},
	},
	{
		ClassId: "Notice",
		Actions: []string{"GET", "ACK", "TIMEOUT", "ACK_ALL"},
	},
	{
		ClassId: "View",
		Actions: []string{"CREATE", "STREAM"},
	},
	{
		ClassId: "Track",
		Actions: []string{"GET", "UPDATE", "DELETE"},
	},
	{
		ClassId: "History",
		Actions: []string{"GET"},
	},
	{
		ClassId: "Registry",
		Actions: []string{"GET"},
	},
	{
		ClassId: "Search",
		Actions: []string{"READ"},
	},
}

var Permissions_FleetManager = []*message.Permission{
	{
		ClassId: CLASSIDFleet,
		Actions: []string{
			moc.Fleet_CREATE.String(),
			moc.Fleet_READ.String(),
			moc.Fleet_UPDATE.String(),
			moc.Fleet_DELETE.String(),
			moc.Fleet_ADD_VESSEL.String(),
			moc.Fleet_REMOVE_VESSEL.String(),
		},
	},
	{
		ClassId: CLASSIDVessel,
		Actions: []string{
			moc.Vessel_CREATE.String(),
			moc.Vessel_READ.String(),
			moc.Vessel_UPDATE.String(),
			moc.Vessel_DELETE.String(),
		},
	},
	{
		ClassId: CLASSIDDevice,
		Actions: []string{
			moc.Device_CREATE.String(),
			moc.Device_READ.String(),
			moc.Device_UPDATE.String(),
			moc.Device_DELETE.String(),
		},
	},
	{
		ClassId: CLASSIDMulticast,
		Actions: []string{
			tms.Multicast_CREATE.String(),
			tms.Multicast_READ.String(),
			tms.Multicast_CANCEL.String(),
			tms.Multicast_DELETE.String(),
		},
	},
	{
		ClassId: CLASSIDRemoteSite,
		Actions: []string{
			moc.RemoteSite_READ.String(),
		},
	},
	{
		ClassId: CLASSIDSit915,
		Actions: []string{
			moc.Sit915_CREATE.String(),
			moc.Sit915_READ.String(),
			moc.Sit915_UPDATE.String(),
		},
	},
	{
		ClassId: CLASSIDMessage,
		Actions: []string{
			sar.SarsatMessage_READ.String(),
		},
	},
}

var Permissions_UserManager = []*message.Permission{
	{
		ClassId: "User",
		Actions: []string{"CREATE", "READ", "UPDATE", "DELETE", "UPDATE_ROLE",
			message.User_DEACTIVATE.String(),
		},
	},
	{
		ClassId: "Role",
		Actions: []string{"READ"},
	},
}

var Permissions_IncidentManager = []*message.Permission{
	{
		ClassId: "User",
		Actions: []string{"READ"},
	},
	{
		ClassId: "Role",
		Actions: []string{"READ"},
	},
	{
		ClassId: INCIDENT_CLASS_ID,
		Actions: []string{
			moc.Incident_CREATE.String(),
			moc.Incident_READ.String(),
			moc.Incident_UPDATE.String(),
			moc.Incident_DELETE.String(),
			moc.Incident_CLOSE.String(),
			moc.Incident_OPEN.String(),
			moc.Incident_ASSIGN.String(),
			moc.Incident_UNASSIGN.String(),
			moc.Incident_ARCHIVE.String(),
			moc.Incident_ADD_NOTE.String(),
			moc.Incident_ADD_NOTE_FILE.String(),
			moc.Incident_ADD_NOTE_ENTITY.String(),
			moc.Incident_DELETE_NOTE.String(),
			moc.Incident_TRANSFER_SEND.String(),
			moc.Incident_TRANSFER_RECEIVE.String(),
			moc.Incident_DETACH_NOTE.String(),
		},
	},
	{
		ClassId: NOTE_CLASS_ID,
		Actions: []string{
			moc.IncidentLogEntry_CREATE.String(),
			moc.IncidentLogEntry_READ.String(),
			moc.IncidentLogEntry_UPDATE.String(),
			moc.IncidentLogEntry_DELETE.String(),
			moc.IncidentLogEntry_ASSIGN.String(),
		},
	},
	{
		ClassId: MARKER_CLASS_ID,
		Actions: []string{
			marker.Marker_CREATE.String(),
			marker.Marker_READ.String(),
			marker.Marker_UPDATE.String(),
			marker.Marker_DELETE.String(),
			marker.MarkerImage_CREATE.String(),
			marker.MarkerImage_READ.String(),
		},
	},
	{
		ClassId: ICON_CLASS_ID,
		Actions: []string{
			moc.Icon_CREATE.String(),
			moc.Icon_READ.String(),
			moc.Icon_UPDATE.String(),
			moc.Icon_DELETE.String(),
			moc.IconImage_CREATE.String(),
			moc.IconImage_READ.String(),
		},
	},
	{
		ClassId: CLASSIDMulticast,
		Actions: []string{
			tms.Multicast_CREATE.String(),
			tms.Multicast_READ.String(),
			tms.Multicast_CANCEL.String(),
			tms.Multicast_DELETE.String(),
		},
	},
	{
		ClassId: CLASSIDSite,
		Actions: []string{
			moc.Site_READ.String(),
		},
	},
	{
		ClassId: CLASSIDRemoteSite,
		Actions: []string{
			moc.RemoteSite_READ.String(),
		},
	},
	{
		ClassId: "File",
		Actions: []string{"CREATE", "READ", "DELETE"},
	},
	{
		ClassId: CLASSIDSit915,
		Actions: []string{
			moc.Sit915_CREATE.String(),
			moc.Sit915_READ.String(),
			moc.Sit915_UPDATE.String(),
		},
	},
	{
		ClassId: CLASSIDMessage,
		Actions: []string{
			sar.SarsatMessage_READ.String(),
		},
	},
}

var Permissions_Administrator = []*message.Permission{
	{
		ClassId: "User",
		Actions: []string{"CREATE", "READ", "UPDATE", "DELETE", "UPDATE_ROLE",
			message.User_DEACTIVATE.String(),
		},
	},
	{
		ClassId: "Role",
		Actions: []string{"READ"},
	},
	{
		ClassId: CLASSIDConfig,
		Actions: []string{
			moc.Config_UPDATE.String(),
		},
	},
	{
		ClassId: CLASSIDSite,
		Actions: []string{
			moc.Site_CREATE.String(),
			moc.Site_READ.String(),
			moc.Site_UPDATE.String(),
			moc.Site_DELETE.String(),
		},
	},
	{
		ClassId: CLASSIDFleet,
		Actions: []string{
			moc.Fleet_CREATE.String(),
			moc.Fleet_READ.String(),
			moc.Fleet_UPDATE.String(),
			moc.Fleet_DELETE.String(),
			moc.Fleet_ADD_VESSEL.String(),
			moc.Fleet_REMOVE_VESSEL.String(),
		},
	},
	{
		ClassId: RULE_CLASS_ID,
		Actions: []string{
			moc.RuleAction_CREATE.String(),
			moc.RuleAction_READ.String(),
			moc.RuleAction_UPDATE.String(),
			moc.RuleAction_DELETE.String(),
			moc.RuleAction_STATE.String(),
		},
	},
	{
		ClassId: CLASSIDVessel,
		Actions: []string{
			moc.Vessel_CREATE.String(),
			moc.Vessel_READ.String(),
			moc.Vessel_UPDATE.String(),
			moc.Vessel_DELETE.String(),
		},
	},
	{
		ClassId: CLASSIDDevice,
		Actions: []string{
			moc.Device_CREATE.String(),
			moc.Device_READ.String(),
			moc.Device_UPDATE.String(),
			moc.Device_DELETE.String(),
		},
	},
	{
		ClassId: "Object",
		Actions: []string{"READ"},
	},
	{
		ClassId: "Audit",
		Actions: []string{"READ"},
	},
	{
		ClassId: INCIDENT_CLASS_ID,
		Actions: []string{
			moc.Incident_CREATE.String(),
			moc.Incident_READ.String(),
			moc.Incident_UPDATE.String(),
			moc.Incident_DELETE.String(),
			moc.Incident_CLOSE.String(),
			moc.Incident_OPEN.String(),
			moc.Incident_ASSIGN.String(),
			moc.Incident_UNASSIGN.String(),
			moc.Incident_ARCHIVE.String(),
			moc.Incident_ADD_NOTE.String(),
			moc.Incident_ADD_NOTE_FILE.String(),
			moc.Incident_ADD_NOTE_ENTITY.String(),
			moc.Incident_DELETE_NOTE.String(),
			moc.Incident_TRANSFER_SEND.String(),
			moc.Incident_TRANSFER_RECEIVE.String(),
			moc.Incident_DETACH_NOTE.String(),
		},
	},
	{
		ClassId: NOTE_CLASS_ID,
		Actions: []string{
			moc.IncidentLogEntry_CREATE.String(),
			moc.IncidentLogEntry_READ.String(),
			moc.IncidentLogEntry_UPDATE.String(),
			moc.IncidentLogEntry_DELETE.String(),
			moc.IncidentLogEntry_ASSIGN.String(),
		},
	},
	{
		ClassId: MARKER_CLASS_ID,
		Actions: []string{
			marker.Marker_CREATE.String(),
			marker.Marker_READ.String(),
			marker.Marker_UPDATE.String(),
			marker.Marker_DELETE.String(),
			marker.MarkerImage_CREATE.String(),
			marker.MarkerImage_READ.String(),
		},
	},
	{
		ClassId: ICON_CLASS_ID,
		Actions: []string{
			moc.Icon_CREATE.String(),
			moc.Icon_READ.String(),
			moc.Icon_UPDATE.String(),
			moc.Icon_DELETE.String(),
			moc.IconImage_CREATE.String(),
			moc.IconImage_READ.String(),
		},
	},
	{
		ClassId: "Profile",
		Actions: []string{"READ", "UPDATE", "UPDATE_PASSWORD"},
	},
	{
		ClassId: "Zone",
		Actions: []string{"CREATE", "READ", "UPDATE", "DELETE"},
	},
	{
		ClassId: "Geofence",
		Actions: []string{"CREATE", "READ", "UPDATE", "DELETE"},
	},
	{
		ClassId: CLASSIDMulticast,
		Actions: []string{
			tms.Multicast_CREATE.String(),
			tms.Multicast_READ.String(),
			tms.Multicast_CANCEL.String(),
			tms.Multicast_DELETE.String(),
		},
	},
	{
		ClassId: "Notice",
		Actions: []string{"GET", "ACK", "TIMEOUT", "ACK_ALL"},
	},
	{
		ClassId: "File",
		Actions: []string{"CREATE", "READ", "DELETE"},
	},
	{
		ClassId: "View",
		Actions: []string{"CREATE", "STREAM"},
	},
	{
		ClassId: "Track",
		Actions: []string{"GET", "UPDATE", "DELETE"},
	},
	{
		ClassId: "History",
		Actions: []string{"GET"},
	},
	{
		ClassId: "Search",
		Actions: []string{"READ"},
	},
	{
		ClassId: "Registry",
		Actions: []string{"GET"},
	},
	{
		ClassId: "Config",
		Actions: []string{"UPDATE"},
	},
	{
		ClassId: CLASSIDRemoteSite,
		Actions: []string{
			moc.RemoteSite_CREATE.String(),
			moc.RemoteSite_READ.String(),
			moc.RemoteSite_UPDATE.String(),
			moc.RemoteSite_DELETE.String(),
		},
	},
	{
		ClassId: POLICY_CLASS_ID,
		Actions: []string{
			policy.Policy_READ.String(),
			policy.Policy_UPDATE.String(),
		},
	},
	{
		ClassId: CLASSIDSit915,
		Actions: []string{
			moc.Sit915_CREATE.String(),
			moc.Sit915_READ.String(),
			moc.Sit915_UPDATE.String(),
		},
	},
	{
		ClassId: CLASSIDMessage,
		Actions: []string{
			sar.SarsatMessage_READ.String(),
		},
	},
}

var Permissions_RuleCreator = []*message.Permission{
	{
		ClassId: RULE_CLASS_ID,
		Actions: []string{
			moc.RuleAction_CREATE.String(),
			moc.RuleAction_READ.String(),
			moc.RuleAction_DELETE.String(),
		},
	},
}

var Permissions_RuleManager = []*message.Permission{
	{
		ClassId: RULE_CLASS_ID,
		Actions: []string{
			moc.RuleAction_READ.String(),
			moc.RuleAction_UPDATE.String(),
			moc.RuleAction_DELETE.String(),
			moc.RuleAction_STATE.String(),
		},
	},
}
