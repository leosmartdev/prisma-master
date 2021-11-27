import config
import json
import unittest
from utilities import mongo
from utilities.session import authenticate


class CreateFleet(unittest.TestCase):

    @mongo.clean_collections_after("fleets", "vessels")
    @authenticate('fleetmanager', 'one')
    def test_create_fleet_with_vessel_with_valid_input_returns_success(self, session):
        fleet_name = 'fleet1'
        rsp = session.post('fleet', {
            'name':fleet_name,
            'description':''
        })
        self.assertEqual(201, rsp.status_code, 'I should get a 201 response code')
        body = rsp.json()
        fleet_id = body['id']
        # create a vessel that belong to the fleet
        vessel_name = 'vessel1'
        rsp = session.post('vessel', {
            'name':vessel_name,
            'fleet':{
                'id': fleet_id,
            }
        })
        self.assertEqual(201, rsp.status_code, 'I should get a 201 response code')
        body = rsp.json()
        vessel_id = body['id']
        # get the fleet
        rsp = session.get('fleet/' + fleet_id)
        body = rsp.json()
        self.assertDictEqual(body, {
            'id':fleet_id,
            'name':fleet_name,
            'vessels':[
                {
                    'id':vessel_id,
                    'name':vessel_name,
                    'devices':[],
                    'crew':[],
                    'fleet':{
                        'id':fleet_id
                    }
                }
            ]
        })

    @mongo.clean_collections_after("fleets", "vessels", "devices")
    @authenticate('fleetmanager', 'one')
    def test_with_valid_input_returns_success(self, session):
        response = session.post('fleet', {
            "name": "gold",
            "person": {
                "name": "Ad"
            },
            "vessels": [{
                "name": "zuby",
                "devices": [
                    {
                        "type": "omnicom-vms",
                        "deviceId": "900000234234",
                        "networks": [
                            {
                                "subscriberId": "356938035643809",
                                "type": "satellite-data",
                                "providerId": "Iridium"
                            }
                        ]
                    }
                ],
                "crew": [
                    {
                        "name": "Adalberto Zetta"
                    },
                    {
                        "name": "Adrian Yuri"
                    }
                ]
            }]
        })
        self.assertEqual(201, response.status_code, str(response.json()))
        fleet = response.json()
        self.assertIn('id', fleet)
        self.assertEqual('gold', fleet['name'])
        self.assertIn('id', fleet['vessels'][0])
        response = session.get('vessel/' + fleet['vessels'][0]['id'])
        self.assertEqual(200, response.status_code, str(response.json()))
        vessel = response.json()
        self.assertEqual('zuby', vessel['name'], str(vessel))

    @mongo.clean_collections_after("fleets", "vessels", "devices")
    @authenticate('fleetmanager', 'one')
    def test_create_vessel_with_device_and_crew(self, session):
        response = session.post('vessel', {
            "name": "zuby",
            "devices": [
                {
                    "type": "omnicom-vms",
                    "deviceId": "900000234234",
                    "networks": [
                        {
                            "subscriberId": "356938035643809",
                            "type": "satellite-data",
                            "providerId": "Iridium"
                        }
                    ]
                }
            ],
            "crew": [
                {
                    "name": "Adalberto Zetta"
                },
                {
                    "name": "Adrian Yuri"
                }
            ]
        })
        self.assertEqual(201, response.status_code, str(response.json()))
        vessel = response.json()
        vessel_id = vessel['id']
        response = session.post('fleet', {
            "name": "gold",
            "person": {
                "name": "Ad"
            },
            "vessels": [vessel]
        })
        self.assertEqual(201, response.status_code, str(response.json()))
        fleet = response.json()
        self.assertIn('id', fleet)
        self.assertEqual('gold', fleet['name'])
        self.assertIn('id', fleet['vessels'][0])
        self.assertIn('fleet', fleet['vessels'][0])
        # check vessel endpoint to make sure in sync
        response = session.get('vessel/' + vessel_id)
        self.assertEqual(200, response.status_code, str(response.json()))
        vessel = response.json()
        self.assertIn('id', vessel, str(vessel))
        self.assertEqual('zuby', vessel['name'], str(vessel))
        self.assertIn('fleet', vessel, str(vessel))
        self.assertEqual(fleet['id'], vessel['fleet']['id'], str(vessel))

    @mongo.clean_collections_after("fleets", "vessels", "devices")
    @authenticate('fleetmanager', 'one')
    def test_with_create_vessel_create_fleet_double_vessel(self, session):
        response = session.post('vessel', {
            "name": "zuby",
            "devices": [
                {
                    "type": "omnicom-vms",
                    "deviceId": "900000234234",
                    "networks": [
                        {
                            "subscriberId": "356938035643809",
                            "type": "satellite-data",
                            "providerId": "Iridium"
                        }
                    ]
                }
            ],
            "crew": [
                {
                    "name": "Adalberto Zetta"
                },
                {
                    "name": "Adrian Yuri"
                }
            ]
        })
        self.assertEqual(201, response.status_code, str(response.json()))
        vessel = response.json()
        vessel_id = vessel['id']
        response = session.post('fleet', {
            "name": "gold",
            "person": {
                "name": "Ad"
            },
            "vessels": [vessel,vessel]
        })
        self.assertEqual(201, response.status_code, str(response.json()))
        fleet = response.json()
        self.assertIn('id', fleet)
        self.assertEqual('gold', fleet['name'])
        self.assertIn('id', fleet['vessels'][0])
        response = session.get('vessel/' + vessel_id)
        self.assertEqual(200, response.status_code, str(response.json()))
        vessel = response.json()
        self.assertEqual('zuby', vessel['name'], str(vessel))

    @mongo.clean_collections_after('fleets')
    @authenticate('fleetmanager', 'one')
    def test_list_all_fleets(self, session):
        fleetCollection = mongo.get_trident_collection("fleets")
        count = fleetCollection.count()
        self.assertIsNone(fleetCollection.find_one({'name': 'gould'}))
        response = session.post('fleet', {
            'name': 'gould'
        })
        self.assertIsNotNone(fleetCollection.find_one({'name': 'gould'}))

    @mongo.clean_collections_after("fleets", "vessels", "devices")
    @authenticate('fleetmanager', 'one')
    def test_with_delete_endpoint_returns_success(self, session):
        response = session.post('fleet', {
            "name": "gold",
            "person": {
                "name": "Ad"
            },
            "vessels": [{
                "name": "zuby",
                "devices": [
                    {
                        "type": "omnicom-vms",
                        "deviceId": "900000234234",
                        "networks": [
                            {
                                "subscriberId": "356938035643809",
                                "type": "satellite-data",
                                "providerId": "Iridium"
                            }
                        ]
                    }
                ],
                "crew": [
                    {
                        "name": "Adalberto Zetta"
                    },
                    {
                        "name": "Adrian Yuri"
                    }
                ]
            }]
        })
        self.assertEqual(201, response.status_code)
        fleet = response.json()
        vessel = fleet['vessels'][0]
        response = session.delete('fleet/' + fleet['id'])
        self.assertEqual(204, response.status_code)
        response = session.get('vessel/' + vessel['id'])
        self.assertEqual(200, response.status_code, str(response.json()))
        vessel = response.json()
        self.assertEqual('zuby', vessel['name'], str(vessel))
        self.assertEqual({}, vessel['fleet'], str(vessel))

    @mongo.clean_collections_after("fleets")
    @authenticate('admin', 'admin')
    def test_with_empty_string_name_returns_bad_request(self, session):
        response = session.post('fleet', {
            'name': ''
        })
        self.assertEqual(400, response.status_code)
        body = response.json()
        self.assertEqual('name', body[0]['property'])
        self.assertEqual('Required', body[0]['rule'])
        self.assertEqual('Required non-empty property', body[0]['message'])

    @authenticate('admin', 'admin')
    def test_minLength(self, session):
        pass

    @mongo.clean_collections_after("fleets")
    @authenticate('usermanager', 'one')
    def test_create_a_fleet_as_a_user_manager(self, session):
        rsp = session.post('fleet', {
            "name": "edgecase1",
            "person": {"name": "Adrian"}
        })
        self.assertEqual(403, rsp.status_code)

    @mongo.clean_collections_after("fleets")
    @authenticate('incident', 'manager')
    def test_create_a_fleet_as_a_incident_manager(self, session):
        rsp = session.post('fleet', {
            "name": "edgecase3",
            "person": {"name": "Ad"}
        })
        self.assertEqual(403, rsp.status_code)


if __name__ == '__main__':
    unittest.main()
