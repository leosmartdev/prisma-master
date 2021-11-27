package tmsg

import (
	. "prisma/tms"
	. "prisma/tms/routing"
)

func MatchesEP(l *EndPoint, tgt *EndPoint, lclSite uint32) bool {
	if l == nil {
		return true
	}

	if tgt == nil {
		return false
	}

	if l.Site != 0 && l.Site != tgt.Site && !(l.Site == TMSG_LOCAL_SITE && tgt.Site == lclSite) {
		return false
	}
	if l.Aid != 0 && l.Aid != tgt.Aid {
		return false
	}
	if l.Eid != 0 && l.Eid != tgt.Eid {
		return false
	}
	if l.Pid != 0 && l.Pid != tgt.Pid {
		return false
	}

	return true
}

func Matches(l Listener, msg *TsiMessage, lclSite uint32) bool {
	if !MatchesEP(l.Source, msg.Source, lclSite) {
		return false
	}

	foundDstMatch := l.Destination == nil
	for _, dest := range msg.Destination {
		foundDstMatch = foundDstMatch || MatchesEP(l.Destination, dest, lclSite)
	}
	if !foundDstMatch {
		return false
	}

	if l.MessageType != "" && l.MessageType != MessageType(msg) {
		return false
	}

	return true
}
