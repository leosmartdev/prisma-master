
//load("/etc/trident/sites.js")
sites = [
    {
        "siteid" : 1, 
        "type" : "MRCC", 
        "name" : "Vagrant", 
        "description" : "This site is local site for dev enviroment", 
        "address" : "localhost:31228", 
        "country" : "Null Island", 
        "point" : { 
            "latitude" : 0, 
            "longitude" : 0, 
            "altitude" : 0 
        }, 
        "connectionstatus" : 0, 
        "parentid" : "", 
        "incidentidprefix" : "DEV", 
        "devices" : [ ], 
        "capability" : { 
            "inputincident" : true, 
            "outputincident" : true 
        } 
    }
];

let ldb = db.getSiblingDB('trident');

for (i = 0; i < sites.length; i++) { 
  id = ObjectId();
  sites[i]._id = id;
  sites[i].id = id.str;
  let sitescol = ldb.getCollection("sites");
  printjson(sitescol.update({"_id": id},sites[i],{upsert: true}));
}

