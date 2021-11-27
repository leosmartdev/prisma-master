// Retrieve the config file from /etc/trident
load("/etc/trident/config.js")

// Load the trident db
let ldb = db.getSiblingDB('trident');

// New config ID
id = ObjectId();
config._id = id;
config.id = id.str;

// Load the config collection
let configcol = ldb.getCollection("config");

// Clear the old config
configcol.remove({});

// Insert the new config
printjson(configcol.insert(config));
