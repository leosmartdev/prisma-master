package rule

import (
	"prisma/tms"
	"prisma/tms/moc"
	"testing"

	google_protobuf1 "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/assert"
)

var _ = google_protobuf1.BoolValue{}

var mapRules = map[string]Rule{
	"Target registry id == 196352 && speed == 5": {
		Id:   "idTest1",
		Name: "test MMSI==97",
		All: &Rule_IfAll{
			OperandType: OperandType_TRACK,
			Operation:   Rule_TOWARD,
			Track: &tms.Track{
				RegistryId: "196352",
			},
			All: &Rule_IfAll{
				OperandType: OperandType_TARGET,
				Operation:   Rule_EQUAL,
				Target: &tms.Target{
					Speed: &google_protobuf1.DoubleValue{Value: 5},
				},
				CheckFields: []string{"Speed"},
			},
		},
	},
	"bad second operandType": {
		Id:   "idTest2",
		Name: "test MMSI==97",
		All: &Rule_IfAll{
			OperandType: OperandType_TRACK,
			Operation:   Rule_TOWARD,
			Track: &tms.Track{
				RegistryId: "196352",
			},
			All: &Rule_IfAll{
				OperandType: 111111,
				Operation:   Rule_EQUAL,
				Target: &tms.Target{
					Speed: &google_protobuf1.DoubleValue{Value: 5},
				},
				CheckFields: []string{"Speed"},
			},
		},
	},
	"MMSIEqual97RuleAny": {
		Id:   "idTest3",
		Name: "test MMSI==97",
		Any: &Rule_IfAny{
			OperandType: OperandType_TARGET,
			Operation:   Rule_EQUAL,
			Target: &tms.Target{
				Mmsi: "97",
			},
			CheckFields: []string{"Mmsi"},
		},
	},
	"Deep structure imei == IOJoijqw": {
		Id:   "idTest4",
		Name: "imei == IOJoijqw",
		Any: &Rule_IfAny{
			OperandType: OperandType_TARGET,
			Operation:   Rule_EQUAL,
			Target: &tms.Target{
				Imei: &google_protobuf1.StringValue{Value: "IOJoijqw"},
			},
			CheckFields: []string{"Imei.Value"},
		},
	},
	"Deep structure imei == IOJoijqw ALL": {
		Id:   "idTest5",
		Name: "imei == IOJoijqw",
		All: &Rule_IfAll{
			OperandType: OperandType_TARGET,
			Operation:   Rule_EQUAL,
			Target: &tms.Target{
				Imei: &google_protobuf1.StringValue{Value: "IOJoijqw"},
			},
			CheckFields: []string{"Imei.Value"},
		},
	},
	//Metadata has no iridium field anymore.
	/*	"MetaData deep": {
			Id:   "idTest6",
			Name: "imei == IOJoijqw",
			All: &Rule_IfAll{
				OperandType: "Metadata",
				Operation:   Rule_EQUAL,
				Metadata: &tms.TrackMetadata{
					Iridium: &iridium.Iridium{
						Moh: &iridium.MobileOriginatedHeader{
							IMEI: "IOJoijqw",
						},
						Mth: &iridium.MobileTerminatedHeader{
							MTflag: 21,
						},
					},
				},
				CheckFields: []string{"Iridium.Moh.IMEI"},
			},
		},
		"MetaData deep icase": {
			Id:   "idTest7",
			Name: "imei == IOJoijqw",
			All: &Rule_IfAll{
				OperandType: "Metadata",
				Operation:   Rule_EQUAL,
				Metadata: &tms.TrackMetadata{
					Iridium: &iridium.Iridium{
						Moh: &iridium.MobileOriginatedHeader{
							IMEI: "IOJoijqw",
						},
						Mth: &iridium.MobileTerminatedHeader{
							MTflag: 21,
						},
					},
				},
				CheckFields: []string{"iridium.moh.imei"},
			},
		},*/
	"Bad naming in Deep structure imei == IOJoijqw": {
		Id:   "idTest8",
		Name: "imei == IOJoijqw",
		Any: &Rule_IfAny{
			OperandType: OperandType_TARGET,
			Operation:   Rule_EQUAL,
			Target: &tms.Target{
				Imei: &google_protobuf1.StringValue{Value: "IOJoijqw"},
			},
			CheckFields: []string{"Iridium.Bad_Moh.IMEI"},
		},
	},
	"MMSIEqual97RuleAll": {
		Id:   "idTest9",
		Name: "test MMSI==97",
		All: &Rule_IfAll{
			OperandType: OperandType_TARGET,
			Operation:   Rule_EQUAL,
			Target: &tms.Target{
				Mmsi: "97",
			},
			CheckFields: []string{"Mmsi"},
		},
	},
	"test MMSI==97 && (speed == 5 || speed == 6)": {
		Id: "idTest10",
		Name: "test MMSI==97 && (speed == 5 || speed == 6) that's can be transformed to" +
			"speed == 5 || speed == 6 && MMSI==97",
		Any: &Rule_IfAny{
			OperandType: OperandType_TARGET,
			Operation:   Rule_EQUAL,
			Target: &tms.Target{
				Speed: &google_protobuf1.DoubleValue{Value: 5.0},
			},
			CheckFields: []string{"Speed"},
			All: &Rule_IfAll{
				OperandType: OperandType_TARGET,
				Operation:   Rule_EQUAL,
				Target: &tms.Target{
					Speed: &google_protobuf1.DoubleValue{Value: 6.0},
				},
				CheckFields: []string{"Speed"},
				All: &Rule_IfAll{
					OperandType: OperandType_TARGET,
					Operation:   Rule_EQUAL,
					Target: &tms.Target{
						Mmsi: "97",
					},
					CheckFields: []string{"Mmsi"},
				},
			},
		},
	},
	"Big depth": {
		Id:   "idTest11",
		Name: "test MMSI==97 && (speed == 5 || (speed == 6 && course == 20))",
		All: &Rule_IfAll{
			OperandType: OperandType_TARGET,
			Operation:   Rule_EQUAL,
			Target: &tms.Target{
				Mmsi: "97",
			},
			CheckFields: []string{"Mmsi"},
			All: &Rule_IfAll{
				OperandType: OperandType_TARGET,
				Operation:   Rule_EQUAL,
				Target: &tms.Target{
					Speed: &google_protobuf1.DoubleValue{Value: 5.0},
				},
				CheckFields: []string{"Speed"},
				Any: &Rule_IfAny{
					OperandType: OperandType_TARGET,
					Operation:   Rule_EQUAL,
					Target: &tms.Target{
						Speed: &google_protobuf1.DoubleValue{Value: 6.0},
					},
					CheckFields: []string{"Speed"},
					Any: &Rule_IfAny{
						OperandType: OperandType_TARGET,
						Operation:   Rule_EQUAL,
						Target: &tms.Target{
							Course: &google_protobuf1.DoubleValue{Value: 20.0},
						},
						CheckFields: []string{"Course"},
						All: &Rule_IfAll{
							OperandType: OperandType_ZONE,
						},
					},
				},
			},
		},
	},
	"speed > 5 && speed <= 10": {
		Id:   "idTest12",
		Name: "speed > 5 && speed <= 10",
		All: &Rule_IfAll{
			OperandType: OperandType_TARGET,
			Operation:   Rule_GREATER,
			Target: &tms.Target{
				Speed: &google_protobuf1.DoubleValue{Value: 5.0},
			},
			CheckFields: []string{"Speed"},
			Any: &Rule_IfAny{
				OperandType: OperandType_TARGET,
				Operation:   Rule_LESSER_EQUAL,
				Target: &tms.Target{
					Speed: &google_protobuf1.DoubleValue{Value: 10.0},
				},
				CheckFields: []string{"Speed"},
			},
		},
	},
	"in zone{X:1.0, X1:2.0, Y:1.0, Y1:2.0}": {
		Id:   "idTest13",
		Name: "in zone{X:1.0, X1:2.0, Y:1.0, Y1:2.0}",
		All: &Rule_IfAll{
			OperandType: OperandType_ZONE,
			Operation:   Rule_IN,
			Zone: &moc.Zone{
				Poly: &tms.Polygon{
					Lines: []*tms.LineString{
						&tms.LineString{
							Points: []*tms.Point{
								&tms.Point{
									Latitude:  1,
									Longitude: 1,
								},
								&tms.Point{
									Latitude:  2,
									Longitude: 1,
								},
								&tms.Point{
									Latitude:  2,
									Longitude: 2,
								},
								&tms.Point{
									Latitude:  1,
									Longitude: 2,
								},
								&tms.Point{
									Latitude:  1,
									Longitude: 1,
								},
							},
						},
					},
				},
			},
		},
	},
	"in zone{X:1.0, X1:2.0, Y:1.0, Y1:2.0} Any": {
		Id:   "idTest14",
		Name: "in zone{X:1.0, X1:2.0, Y:1.0, Y1:2.0}",
		Any: &Rule_IfAny{
			OperandType: OperandType_ZONE,
			Operation:   Rule_IN,
			Zone: &moc.Zone{
				Poly: &tms.Polygon{
					Lines: []*tms.LineString{
						&tms.LineString{
							Points: []*tms.Point{
								&tms.Point{
									Latitude:  1,
									Longitude: 1,
								},
								&tms.Point{
									Latitude:  2,
									Longitude: 1,
								},
								&tms.Point{
									Latitude:  2,
									Longitude: 2,
								},
								&tms.Point{
									Latitude:  1,
									Longitude: 2,
								},
								&tms.Point{
									Latitude:  1,
									Longitude: 1,
								},
							},
						},
					},
				},
			},
		},
	},
	/*	"MetaData deep icase Any": {
		Id:   "idTest15",
		Name: "imei == IOJoijqw",
		Any: &Rule_IfAny{
			OperandType: "Metadata",
			Operation:   Rule_EQUAL,
			Metadata: &tms.TrackMetadata{
				Iridium: &iridium.Iridium{
					Moh: &iridium.MobileOriginatedHeader{
						IMEI: "IOJoijqw",
					},
					Mth: &iridium.MobileTerminatedHeader{
						MTflag: 21,
					},
				},
			},
			CheckFields: []string{"iridium.moh.imei"},
		},
	},*/
	"Target registry id == 196352 || speed == 5000": {
		Id:   "idTest16",
		Name: "test MMSI==97",
		Any: &Rule_IfAny{
			OperandType: OperandType_TRACK,
			Operation:   Rule_TOWARD,
			Track: &tms.Track{
				RegistryId: "196352",
			},
			All: &Rule_IfAll{
				OperandType: OperandType_TARGET,
				Operation:   Rule_EQUAL,
				Target: &tms.Target{
					Speed: &google_protobuf1.DoubleValue{Value: 5},
				},
				CheckFields: []string{"Speed"},
			},
		},
	},
}

func TestTmsEngine_UpsertRule(t *testing.T) {
	tmsEngine, err := NewTmsEngine(nil, nil)
	assert.NoError(t, err)
	// test insert
	for _, valueRule := range mapRules {
		assert.NoError(t, tmsEngine.UpsertRule(valueRule))
	}
	// test update
	assert.NoError(t, tmsEngine.UpsertRule(Rule{Id: "idTest1", Name: "q"}))
	updatedField, err := tmsEngine.GetRule("idTest1")
	assert.NoError(t, err)
	assert.Equal(t, &Rule{Id: "idTest1", Name: "q"}, updatedField)
}

func TestTmsEngine_DeleteRule(t *testing.T) {
	tmsEngine, err := NewTmsEngine(nil, nil)
	assert.NoError(t, err)
	// test insert
	for _, valueRule := range mapRules {
		assert.NoError(t, tmsEngine.UpsertRule(valueRule))
		assert.NoError(t, err)
	}
	for _, valueRule := range mapRules {
		assert.NoError(t, tmsEngine.DeleteRule(valueRule.Id))
		assert.NoError(t, err)
	}
	assert.Len(t, tmsEngine.storage.GetAll(), 0)
}

func TestTmsEngine_GetAll(t *testing.T) {
	tmsEngine, err := NewTmsEngine(nil, nil)
	assert.NoError(t, err)
	// test insert
	for _, valueRule := range mapRules {
		assert.NoError(t, tmsEngine.UpsertRule(valueRule))
	}
	assert.Len(t, tmsEngine.GetAll(), len(mapRules))
}

func TestTmsEngine_GetByType(t *testing.T) {
	tmsEngine, err := NewTmsEngine(nil, nil)
	assert.NoError(t, err)
	// test insert
	for _, valueRule := range mapRules {
		assert.NoError(t, tmsEngine.UpsertRule(valueRule))
	}
	assert.Len(t, tmsEngine.GetByType(OperandType_ZONE), 3)
}

func TestTmsEngine_TestRule(t *testing.T) {
	/*track8 := tms.Track{
		Metadata: []*tms.TrackMetadata{
			{
				Iridium: &iridium.Iridium{
					Moh: &iridium.MobileOriginatedHeader{
						IMEI: "IOJoijqw",
					},
				},
			},
		},
	}*/
	track8 := tms.Track{
		Targets: []*tms.Target{
			{
				Imei: &google_protobuf1.StringValue{Value: "IOJoijqw"},
			},
		},
	}
	track9 := tms.Track{}

	tmsEngine, err := NewTmsEngine(nil, nil)
	assert.NoError(t, err)
	// test insert
	for _, valueRule := range mapRules {
		assert.NoError(t, tmsEngine.UpsertRule(valueRule))
	}
	acts, err := tmsEngine.CheckRule(track8)
	assert.NoError(t, err)
	assert.Len(t, acts, 2)

	acts, err = tmsEngine.CheckRule(track9)
	assert.NoError(t, err)
	assert.Len(t, acts, 0)
}

func TestCheckRule(t *testing.T) {
	tmsEngine, err := NewTmsEngine(nil, nil)
	assert.NoError(t, err)

	track1 := tms.Track{
		Targets: []*tms.Target{
			{
				Mmsi:     "97",
				Speed:    &google_protobuf1.DoubleValue{Value: 5.0},
				Position: &tms.Point{Latitude: 1.5, Longitude: 1.5},
			},
		},
		RegistryId: "196352",
	}
	track2 := tms.Track{
		Targets: []*tms.Target{
			{
				Mmsi:     "1",
				Speed:    &google_protobuf1.DoubleValue{Value: 6.0},
				Position: &tms.Point{Latitude: 1.5, Longitude: 2.5},
			},
		},
	}
	track3 := tms.Track{
		Targets: []*tms.Target{
			{
				Mmsi:  "97",
				Speed: &google_protobuf1.DoubleValue{Value: 6.0},
			},
		},
	}
	track4 := tms.Track{
		Targets: []*tms.Target{
			{
				Mmsi:   "97",
				Speed:  &google_protobuf1.DoubleValue{Value: 6.0},
				Course: &google_protobuf1.DoubleValue{Value: 20.0},
			},
		},
	}
	track5 := tms.Track{
		Targets: []*tms.Target{
			{
				Mmsi:   "97",
				Speed:  &google_protobuf1.DoubleValue{Value: 6.0},
				Course: &google_protobuf1.DoubleValue{Value: 200000.0},
			},
		},
	}
	track6 := tms.Track{
		Targets: []*tms.Target{
			{
				Mmsi:   "97",
				Speed:  &google_protobuf1.DoubleValue{Value: 60.0},
				Course: &google_protobuf1.DoubleValue{Value: 200000.0},
			},
		},
	}
	track7 := tms.Track{
		Targets: []*tms.Target{
			{
				Imei: &google_protobuf1.StringValue{Value: "IOJoijqw"},
			},
		},
	}
	/*track8 := tms.Track{
		Targets: []*tms.Target{
			{
				Imei: &google_protobuf1.StringValue{"IOJoijqw"},
			},
		},
		Metadata: []*tms.TrackMetadata{
			{
				Iridium: &iridium.Iridium{
					Moh: &iridium.MobileOriginatedHeader{
						IMEI: "IOJoijqw",
					},
				},
			},
		},
	}*/
	track9 := tms.Track{}

	_ = track1
	_ = track2
	_ = track3
	_ = track4
	_ = track5
	_ = track6
	_ = track7
	//_ = track8

	assert.True(t, tmsEngine.checkRuleTree(mapRules["MMSIEqual97RuleAll"], track1))
	assert.False(t, tmsEngine.checkRuleTree(mapRules["MMSIEqual97RuleAll"], track2))
	assert.True(t, tmsEngine.checkRuleTree(mapRules["MMSIEqual97RuleAll"], track3))

	assert.False(t, tmsEngine.checkRuleTree(mapRules["bad second operandType"], track1))
	assert.False(t, tmsEngine.checkRuleTree(mapRules["bad second operandType"], track2))
	assert.False(t, tmsEngine.checkRuleTree(mapRules["bad second operandType"], track3))

	assert.True(t, tmsEngine.checkRuleTree(mapRules["test MMSI==97 && (speed == 5 || speed == 6)"], track1))
	assert.False(t, tmsEngine.checkRuleTree(mapRules["test MMSI==97 && (speed == 5 || speed == 6)"], track2))
	assert.True(t, tmsEngine.checkRuleTree(mapRules["test MMSI==97 && (speed == 5 || speed == 6)"], track3))

	assert.True(t, tmsEngine.checkRuleTree(mapRules["Big depth"], track4))
	assert.True(t, tmsEngine.checkRuleTree(mapRules["Big depth"], track4))
	assert.False(t, tmsEngine.checkRuleTree(mapRules["Big depth"], track3))
	assert.False(t, tmsEngine.checkRuleTree(mapRules["Big depth"], track5))

	assert.True(t, tmsEngine.checkRuleTree(mapRules["speed > 5 && speed <= 10"], track2))
	assert.False(t, tmsEngine.checkRuleTree(mapRules["speed > 5 && speed <= 10"], track6))

	// couldn't check because don't have a mongodb for tests
	//assert.True(t, tmsEngine.checkRuleTree(mapRules["in zone{X:1.0, X1:2.0, Y:1.0, Y1:2.0}"], track1))
	//assert.False(t, tmsEngine.checkRuleTree(mapRules["in zone{X:1.0, X1:2.0, Y:1.0, Y1:2.0}"], track2))
	//assert.True(t, tmsEngine.checkRuleTree(mapRules["in zone{X:1.0, X1:2.0, Y:1.0, Y1:2.0} Any"], track1))
	//assert.False(t, tmsEngine.checkRuleTree(mapRules["in zone{X:1.0, X1:2.0, Y:1.0, Y1:2.0} Any"], track2))

	assert.True(t, tmsEngine.checkRuleTree(mapRules["Target registry id == 196352 && speed == 5"], track1))
	assert.False(t, tmsEngine.checkRuleTree(mapRules["Target registry id == 196352 && speed == 5"], track2))

	assert.True(t, tmsEngine.checkRuleTree(mapRules["Deep structure imei == IOJoijqw"], track7))
	assert.True(t, tmsEngine.checkRuleTree(mapRules["Deep structure imei == IOJoijqw ALL"], track7))
	assert.False(t, tmsEngine.checkRuleTree(mapRules["Bad naming in Deep structure imei == IOJoijq"], track7))

	//assert.True(t, tmsEngine.checkRuleTree(mapRules["MetaData deep"], track8))
	assert.False(t, tmsEngine.checkRuleTree(mapRules["MetaData deep"], track1))
	//assert.True(t, tmsEngine.checkRuleTree(mapRules["MetaData deep icase"], track8))
	assert.False(t, tmsEngine.checkRuleTree(mapRules["MetaData deep icase"], track1))
	//assert.True(t, tmsEngine.checkRuleTree(mapRules["MetaData deep icase Any"], track8))
	assert.False(t, tmsEngine.checkRuleTree(mapRules["MetaData deep icase Any"], track1))

	for _, valRule := range mapRules {
		assert.False(t, tmsEngine.checkRuleTree(valRule, track9))
	}
}
