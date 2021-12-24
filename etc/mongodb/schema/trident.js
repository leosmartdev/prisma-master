let ldb = db.getSiblingDB('trident');

ldb.createCollection("config");

ldb.createCollection("sites");
ldb.sites.createIndex({"siteid": 1}, {
    "v": 2,
    "unique": true,
    "name": "siteIdUnique",
    "ns": "trident.sites"
});
ldb.sites.createIndex({"name": 1}, {
    "v": 2,
    "unique": true,
    "name": "nameUnique",
    "ns": "trident.sites"
});

ldb.createCollection("remoteSites");

ldb.createCollection("fleets");
ldb.fleets.createIndex({"name": 1}, {
    name: "fleetNameUnique",
    unique: true
});
ldb.fleets.createIndex({"person.name": "text", "name": "text"}, {"name": "text-search"})


ldb.createCollection("incidents");
ldb.incidents.createIndex({"me.incidentId": 1}, {
    name: "incidentIdUnique",
    unique: true
});

ldb.createCollection("notes");

ldb.createCollection("markers");
ldb.createCollection("markerImages");

ldb.createCollection("icons");
ldb.createCollection("iconImages");

ldb.createCollection("activity");
ldb.activity.createIndex({time: 1}, {
    name: "time",
    background: true
});
ldb.activity.createIndex({request_id: 1}, {
    name: "request_id",
    background: true
});
ldb.activity.createIndex({activity_id: 1, time: 1}, {
    name: "activity_id_time",
    unique: true
});

ldb.createCollection("request");
ldb.request.createIndex({time: 1}, {
    name: "time",
    background: true
});
ldb.request.createIndex({request_id: 1, time: 1}, {
    name: "request_id_time",
    unique: true
});

ldb.createCollection("notices");
ldb.notices.createIndex({
    name: "ctime",
    ctime: 1,
    background: true
});
ldb.notices.createIndex({
    name: "track_id",
    track_id: 1,
    background: true
});

ldb.createCollection("referenceSequence");
ldb.createCollection("transmissions");
ldb.createCollection("multicasts");

ldb.createCollection("vessels");
ldb.vessels.createIndex({"type": "text", "name": "text"}, {"name": "text-search"});
ldb.vessels.createIndex({"devices.deviceid": 1, "devices.type": 1}, {
    "v": 2,
    "unique": true,
    "name": "deviceid-type",
    "ns": "trident.vessels",
    "partialFilterExpression": {"devices.deviceid": {"$exists": true}}
});
ldb.vessels.createIndex({"devices.networks.subscriberid": 1, "devices.networks.providerid": 1}, {
    "v": 2,
    "unique": true,
    "name": "network-subscriberid-providerid",
    "ns": "trident.vessels",
    "partialFilterExpression": {"devices.networks": {"$exists": true}}
});

ldb.createCollection("devices");
ldb.devices.createIndex({"deviceid": 1, "type": 1}, {unique: true, name: "deviceid-type"});
ldb.devices.createIndex({"networks.subscriberid": 1, "networks.providerid": 1}, {
    "v": 2,
    "unique": true,
    "name": "network-subscriberid-providerid",
    "ns": "trident.devices",
    "partialFilterExpression": {"networks": {"$exists": true}}
});

ldb.createCollection("registry");
ldb.registry.createIndex({"me.registry_id": 1}, {unique: true, name: "registry_id"});

ldb.createCollection("tracks");
ldb.tracks.createIndex({time: 1}, {
    name: "time",
    background: true
});
ldb.tracks.createIndex({update_time: 1}, {
    name: "update_time",
    background: true
});
ldb.tracks.createIndex({track_id: 1, time: 1, update_time: 1}, {
    name: "track_id_time",
    unique: true
});
ldb.tracks.createIndex({track_id: 1, time: -1}, {
    name: "track_id_time_desc",
    background: true
});

ldb.tracks.createIndex({registry_id: 1},{
    name: "registry_id",
    background: true
});

ldb.createCollection("zones");

ldb.createCollection("sit915");

ldb.createCollection("mapconfig");
// ldb.mapconfig.createIndex({"me.key": 1}, {
//     name: "mapconfigKey",
//     unique: true
// });

ldb.createCollection("filtertracks");
// ldb.filtertracks.createIndex({"me.type": 1}, {
//     name: "tracktype",
//     unique: true
// });