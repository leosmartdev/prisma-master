package main

import (
	"errors"
	"fmt"
	"prisma/gogroup"
	"prisma/tms"
	api "prisma/tms/client_api"
	"prisma/tms/db/mongo"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/routing"
	"prisma/tms/rule"
	"prisma/tms/tmsg"
	"prisma/tms/util/ident"
)

// RuleStage is used for handling updates of tracks to apply rules on them
type RuleStage struct {
	ruleE rule.Engine
	n     Notifier
	fwd   *Forwarder
	// initilized is bool variable that gets set to true when the init function is done
	// only when initilized is true, we can analyze the stage
	initialized bool
}

func newRuleStage(n Notifier) *RuleStage {
	return &RuleStage{
		n:           n,
		initialized: false,
	}
}

func (rs *RuleStage) init(ctxt gogroup.GoGroup, client *mongo.MongoClient) (err error) {
	rs.ruleE, err = rule.NewTmsEngine(client, ctxt)
	if err != nil {
		return err
	}
	rs.fwd, err = NewForwarder(ctxt, *consulServer, *dc)
	if err != nil {
		return err
	}
	rs.initialized = true
	return err 
}

func (rs *RuleStage) start() {}

func (rs *RuleStage) analyze(update api.TrackUpdate) error {
	if rs.initialized == false {
		return nil
	}
	t := update.Track
	if t == nil {
		return errors.New("empty update")
	}
	actions, err := rs.ruleE.CheckRule(*t)
	if err != nil {
		return err
	}
	for _, action := range actions {
		switch actType := action.Act.(type) {
		case *rule.Rule_ThenNotice:
			priority := moc.Notice_Info
			if actType.Action == "alert" {
				priority = moc.Notice_Alert
			}
			ruleMatch := update.Status == api.Status_Current
			if err := rs.n.Notify(ruleNotice(t, action.Name, priority), ruleMatch); err != nil {
				log.Error(err.Error())
			}
		case *rule.Rule_ThenForward:
			rs.fwd.Share(*update.Track, actType.GetDc())
		default:
		}
	}
	return nil
}

func ruleNotice(track *tms.Track, name string, priority moc.Notice_Priority) *moc.Notice {
	id := ident.With("event", moc.Notice_Rule).With("track", track.RegistryId).
		With("id", track.Id).With("name", name).Hash()
	return &moc.Notice{
		NoticeId: id,
		Event:    moc.Notice_Rule,
		Priority: priority,
		Target:   TargetInfoFromTrack(track),
	}
}

func RuleChangeNotifier(ctx gogroup.GoGroup) {
	msgChan := tmsg.GClient.Listen(ctx, routing.Listener{
		MessageType: "prisma.tms.RuleInfo",
		// Destination: &tms.EndPoint{
		// 	Site: tmsg.TMSG_LOCAL_SITE,
		// 	Aid:  tmsg.APP_ID_TMSD,
		// },
	})

	for {
		select {
		case <-ctx.Done():
			return
		case m := <-msgChan:
			fmt.Println("The Rule changed", m.Body.String()) // The Rule changed id:"1" stateA:"nonstate" stateB:"ActiveSimulation"
		}
	}
}
