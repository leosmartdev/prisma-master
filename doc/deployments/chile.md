# Chile RCCs

This document describes the Chile Sites including site numbers, names, config.json configurations, and IP addresses.

## Sites

  * RCC 1 Iquique
  * RCC 2 Antofagasta
  * RCC 3 Santiago
  * RCC 4 Puerto Montt
  * RCC 5 Punta Arenas
  * RCC 6 Isla de Pascua

Each site has two servers, .101 and .102 for primary and secondary failover system. They each have a single windows RCC client reachable through TeamViewer.


### Server IP Addresses

All IPs start with 10.181 and end with .101 or .102 (primary and secondary respectively). The third number starts at 99 for RCC 1 and goes to 104 for RCC 6.

Every server has a name as well, they all begin with C2RCC1 then continue with a letter from A to L.

  * RCC 1
    - 10.181.99.101  C2RCC1A
    - 10.181.99.102  C2RCC1B
  * RCC 2
    - 10.181.100.101 C2RCC1C
    - 10.181.100.102 C2RCC1D
  * RCC 3
    - 10.181.101.101 C2RCC1E
    - 10.181.101.102 C2RCC1F
  * RCC 4
    - 10.181.102.101 C2RCC1G
    - 10.181.102.102 C2RCC1H
  * RCC 5
    - 10.181.103.101 C2RCC1I
    - 10.181.103.102 C2RCC1J
  * RCC 6
    - 10.181.104.101 C2RCC1K
    - 10.181.105.102 C2RCC1L

## Incident Forwarding Configuration

Every site has a unique site number that matches their RCC number (1-6) and a site name that is RCCX City. All incidents are opened with a site prefix of RCCX where X is the site number.

tgwad must be configured to be able to reach all the sites. To do this, we must pass a list of sites using a flag for each site:

```
tgwad -site <name>,<number>,gw,tcp:<idaddr>:<tgwad port> -site <name>....
```

tgwad also needs to know what site it is a part of, so you must pass `--num` and `--name` flags to tgwad as well.

```
tgwad --num 1 --name "RCC1 Iquique"
```

### RCC 1 Iquique

Site Information
```
tgwad --num 1 --name "RCC1 Iquique"
```

Primary
```
--site "RCC1 Iquique",1,gw,tcp:10.181.99.101:31228
```

Secondary
```
--site "RCC1 Iquique",1,gw,tcp:10.181.99.102:31228
```

### RCC 2 Antofagasta

Site Information
```
tgwad --num 2 --name "RCC2 Antofagasta"
```

Primary
```
--site "RCC2 Antofagasta",2,gw,tcp:10.181.100.101:31228
```

Secondary
```
--site "RCC2 Antofagasta",2,gw,tcp:10.181.100.102:31228
```

### RCC 3 Santiago

Site Information
```
tgwad --num 3 --name "RCC3 Santiago"
```

Primary
```
--site "RCC3 Santiago",3,gw,tcp:10.181.101.101:31228
```

Secondary
```
--site "RCC3 Santiago",3,gw,tcp:10.181.101.102:31228
```

### RCC 4 Puerto Montt

Site Information
```
tgwad --num 4 --name "RCC4 Puerto Montt"
```

Primary
```
--site "RCC4 Puerto Montt",4,gw,tcp:10.181.102.101:31228
```

Secondary
```
--site "RCC4 Puerto Montt",4,gw,tcp:10.181.102.102:31228
```

### RCC 5 Punta Arenas

Site Information
```
tgwad --num 5 --name "RCC5 Punta Arenas"
```

Primary
```
--site "RCC5 Punta Arenas",5,gw,tcp:10.181.103.101:31228
```

Secondary
```
--site "RCC5 Punta Arenas",5,gw,tcp:10.181.103.102:31228
```

### RCC 6 Isla de Pascua

Site Information
```
tgwad --num 6 --name "RCC6 Isla de Pascua"
```

Primary
```
--site "RCC6 Isla de Pascua",6,gw,tcp:10.181.104.101:31228
```

Secondary
```ßß
--site "RCC6 Isla de Pascua",6,gw,tcp:10.181.104.102:31228
```

## Sites in Database

Required for the UI to have a list of sites, you will also need to insert the list of Site objects into the database.

```json
{
  "sites": [{
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
  }]
}
```

Script for injecting the above sites into a mongo db can be found in
[`doc/deployments/chile-inject-sites.js](./chile-inject-sites.js). To run:

```
mongo localhost/trident chile-inject-sites.js
```
