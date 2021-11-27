# Server Configuration

## tmsd.conf

## config.js

!!! note
    The following is applicable as of 1.7.7. Previous versions must edit this information directory in the mongo config collection.

For all installations, you will need to create `/etc/trident/config.js` to setup the configuration
that is sent to the client.

The file must have the following format: `var config = {...config property object};`

!!! warning "TODO"
    Document the config file with the new properties for 1.7.7 as the config has drastically changed for this version.

Example config.js:

```js tab="1.7.6 and Earlier"
var config = {
  "brand": {
   "name": "PRISMA RCC",
   "version": "1.7.6",
   "releaseDate": "May 24, 2019"
  },
  "meta": {
   "device": {
    "actions": {
     "CREATE": 0,
     "DELETE": 3,
     "READ": 1,
     "UPDATE": 2
    },
    "network": {
     "type": [
      "iridium"
     ]
    },
    "type": [
     "phone",
     "email",
     "ais",
     "omnicom-vms",
     "omnicom-solar",
     "epirb",
     "sart-radar",
     "sart-ais",
     "elt",
     "plb",
     "mob-ais"
    ]
   },
   "fleet": {
    "actions": {
     "ADD_VESSEL": 4,
     "CREATE": 0,
     "DELETE": 3,
     "READ": 1,
     "REMOVE_VESSEL": 5,
     "UPDATE": 2
    }
   },
   "vessel": {
    "actions": {
     "ADD_DEVICE": 4,
     "CREATE": 0,
     "DELETE": 3,
     "READ": 1,
     "REMOVE_DEVICE": 5,
     "UPDATE": 2
    },
    "type": [
     "ship-fishing",
     "ship-passenger",
     "ship-cargo",
     "ship-tanker",
     "ship-pleasure",
     "ship-supply",
     "ship-utility",
     "ship-research",
     "ship-military"
    ]
   }
  },
  "site": {
   "siteId": 10,
   "type": "MRCC",
   "name": "RCC Kota Kinabalu",
   "description": "",
   "address": "",
   "country": "MY",
   "incidentidprefix": "KK",
   "capability": {
    "inputIncident": true,
    "outputIncident": true
   }
  },
  "service": {
   "map": {
       "base": "https://api.maptiler.com/maps/8d5f5371-20d2-4f4d-93d7-89a0ab809028/{z}/{x}/{y}.png?key=zbnPRwOrABWw5gM6d3qH"
   },
   "ws": {
    "map": "wss://10.133.65.105:8080/ws/v2/view/stream",
    "tms": "wss://10.133.65.105:8080/ws/v2/"
   },
   "tms": {
    "base": "https://10.133.65.105:8080/api/v2",
    "device": "https://10.133.65.105:8080/api/v2/device",
    "fleet": "https://10.133.65.105:8080/api/v2/fleet",
    "track": "https://10.133.65.105:8080/api/v2/track",
    "incident": "https://10.133.65.105:8080/api/v2/incident",
    "communication": "https://10.133.65.105:8080/api/v2/communication",
    "notification": "https://10.133.65.105:8080/api/v2/notification",
    "vessel": "https://10.133.65.105:8080/api/v2/vessel",
    "registry": "https://10.133.65.105:8080/api/v2/registry",
    "rule": "https://10.133.65.105:8080/api/v2/rule",
    "map": "https://10.133.65.105:8080/api/v2/view",
    "zone": "https://10.133.65.105:8080/api/v2/zone",
    "file": "https://10.133.65.105:8080/api/v2/file",
    "pagination": "https://10.133.65.105:8080",
    "swagger": "https://10.133.65.105:8080/api/v2/apidocs.json",
    "activity": "http://10.133.65.105:7077/activity",
    "request": "http://10.133.65.105:7077/request"
   },
   "aaa": {
    "base": "https://10.133.65.105:8181/api/v2",
    "session": "https://10.133.65.105:8181/api/v2/auth/session",
    "user": "https://10.133.65.105:8181/api/v2/auth/user",
    "role": "https://10.133.65.105:8181/api/v2/auth/role",
    "policy": "https://10.133.65.105:8181/api/v2/auth/policy",
    "profile": "https://10.133.65.105:8181/api/v2/auth/profile",
    "audit": "https://10.133.65.105:8181/api/v2/auth/audit",
    "pagination": "https://10.133.65.105:8181",
    "swagger": "https://10.133.65.105:8181/api/v2/auth/apidocs.json"
   },
   "sim": {
    "alert": "http://10.133.65.105:8089/v1/alert",
    "target": "http://10.133.65.105:8089/v1/target",
    "route": "http://10.133.65.105:8089/v1/route"
   }
  },
  "client": {
   "locale": "en-US",
   "distance": "nauticalMiles",
   "shortDistance": "meters",
   "speed": "knots",
   "coordinateFormat": "degreesMinutes",
   "timeZone": "UTC"
  },
  "policy": { "description": "" },
  "spider": false,
  "lat": 5.33,
  "lon": 113.2,
  "zoom": 7
};
```

```js tab="1.7.7 and Later"
var config = {
    "name": "PRISMA RCC",
    "locale": "fr",
    "theme": "dark",
    "units": {
        "distance": "nauticalMiles",
        "shortDistance": "meters",
        "speed": "knots",
        "coordinateFormat": "degreesMinutes"
    },
    "map": {
        "baseUrl": "https://maps.tilehosting.com/styles/positron/{z}/{x}/{y}.png?key=<APIKEY>",
        "center": {
            "lat": 12.2,
            "lon": -113.2,
            "zoom": 7
        }
    }
};
```

## sites.js

!!! note
    The following is applicable as of 1.7.7. Previous versions must edit this information directory in the mongo sites collection.

For installations that require incident forwarding, you must create a sites.js file in `/etc/trident/sites.js`. You can also create the sites config to add the site location to the map as well.

To configure the sites, create `/etc/trident/sites.js` and fill it with the following template, changing the values in `<>` to the actual value for the site.
Since the sites is an array, you can add any number of addition sites by adding another site block that is comma separated from the first site.

`/etc/trident/sites.js`
```
var sites = [{
    "siteid": <any integer larger than or equal to 10>,
    "type": "MRCC",
    "name": "<SITE NAME>",
    "description": "<DESCRIPTION OF SITE>",
    "address": "<ADDRESS>",
    "country": "<COUNTRY CODE>",
    "point": {
        "latitude": <LATITUDE>,
        "longitude": <LONGITUDE>,

    },
    "parentid": "",
    "incidentidprefix": "<2-3 Character Prefix>",
    "devices": [],
    "capability": { "inputincident": true, "outputincident": true }
}];
```

### Properties

* __siteid__: ID of the site that is used to configure tgwad. This must be an integer value from 10 or higher. Usually sites are configured in 10s, so 10, 20, 30, 40 for a group of RCC sites.
* __type__: Must be MRCC or clicking on the site in client that opens the sidebar will not work
* __name__: String name of the site.
* __description__: Description of the site.
* __address__: Address of the site.
* __country__: The two digit country code of the site.
* __incidentprefix__: Prefix that will be prepended to all incident IDs. Usually this is a 2-3 character uppercase combination, but can be more like in Chile where it was RCC1,RCC2, etc...
* __point__: The lat/long of the site. This is required for the site to show on the map.
* __parentid__: Leave as empty string.
* __devices__: Leave as empty array.
* __capability__: Leave as `{ "inputincident": true, "outputincident": true }` or incident transfer will not work.

Here is an example with multiple sites from Malaysia:

```
var sites = [{
    "siteid": 10,
    "type": "MRCC",
    "name": "RCC Kota Kinabalu",
    "description": "RCC Kota Kinabalu",
    "address": "",
    "country": "MY",
    "point": {
        "latitude": 5.91964,
        "longitude": 116.052696,

    },
    "parentid": "",
    "incidentidprefix": "KK",
    "devices": [],
    "capability": { "inputincident": true, "outputincident": true }
}, {
    "siteid": 20,
    "type": "MRCC",
    "name": "RCC Kuala Lumpur",
    "description": "RCC Kuala Lumpur",
    "address": "",
    "country": "MY",
    "point": {
        "latitude": 2.758470,
        "longitude": 101.708092,

    },
    "parentid": "",
    "incidentidprefix": "KL",
    "devices": [],
    "capability": { "inputincident": true, "outputincident": true }
}];
```

!!! note
    The sites.js file will be processed by the tms application everytime tms is restarted. So if you make any changes, you must restart the application for the changes to take effect.
