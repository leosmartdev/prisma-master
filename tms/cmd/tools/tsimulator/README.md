# Introduction
tsimulator simulates objects. This program provides information about objects and stations.
    Each stations can see vessels, radars and etc and send information to clients.
The clients have to connect to ports which are provided in a file of stations and
receive several messages which are explained in links below:
- https://www.navcen.uscg.gov/?pageName=AISMessagesA
- https://www.navcen.uscg.gov/?pageName=AISMessagesAStatic
In Additional, the simulator provides web REST service to manage the objects.

## Cmd parameters
Use: tsimulator -help for reading information about use

## File structures
tsimulator has two files:
- data for objects
- with stations

The files must contain encoding data by json

#### A structure of a sea objects's file
Overall the file has two sections.
- The first one contains a configure for sarsat beacons
- The second one contains objects

##### The configure times for objects
```json
{
  "time_config": {
    "sleep_get_info_position": 2,
    "sleep_get_info_static_information": 2,
    "sleep_get_tracked_target_message": 2,
    "sleep_move_vessels": 2
  }
}
```
- sleep_get_info_position - in time will be sent a message for position
- sleep_get_info_static_information - in time will be sent a message for static information
- sleep_get_tracked_target_message - in time will be sent TTM message
- sleep_move_vessels - in time will be moved objects. If it is zero it will assign to sleep_get_info_position
**It can be assign to an object and the object will have own times for actions**

##### The configure for objects
```json
{
    "objects": [
      {
        "device": "ais"
      }
    ]
}
```
- Pos: It has an array.
    - Latitude: an object go to and from
    - Longitude: an object go to and from
    - Speed: uses to manage time to reach the next position
    - Damage: ENUM ("low", "medium", "high") the level of damage if this object will be intersected with another object
    - ArrivalTime: it will be used to compute the speed to reach this point from the previous point.
        - It must contain seconds time since the starting program
        - if speed has been already pointed then it will be skipped
        - if the first position contains this field and the speed is 0, then this object will be ignored until
         this time
         ***it can provide problems for some objects like omnicom. They have limitations for their speed***
- Device: Type of an object. It can be:
    - ais
    - radar
    - omnicom
    - sarsat
    - spidertracks
1. Fields for an AIS object
    - MMSI: A Maritime Mobile Service Identity is a series of nine digits which are sent in digital form over a radio frequency channel in order to uniquely identify ship stations, ship earth stations, coast stations, coast earth stations, and group calls
    - Destination: Where a vessel should arrive. For it, UN/LOCODE and ERI terminal codes should be used
    - Name: A name of a vessel, also it is used as callSign
    - ETA: Estimated time of arrival. It uses a format "MMDDHHMM" 01021530 - 02 Jan 15:30:00
    - Type: A type of a vessel
    - Speed: A speed of a vessel. It can apply a range 1...123.0. If it is zero the one will apply a random value. A unit is knots
    - Pos: see Pos above
    >If the name didn't point the name will get a random value
1. Fields for radars
    - Speed: A speed of an object. It can apply a range 1...123.0. If it is zero the one will apply a random value. A unit is knots
    - Pos: see Pos
        >The position should be passed to tnoid if the station can see radars,
        cause radars use a relative position
1. Fields foromnicom beacons and spidertracks
    - Imei: A number of an omnicom beacon
    - Speed: A speed of an object. It can apply a range 1...123.0. If it is zero the one will apply a random value. A unit is knots
    - Pos: It has an array.
        - Latitude: an object go to and from
        - Longitude: an object go to and from
        - Speed: this speed an object uses for a way to a next position
        >The position should be passed to tnoid if the station can see radars,
        cause radars use a relative position
    >If it doesn't have imei then imei will get a random value
1. Fields for a sarsat beacon
    - beacon_id: it is used by sarsat beacons for identification
    - unlocated: if true, issue a target without a position
    - located: if true, issue a target with an indeterminate position (two points)
    - Pos: see Pos
    - Elemental: It is an array.
            - dopplerA: It is an structure for position type A
            - dopplerA: It is an structure for position type B

#### A structure of a stations's file
```json
{
  "device": "radar",
  "latitude": 1.171931,
  "longitude": 103.131706,
  "radius": 0.4,
  "addr": ":9000"
}
```
- Device: What the devices station can see.
A station cannot observe on omnicom beacons, sarsat beacons. They use other channels for communicating.
- Latitude: Where a station is
- Longitude: Where a station is
- Radius: A radius for watching vessels. A unit is Nm
- Addr: <interface>:<port> for listening to connections

#### Rest Web service
+ GET /v1/get/: get a js object of sea objects. It contains a field an id and keys as id - it will use for requests below
+ POST /v1/target/id/{target-id}: update a target with {target-id}. Body should be as json of an object above
+ POST /v1/target/: create a target. Body should be as json of an object above. You can pass your own id for a new object.
+ DELETE /v1/target/id/{target-id}: delete a target with {target-id}.

- POST /v1/alert/type/{type-alert}/target-id/{target-id}: create an alert of a type {type-alert} for an object with {target-id}
- DELETE /v1/alert/type/{type-alert}/target-id/{target-id}: delete an alert of a type {type-alert} for an object with {target-id}
The type can be the one from follow below:
```
const (
	PU               = iota
	PD
	BA
	IA
	NPF
	JBDA
	LMC
	DA
	AA
	TM
	LastTypeAlerting
)
```
The types can be in lower or up cases

- POST /route/id/{route-id}/target-id/{target-id}: update the {route-id}'th route for an object with {target-id}
- POST /route/target-id/{target-id}: update the whole route for an object with {target-id}
- DELETE /route/target-id/{target-id}: delete the whole route for an object

- POST /aff/feed: to get data about spidertracks

#### Useful links
- https://en.wikipedia.org/wiki/Maritime_Mobile_Service_Identity
- http://catb.org/gpsd/AIVDM.html#_types_1_2_and_3_position_report_class_a
- http://catb.org/gpsd/AIVDM.html#_type_5_static_and_voyage_related_data
- http://www.catb.org/gpsd/NMEA.html#_ttm_tracked_target_message