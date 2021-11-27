package public

import (
	"fmt"
	"net/http"
	"strconv"

	"prisma/gogroup"
	"prisma/tms"
	"prisma/tms/db/mongo"
	"prisma/tms/moc"
	"prisma/tms/rule"
	"prisma/tms/security"
	"prisma/tms/tmsg"

	"golang.org/x/net/context"

	"prisma/tms/log"

	restful "github.com/orolia/go-restful" 
	"github.com/golang/protobuf/jsonpb"
	"prisma/tms/rest"
)

type RuleRest struct {
	RuleEngine *rule.TmsEngine
	group      gogroup.GoGroup
}

func newAuditor(classID, actionID string) func(reqCtx context.Context, result string, payload ...interface{}) {
	return func(reqCtx context.Context, result string, payload ...interface{}) {
		security.Audit(reqCtx, classID, actionID, result, payload...)
	}
}

func NewRuleRest(mc *mongo.MongoClient, group gogroup.GoGroup) (*RuleRest, error) {
	re, err := rule.NewTmsEngine(mc, group)
	if err != nil {
		return nil, err
	}

	return &RuleRest{
		group:      group,
		RuleEngine: re,
	}, nil
}

// Meta returns metadata of rules. For example which fields can apply which values
func (rr *RuleRest) Meta(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.RuleAction_READ.String()

	if !authorized(req, rsp, security.RULE_CLASS_ID, ACTION) {
		return
	}
	metadata := new(rest.RuleMetadata)
	for _, val := range rule.OperandType_name {
		metadata.OperandType = append(metadata.OperandType, val)
	}
	rsp.WriteHeaderAndEntity(http.StatusOK, metadata)
}

func (rr *RuleRest) Post(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.RuleAction_CREATE.String()
	reqCtx := req.Request.Context()
	audit := newAuditor(security.RULE_CLASS_ID, ACTION)

	if authorized(req, rsp, security.RULE_CLASS_ID, ACTION) {

		newRule := new(rule.Rule)
		err := jsonpb.Unmarshal(req.Request.Body, newRule) // req.ReadEntity(newRule)
		if err != nil {
			audit(reqCtx, "FAIL_ERROR")
			rsp.WriteError(http.StatusBadRequest, fmt.Errorf("jsonpb.Unmarshal(req.Request.Body, newRule): %s", err))
			return
		}

		err = rr.RuleEngine.UpsertRule(*newRule)
		log.Debug(log.Spew(err, rr))
		if err != nil {
			audit(reqCtx, "FAIL_ERROR")
			rsp.WriteError(http.StatusNotFound, fmt.Errorf("RuleEngine.UpsertRule(*newRule): %s", err))
			return
		}

		audit(reqCtx, "SUCCESS", newRule.Id)
		rsp.WriteHeaderAndEntity(http.StatusCreated, newRule)
	}
}

func (rr *RuleRest) Get(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.RuleAction_READ.String()
	reqCtx := req.Request.Context()
	audit := newAuditor(security.RULE_CLASS_ID, ACTION)

	if authorized(req, rsp, security.RULE_CLASS_ID, ACTION) {

		arr := rr.RuleEngine.GetAll()
		audit(reqCtx, "SUCCESS", arr)
		rsp.WriteHeaderAndEntity(http.StatusOK, arr)
	}
}

func (rr *RuleRest) GetOne(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.RuleAction_READ.String()
	reqCtx := req.Request.Context()
	audit := newAuditor(security.RULE_CLASS_ID, ACTION)

	if authorized(req, rsp, security.RULE_CLASS_ID, ACTION) {

		id := req.PathParameter("rule-id")
		aRule, err := rr.RuleEngine.GetRule(id)
		if err != nil {
			audit(reqCtx, "FAIL_ERROR")
			rsp.WriteError(http.StatusNotFound, fmt.Errorf("RuleEngine.GetRule(id): %s", err))
			return
		}

		audit(reqCtx, "SUCCESS", aRule.Id)
		rsp.WriteHeaderAndEntity(http.StatusOK, aRule)
	}
}

func (rr *RuleRest) Put(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.RuleAction_UPDATE.String()
	reqCtx := req.Request.Context()
	audit := newAuditor(security.RULE_CLASS_ID, ACTION)

	if authorized(req, rsp, security.RULE_CLASS_ID, ACTION) {

		aRule := new(rule.Rule)
		err := jsonpb.Unmarshal(req.Request.Body, aRule) // req.ReadEntity(newRule)
		if err != nil {
			audit(reqCtx, "FAIL_ERROR")
			if err := rsp.WriteError(http.StatusBadRequest,
				fmt.Errorf("jsonpb.Unmarshal(req.Request.Body, newRule): %s", err)); err != nil {
				log.Error(err.Error())
			}
			return
		}

		err = rr.RuleEngine.UpsertRule(*aRule)
		if err != nil {
			audit(reqCtx, "FAIL_ERROR")
			rsp.WriteError(http.StatusNotFound, fmt.Errorf("RuleEngine.UpsertRule(*aRule): %s", err))
			return
		}

		audit(reqCtx, "SUCCESS", aRule.Id)
		rsp.WriteHeaderAndEntity(http.StatusOK, aRule)
	}
}

func NotifyTanalyzedAboutChangedRule(reqCtx context.Context, id, a, b string) (err error) {
	body, err := tmsg.PackFrom(&tms.RuleInfo{
		Id:     id,
		StateA: a,
		StateB: b,
	})
	if err != nil {
		return
	}

	tmsg.GClient.Send(reqCtx, &tms.TsiMessage{
		Destination: []*tms.EndPoint{
			&tms.EndPoint{
				Site: tmsg.TMSG_HQ_SITE,
			},
		},
		WriteTime: tms.Now(),
		SendTime:  tms.Now(),
		Body:      body,
	})

	return
}

func (rr *RuleRest) UpdateState(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.RuleAction_STATE.String()
	reqCtx := req.Request.Context()
	audit := newAuditor(security.RULE_CLASS_ID, ACTION)

	if authorized(req, rsp, security.RULE_CLASS_ID, ACTION) {

		rid := req.PathParameter("rule-id")
		aRule, err := rr.RuleEngine.GetRule(rid)
		if err != nil {
			audit(reqCtx, "FAIL_ERROR")
			rsp.WriteError(http.StatusNotFound, fmt.Errorf("RuleEngine.GetRule(rid): %s", err))
			return
		}

		sid := req.PathParameter("state-id")
		state, err := strconv.ParseUint(sid, 10, 32)
		if err != nil {
			audit(reqCtx, "FAIL_ERROR")
			rsp.WriteError(http.StatusNotFound, fmt.Errorf("strconv.ParseUint(sid): %s", err))
			return
		}

		bRuleState := rule.Rule_State(state)
		// tanalyzed will need to know when rule have changed
		if aRule.State != bRuleState {
			err = NotifyTanalyzedAboutChangedRule(reqCtx, aRule.Id, aRule.State.String(), bRuleState.String())
			if err != nil {
				audit(reqCtx, "FAIL_ERROR")
				rsp.WriteError(http.StatusNotFound, fmt.Errorf("tmsg.PackFrom(&tms.RuleInfo{...}): %s", err))
				return
			}
		}

		aRule.State = bRuleState
		err = rr.RuleEngine.UpsertRule(*aRule)
		if err != nil {
			audit(reqCtx, "FAIL_ERROR")
			rsp.WriteError(http.StatusNotFound, fmt.Errorf("RuleEngine.UpsertRule(*aRule): %s", err))
			return
		}

		audit(reqCtx, "SUCCESS", aRule.Id)
		rsp.WriteHeaderAndEntity(http.StatusOK, aRule)
	}
}

func (rr *RuleRest) Delete(req *restful.Request, rsp *restful.Response) {
	ACTION := moc.RuleAction_DELETE.String()
	reqCtx := req.Request.Context()
	audit := newAuditor(security.RULE_CLASS_ID, ACTION)

	if authorized(req, rsp, security.RULE_CLASS_ID, ACTION) {

		id := req.PathParameter("rule-id")

		err := rr.RuleEngine.DeleteRule(id)
		if err != nil {
			audit(reqCtx, ACTION, "FAIL_ERROR")
			rsp.WriteError(http.StatusNotFound, fmt.Errorf("RuleEngine.GetRule(id): %s", err))
			return
		}

		audit(reqCtx, "SUCCESS", id)
		rsp.WriteHeaderAndEntity(http.StatusOK, id)
	}
}
