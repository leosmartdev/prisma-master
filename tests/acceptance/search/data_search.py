mmsi_only = {
    "me": {
        "registry_id": "1",
        "keywords": ["636011907"],
        "label": "636011907",
        "target_fields": [
            {
                "name": "mmsi",
                "value": "636011907"
            }
        ],
    }
}
wavemaster3 = {
    "me": {
        "registry_id": "2",
        "keywords": ["WAVEMASTER 3", "563002830", "9257840"],
        "label": "WAVEMASTER 3",
        "target_fields": [
            {
                "name": "mmsi",
                "value": "563002830"
            },
            {
                "name": "label",
                "value": "WAVEMASTER 3"
            },
            {
                "name": "imo",
                "value": "9257840"
            }
        ],
    }
}
wavemaster5 = {
    "me": {
        "registry_id": "3",
        "keywords": ["WAVEMASTER 5", "563002840", "9257888"],
        "label": "WAVEMASTER 5",
        "target_fields": [
            {
                "name": "mmsi",
                "value": "563002840"
            },
            {
                "name": "label",
                "value": "WAVEMASTER 5"
            },
            {
                "name": "imo",
                "value": "9257888"
            }
        ],
    }
}
fleet = [{
	"id" : "5a5df1d577f94b46842d8ade",
	"type" : "",
	"name" : "gold",
	"person" : {
		"id" : "",
		"userid" : "",
		"name" : "A B",
		"address" : "",
		"devices" : [ ]
	},
	"description" : "",
	"vessels" : [ ]
}]
vessels = [{
	"id" : "5a5df66077f94b5466787de6",
	"type" : "",
	"name" : "zuby",
	"devices" : [
		{
			"id" : "5a348a18f6ffcb32fa919e1f",
			"type" : "omnicom-vms",
			"deviceid" : "900000234234",
			"networks" : [
				{
					"subscriberid" : "356938035643809",
					"type" : "satellite-data",
					"providerid" : "Iridium",
					"trackid" : ""
				},
				{
					"subscriberid" : "856938321643802",
					"type" : "cellular-data",
					"providerid" : "Verizon",
					"trackid" : ""
				}
			]
		}
	],
	"crew" : [
		{
			"id" : "5a392446e1b2f9c0bab3b8b4",
			"userid" : "",
			"name" : "Adalberto Zetta",
			"address" : "",
			"devices" : [ ]
		},
		{
			"id" : "5a392446e1b2f9c0bab3b8b5",
			"userid" : "",
			"name" : "Adrian Yuri",
			"address" : "",
			"devices" : [ ]
		}
	],
}]

data = [
    mmsi_only,
    wavemaster3,
    wavemaster5,
]
