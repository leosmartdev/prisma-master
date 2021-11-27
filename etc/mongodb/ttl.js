let ldb = db.getSiblingDB('trident');

// setting up TTL indexes for tracks, activity and notices
let tracks = ldb.getCollection("tracks");
let activity = ldb.getCollection("activity");
let notices  = ldb.getCollection("notices");

// createTtlIndex takes collection, index name, and duration to create a ttl index.
var createTtlIndex = function(collection, indexName, duration){

    const ikey = {
        [indexName]: 1
    };

    var result = collection.createIndex(ikey, {
        name: indexName,
        background: true,
        expireAfterSeconds: duration
    });

   if (result.errmsg){
      collection.dropIndex(indexName);
      result = collection.createIndex(ikey, {
        name: indexName,
        background: true,
        expireAfterSeconds: duration
    });
   }

   printjson(result);
};

// calling createTtlIndex and passing collection, index name, and expiration time.
printjson("Update TTL indexes for tracks collection ...");
createTtlIndex(tracks,"time",60*60*24*10);
printjson("Update TTL indexes for activity collection ...");
createTtlIndex(activity,"time",60*60*24*10);
printjson("Update TTL indexes for notices collection ...");
createTtlIndex(notices,"etime",60*60*24*10);




