package tmsg

const (
	TMSG_UNKNOWN_SITE uint32 = 0x00
	TMSG_LOCAL_SITE   uint32 = 0x01
	TMSG_HQ_SITE      uint32 = 0x02

	APP_ID_UNKNOWN          uint32 = 0x00
	APP_ID_TGWAD            uint32 = 0x02
	APP_ID_TREPORTD         uint32 = 0x18
	APP_ID_TWATCH           uint32 = 0x1E
	APP_ID_TDRUNKCPT        uint32 = 0x2E
	APP_ID_TCLIENTD         uint32 = 0x30
	APP_ID_TDATABASED       uint32 = 0x31
	APP_ID_TPING            uint32 = 0x32
	APP_ID_TANALYZED        uint32 = 0x33
	APP_ID_TWEBD            uint32 = 0x34
	APP_ID_TMSD             uint32 = 0x35
	APP_ID_TNOID            uint32 = 0x36
	APP_ID_TFLEETD          uint32 = 0x37
	APP_ID_TORBCOMMD        uint32 = 0x38
	APP_ID_TSIMULATORD      uint32 = 0x40
	APP_ID_OMNICOMSIMLATORD uint32 = 0x41
	APP_ID_TNAFEXPORTD      uint32 = 0x42
	APP_ID_TAAAD            uint32 = 0x43
	APP_ID_TMCCD            uint32 = 0x44
	APP_ID_TLOGGERNMEA      uint32 = 0x45
	APP_ID_TMCCRD           uint32 = 0x46
	APP_ID_CONSUL           uint32 = 0x47
	APP_ID_TSPIDERD         uint32 = 0x48
	APP_ID_TADSBD           uint32 = 0x49
	APP_ID_TVTSD            uint32 = 0x4a

	ENTRY_ID_UNKNOWN uint32 = 0x00
)

var (
	AppIdNames = map[uint32]string{
		APP_ID_UNKNOWN:          "AppUnknown",
		APP_ID_TGWAD:            "tgwad",
		APP_ID_TREPORTD:         "treportd",
		APP_ID_TWATCH:           "twatch",
		APP_ID_TDRUNKCPT:        "tdrunkcpt",
		APP_ID_TCLIENTD:         "tclientd",
		APP_ID_TDATABASED:       "tdatabased",
		APP_ID_TPING:            "tping",
		APP_ID_TWEBD:            "twebd",
		APP_ID_TMSD:             "tmsd",
		APP_ID_TNOID:            "tnoid",
		APP_ID_TFLEETD:          "tfleetd",
		APP_ID_TORBCOMMD:        "torbcommd",
		APP_ID_OMNICOMSIMLATORD: "omn-pos-gen",
		APP_ID_TNAFEXPORTD:      "tnafexportd",
		APP_ID_TAAAD:            "tauthd",
		APP_ID_TMCCD:            "tmccd",
		APP_ID_CONSUL:           "consul",
		APP_ID_TMCCRD:           "tmccrd",
		APP_ID_TSPIDERD:         "tspiderd",
	}
)
