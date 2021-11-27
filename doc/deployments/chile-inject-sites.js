var sites = [{
    "siteid": 1,
    "type": "RCC",
    "name": "RCC1 Iquique",
    "description": "RCC1 Iquique",
    "address": "",
    "country": "",
    "point": null,
    "parentid": "",
    "incidentidprefix": "RCC1",
    "devices": [],
    "capability": { "inputincident": true, "outputincident": true }
}, {
    "siteid": 2,
    "type": "RCC",
    "name": "RCC2 Antofagasta",
    "description": "RCC2 Antofagasta",
    "address": "",
    "country": "",
    "point": null,
    "parentid": "",
    "incidentidprefix": "RCC2",
    "devices": [],
    "capability": { "inputincident": true, "outputincident": true }
}, {
    "siteid": 3,
    "type": "RCC",
    "name": "RCC3 Santiago",
    "description": "RCC3 Santiago",
    "address": "",
    "country": "",
    "point": null,
    "parentid": "",
    "incidentidprefix": "RCC3",
    "devices": [],
    "capability": { "inputincident": true, "outputincident": true }
}, {
    "siteid": 4,
    "type": "RCC",
    "name": "RCC4 Puerto Montt",
    "description": "RCC4 Puerto Montt",
    "address": "",
    "country": "",
    "point": null,
    "connectionstatus": 0,
    "parentid": "",
    "incidentidprefix": "RCC4",
    "devices": [],
    "capability": { "inputincident": true, "outputincident": true }
}, {
    "siteid": 5,
    "type": "RCC",
    "name": "RCC5 Punta Arenas",
    "description": "RCC5 Punta Arenas",
    "address": "",
    "country": "",
    "point": null,
    "connectionstatus": 0,
    "parentid": "",
    "incidentidprefix": "RCC5",
    "devices": [],
    "capability": { "inputincident": true, "outputincident": true }
}, {
    "siteid": 6,
    "type": "RCC",
    "name": "RCC6 Isla de Pascua",
    "description": "RCC6 Isla de Pascua",
    "address": "",
    "country": "",
    "point": null,
    "connectionstatus": 0,
    "parentid": "",
    "incidentidprefix": "RCC6",
    "devices": [],
    "capability": { "inputincident": true, "outputincident": true }
}];

db.sites.remove({});
db.sites.insert(sites);

// It is required that we add an "id" field for incident transfer to work.
// We'll just base it off the auto-generated "_id" value.
db.sites.find().toArray().forEach(element => {
    element.id = element._id.str;
    db.sites.update({"_id": element._id}, element);
});
