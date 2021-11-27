from bson.objectid import ObjectId
import datetime

devices = [{
	"id" : "5a7b1c7f77f94b08386da3f6",
	"type" : "omnicom-vms",
	"deviceid" : "111234010030455",
	"networks" : [
		{
			"subscriberid" : "111234010030455",
			"type" : "satellite-data",
			"providerid" : "Iridium",
			"registryid" : ""
		}
	]
}]

zones = [{
    "_id": ObjectId("5a7c594577f94b43223b7f63"),
	"ctime" : datetime.datetime.now(),
	"etime" : datetime.datetime(2020, 1, 1),
	"me" : {
		"create_alert_on_enter" : False,
		"create_alert_on_exit" : False,
		"fill_color" : {
			"a" : 0.25,
			"r" : 255
		},
		"fill_pattern" : "solid",
		"name" : "Zone",
		"poly" : {
			"lines" : [
				{
					"points" : [
						{
							"type" : "Point",
							"coordinates" : [
								104.25993702015602,
								1.152268192395013
							]
						},
						{
							"type" : "Point",
							"coordinates" : [
								103.9862928816391,
								0.8786672905688704
							]
						},
						{
							"type" : "Point",
							"coordinates" : [
								104.53966213952886,
								0.6825739549636154
							]
						},
						{
							"type" : "Point",
							"coordinates" : [
								104.63695783322376,
								1.2039456016606493
							]
						},
						{
							"type" : "Point",
							"coordinates" : [
								104.25993702015602,
								1.152268192395013
							]
						}
					]
				}
			]
		},
		"stroke_color" : {

		}
	}
}
]
