package rule

import (
	"prisma/tms"
	"prisma/tms/db/mongo"
	"sync"
	"prisma/gogroup"
	"prisma/tms/log"
)

// Actions is an interface to determine what to do the condition of a rule
type Action interface {
}

// ActionMessage contains an action with name to determine the action for named rule
type ActionMessage struct {
	Act  Action
	Name string
}

type fncCmp func(val, valOrigin interface{}) (bool, error)

// Engine is an interface for working with rules
type Engine interface {
	Store
	// Check a track on matching to any rules.
	// Return actions for doing, an error - any errors
	CheckRule(track tms.Track) ([]ActionMessage, error)
}

// TmsEngine is an implementation for the rule.Engine. It uses mongodb and a RAM hashMap for hot access
// So be ensure that you don't put any rules into mongodb directly
// Also it can be used without mongo, it will work using the RAM hashMap only
type TmsEngine struct {
	mu            sync.Mutex
	storage       Store
	zoneStorage   ZoneStore
	objectsInZone map[string]map[string]bool
}

// NewTmsEngine returns an instance of tmsEngine and init a hashMap
// For using without mongodb pass nil
func NewTmsEngine(client *mongo.MongoClient, ctxt gogroup.GoGroup) (*TmsEngine, error) {
	var (
		storage     *StorageMongoDBRule
		zoneStorage *TmsZoneStorage
		err         error
	)
	if storage, err = NewStorageMongoDBRule(client); err != nil {
		return nil, err
	}
	if zoneStorage, err = NewTmsZoneStorage(client, ctxt); client != nil && err != nil {
		return nil, err
	}

	return &TmsEngine{
		storage:       storage,
		zoneStorage:   zoneStorage,
		objectsInZone: make(map[string]map[string]bool),
	}, nil
}

func (t *TmsEngine) GetAll() []*Rule {
	return t.storage.GetAll()
}

func (t *TmsEngine) GetByType(operandType OperandType) []*Rule {
	return t.storage.GetByType(operandType)
}

func (t *TmsEngine) GetRule(id string) (*Rule, error) {
	return t.storage.GetRule(id)
}

func (t *TmsEngine) DeleteRule(id string) error {
	return t.storage.DeleteRule(id)
}

func (t *TmsEngine) UpsertRule(rule Rule) error {
	return t.storage.UpsertRule(rule)
}

func (t *TmsEngine) CheckRule(track tms.Track) ([]ActionMessage, error) {
	var chThenRule []ActionMessage
	for _, chRule := range t.storage.GetAll() {
		if t.checkRuleTree(*chRule, track) {
			switch {
			case chRule.Notice != nil:
				chThenRule = append(chThenRule, ActionMessage{chRule.Notice, chRule.Name})
			case chRule.Forward != nil:
				chThenRule = append(chThenRule, ActionMessage{chRule.Forward, chRule.Name})
			default:
				chThenRule = append(chThenRule, ActionMessage{nil, chRule.Name})
				log.Warn("The action is not provided")
			}
		}
	}
	return chThenRule, nil
}

// Make a boolean tree for a rule and calculate that
// Advantages:
// - easy developing
// - easy supporting
// Disadvantages:
// - It computes ALL conditions for making the boolean tree. It could be improved(a lazy algorithm)
func (t *TmsEngine) checkRuleTree(chRule Rule, track tms.Track) bool {
	tree := new(TreeExpression)
	root := tree
	var whatCheck interface{}
	if chRule.All != nil {
		whatCheck = chRule.All
	} else {
		whatCheck = chRule.Any
	}
	// make a tree
outerLoop:
	for {
		switch cond := whatCheck.(type) {
		case *Rule_IfAll:
			if cond == nil {
				break outerLoop
			}
			switch cond.OperandType {
			case OperandType_TARGET:
				tree.Result = checkRuleForTargetAll(*cond, track)
			case OperandType_ZONE:
				t.mu.Lock()
				tree.Result = checkRuleForZoneAll(*cond, track, t.zoneStorage, t.objectsInZone)
				t.mu.Unlock()
			case OperandType_TRACK:
				tree.Result = checkRuleForTrackAll(*cond, track)
			default:
				log.Error("undefined operand type %v", cond.OperandType)
				return false
			}
			if cond.All != nil {
				whatCheck = cond.All
			} else if cond.Any != nil {
				whatCheck = cond.Any
			} else {
				break outerLoop
			}
			tree.Right = new(TreeExpression)
			tree = tree.Right
		case *Rule_IfAny:
			if cond == nil {
				break outerLoop
			}
			switch cond.OperandType {
			case OperandType_TARGET:
				tree.Result = checkRuleForTargetAny(*cond, track)
			case OperandType_ZONE:
				t.mu.Lock()
				tree.Result = checkRuleForZoneAny(*cond, track, t.zoneStorage, t.objectsInZone)
				t.mu.Unlock()
			case OperandType_TRACK:
				tree.Result = checkRuleForTrackAny(*cond, track)
			default:
				log.Error("undefined operand type %v", cond.OperandType)
				return false
			}
			if cond.All != nil {
				whatCheck = cond.All
			} else if cond.Any != nil {
				whatCheck = cond.Any
			} else {
				break outerLoop
			}
			tree.Left = new(TreeExpression)
			tree = tree.Left
		default:
			log.Error("bad interface %v %T", log.Spew(cond), cond)
			break outerLoop
		}
	}
	return root.Calculate()
}
