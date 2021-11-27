import config
import json
import unittest
from fleets.data_vessel import data
from utilities import mongo
from utilities.session import authenticate


class UpdateFleet(unittest.TestCase):

    @mongo.clean_collections_after("fleets")
    @mongo.insert_collection_data('fleets', data)
    @authenticate('fleetmanager', 'one')
    def test_update_fleet_without_appending_fleetId(self, session):
        response = session.put('fleet', {
            'name': 'BIGGIE',
            'person': {
                'name': 'Ashton Kutcher'
            }
        })
        self.assertEqual(405, response.status_code)

    @mongo.clean_collections_after("fleets")
    @mongo.insert_collection_data('fleets', data)
    @authenticate('fleetmanager', 'one')
    def test_update_fleet_with_valid_input(self, session):
        response = session.get('fleet')
        self.assertEqual(200, response.status_code, 'I should get a 200 response code')
        body = response.json()
        self.assertIn('name', body[0])
        # self.assertEqual("WAVEMASTER 5", body[0]["label"])
        # response = session.put('fleet' + '', {
        #    'name': 'BIGGIE',
        #    'person': {
        # 'name': 'Ashton Kutcher'
        #    }
        # })
        # self.assertEqual(200, response.status_code)
        # fleet_id = response.json()[0]['id']
        # fleet = response.json()
        # self.assertIn('id', fleet)

    @mongo.clean_collections_after("fleets", "vessels", "devices")
    @authenticate('fleetmanager', 'one')
    def test_add_vessel_to_fleet(self, session):
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
        self.assertEqual(201, response.status_code, response.text)
        vessel = response.json()
        vessel_id = vessel['id']
        response = session.post('fleet', {
            "name": "gold",
            "person": {
                "name": "Ad"
            },
            "vessels": []
        })
        self.assertEqual(201, response.status_code, response.text)
        fleet = response.json()
        fleet_id = fleet['id']
        self.assertIn('id', fleet)
        response = session.put('fleet/' + fleet_id + '/vessel/' + vessel_id, None)
        self.assertEqual(200, response.status_code)
        # check vessel endpoint to make sure in sync
        response = session.get('vessel/' + vessel_id)
        self.assertEqual(200, response.status_code, response.text)
        vessel = response.json()
        self.assertIn('id', vessel, str(vessel))
        self.assertIn('fleet', vessel, str(vessel))
        self.assertEqual(fleet['id'], vessel['fleet']['id'], str(vessel))
        # check fleet endpoint to make sure in sync
        response = session.get('fleet/' + fleet_id)
        self.assertEqual(200, response.status_code, response.text)
        fleet = response.json()
        self.assertIn('id', fleet)
        self.assertEqual('gold', fleet['name'])
        self.assertIn('id', fleet['vessels'][0])
        self.assertIn('fleet', fleet['vessels'][0], str(fleet))
        self.assertEqual(fleet['id'], fleet['vessels'][0]['fleet']['id'], str(fleet))

    @mongo.clean_collections_after("fleets", "vessels", "devices")
    @authenticate('fleetmanager', 'one')
    def test_add_vessel_to_fleet_update_fleet_name(self, session):
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
        self.assertEqual(201, response.status_code, response.text)
        vessel = response.json()
        vessel_id = vessel['id']
        response = session.post('fleet', {
            "name": "gold",
            "person": {
                "name": "Ad"
            },
            "vessels": []
        })
        self.assertEqual(201, response.status_code, response.text)
        fleet = response.json()
        fleet_id = fleet['id']
        self.assertIn('id', fleet)
        # add vessel to fleet
        response = session.put('fleet/' + fleet_id + '/vessel/' + vessel_id, None)
        self.assertEqual(200, response.status_code)
        fleet = response.json()
        fleet_id = fleet['id']
        self.assertIn('id', fleet)
        # update fleet name
        fleet['name'] = 'newname'
        response = session.put('fleet/' + fleet_id, fleet)
        self.assertEqual(200, response.status_code)
        fleet = response.json()
        self.assertEqual('newname', fleet['name'])
        self.assertEqual('newname', fleet['vessels'][0]['fleet']['name'], str(fleet))
        # check vessel endpoint to make sure in sync
        response = session.get('vessel/' + vessel_id)
        self.assertEqual(200, response.status_code, response.text)
        vessel = response.json()
        self.assertIn('id', vessel, str(vessel))
        self.assertIn('fleet', vessel, str(vessel))
        self.assertEqual(fleet_id, vessel['fleet']['id'], str(vessel))
        self.assertEqual('newname', vessel['fleet']['name'], str(vessel))
        # check fleet endpoint to make sure in sync
        response = session.get('fleet/' + fleet_id)
        self.assertEqual(200, response.status_code, response.text)
        fleet = response.json()
        self.assertIn('id', fleet)
        self.assertEqual('newname', fleet['name'])
        self.assertIn('id', fleet['vessels'][0])

    @mongo.clean_collections_after("fleets", "vessels", 'devices')
    @authenticate('fleetmanager', 'one')
    def test_add_vessel_to_fleet_update_vessel_name(self, session):
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
        self.assertEqual(201, response.status_code, response.text)
        vessel = response.json()
        vessel_id = vessel['id']
        response = session.post('fleet', {
            "name": "gold",
            "person": {
                "name": "Ad"
            },
            "vessels": []
        })
        self.assertEqual(201, response.status_code, response.text)
        fleet = response.json()
        fleet_id = fleet['id']
        self.assertIn('id', fleet)
        # add vessel to fleet
        response = session.put('fleet/' + fleet_id + '/vessel/' + vessel_id, None)
        self.assertEqual(200, response.status_code)
        fleet = response.json()
        vessel = fleet['vessels'][0]
        # update vessel name
        vessel['name'] = 'newname'
        response = session.put('vessel/' + vessel_id, vessel)
        self.assertEqual(200, response.status_code, response.text)
        self.assertEqual('newname', vessel['name'])
        # check vessel endpoint to make sure in sync
        response = session.get('vessel/' + vessel_id)
        self.assertEqual(200, response.status_code, response.text)
        vessel = response.json()
        self.assertIn('id', vessel, str(vessel))
        self.assertIn('fleet', vessel, str(vessel))
        self.assertEqual(fleet['id'], vessel['fleet']['id'], str(vessel))
        self.assertEqual('newname', vessel['name'], str(vessel))
        # check fleet endpoint to make sure in sync
        response = session.get('fleet/' + fleet_id)
        self.assertEqual(200, response.status_code, response.text)
        fleet = response.json()
        self.assertIn('id', fleet, str(fleet))
        self.assertIn('id', fleet['vessels'][0], str(fleet))
        self.assertIn('fleet', fleet['vessels'][0], str(fleet))
        self.assertEqual(fleet['id'], fleet['vessels'][0]['fleet']['id'], str(fleet))
        self.assertEqual('newname', fleet['vessels'][0]['name'], str(fleet))

    @mongo.clean_collections_after("fleets")
    @mongo.insert_collection_data('fleet', data)
    @authenticate('fleetmanager', 'one')
    def test_update_fleet_with_invalid_input(self, session):
        rsp = session.post('fleet', {
            'name': 'Fleet1',
        })
        self.assertEqual(201, rsp.status_code, 'I should get a 201 response code')
        fleet = rsp.json()
        rsp = session.put('fleet/' + fleet['id'], {'a': ''})
        self.assertEqual(400, rsp.status_code, 'I should get a 400 response code')

    @mongo.clean_collections_after("vessels", 'devices')
    @authenticate('fleetmanager', 'one')
    def test_update_vessel_with_valid_input(self, session):
        rsp = session.post('vessel', {
            'name': 'vessel1',
            "devices": [{"type": "omnicom-vms", "deviceId": "900000234234"}],
        })
        self.assertEqual(201, rsp.status_code, 'I should get a 201 response code. Response: %s' % rsp.text)
        vessel = rsp.json()
        vessel['name'] = 'newname'
        response = session.put('vessel/' + vessel['id'], vessel)
        self.assertEqual(200, response.status_code, 'I should get a 200 response code')
        vessel = response.json()
        self.assertIn('id', vessel)

    @mongo.clean_collections_after("vessels", 'devices')
    @authenticate('fleetmanager', 'one')
    def test_update_vessel_with_invalid_input(self, session):
        rsp = session.post('vessel', {
            'name': 'vessel1',
            "devices": [{"type": "omnicom-vms", "deviceId": "900000234234"}],
        })
        self.assertEqual(201, rsp.status_code, 'I should get a 201 response code')
        rsp = session.get('vessel')
        self.assertEqual(200, rsp.status_code, 'I should get a 200 response code')
        jsonArray = rsp.json()
        self.assertGreater(len(jsonArray), 0)
        vessel_id = jsonArray[0]['id']
        rsp = session.put('vessel/' + vessel_id, {'b': 1})
        self.assertEqual(400, rsp.status_code, 'I should get a 400 response code')


if __name__ == '__main__':
    unittest.main()
