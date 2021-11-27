from bson import ObjectId

oid1 = ObjectId()
oid2 = ObjectId()
site1 = 
data = [{
    "_id": oid1,
    "id": str(oid1),
    "siteid": 27,
    "type": "C2",
    "name": "Integration",
    "description": "McMurdo PRISMA integration Server.",
    "country": "USA",
    "point": {
        "latitude": 39,
        "longitude": -76
    },
    "capability": {
        "inputIncident": True,
        "outputIncident": True
    }
},{
    "_id": oid2,
    "id": str(oid2),
    "siteid": 10,
    "type": "C2",
    "name": "Integration",
    "description": "McMurdo PRISMA integration Server.",
    "country": "USA",
    "point": {
        "latitude": 39,
        "longitude": -76
    },
    "capability": {
        "inputIncident": True,
        "outputIncident": True
    }
}]
