import json
import unittest
from fleets.data_fleet import data as fleets
from fleets.data_vessel import data as vessels
from utilities import mongo
from utilities.session import authenticate


class GetFleetData(unittest.TestCase):

    @mongo.clean_collections_after('fleets')
    @mongo.insert_collection_data('fleets', fleets)
    @authenticate('admin', 'admin')
    def test_get_list_of_fleets_valid_permissions(self, session):
        rsp = session.get('fleet')
        self.assertEqual(200, rsp.status_code, 'I should get a 200 response code')
        fleet = rsp.json()[0]
        self.assertIn('id', fleet)

    @mongo.clean_collections_after('fleets')
    @authenticate('admin', 'admin')
    def test_get_fleet_that_is_registered_in_the_system(self, session):
        rsp = session.post('fleet', {
            "name": "gould",
            "person": {
                "name": "Ad"
            }
        })
        self.assertEqual(201, rsp.status_code, rsp.text)
        rsp = session.get('fleet')
        self.assertEqual(200, rsp.status_code, 'I should get a 200 response code')
        fleet_id = rsp.json()[0]['id']
        rsp = session.get('fleet/' + fleet_id)
        self.assertEqual(200, rsp.status_code, 'I should get a 200 response code')
        fleet = rsp.json()
        self.assertIn('id', fleet)

    @mongo.clean_collections_after('fleets')
    @mongo.insert_collection_data('fleets', fleets)
    @authenticate('admin', 'admin')
    def test_get_fleet_that_is_not_registered_in_the_system(self, session):
        rsp = session.get('fleet/non-existent-fleet-id')
        self.assertEqual(400, rsp.status_code, 'I should get a 400 response code')

    @mongo.clean_collections_after('fleets')
    @mongo.insert_collection_data('fleets', fleets)
    @authenticate('usermanager', 'one')
    def test_get_list_of_fleets_invalid_permissions(self, session):
        rsp = session.get('fleet')
        self.assertEqual(403, rsp.status_code, 'I should get a 403 response code')

    @mongo.clean_collections_after('vessels')
    @authenticate('admin', 'admin')
    def test_get_list_of_vessels_with_valid_permissions(self, session):
        rsp = session.post('vessel', {
            "name": "The vessel",
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
        rsp = session.get('vessel')
        self.assertEqual(200, rsp.status_code, 'I should get a 200 response code')

    @mongo.clean_collections_after('vessels', 'devices')
    @authenticate('admin', 'admin')
    def test_get_vessel_that_is_registered_in_the_system(self, session):
        rsp = session.post('vessel', {
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
        self.assertEqual(201, rsp.status_code, 'I should get a 201 response code')
        rsp = session.get('vessel')
        self.assertEqual(200, rsp.status_code, 'I should get a 200 response code')
        vessel = rsp.json()[0]
        rsp = session.get('vessel/'+vessel['id'])
        self.assertEqual(200, rsp.status_code, 'I should get a 200 response code')
        vessel = rsp.json()
        self.assertIn('id', vessel)
        self.assertIn('devices', vessel)

    @authenticate('admin', 'admin')
    def test_get_vessel_that_is_not_registered_in_the_system(self, session):
        response = session.get('vessel/5a6305a677f64b037e177899')
        self.assertEqual(404, response.status_code, 'I should get a 404 response code')

    @authenticate('usermanager', 'one')
    def test_get_list_of_vessels_invalid_permissions(self, session):
        rsp = session.get('vessel')
        self.assertEqual(403, rsp.status_code, 'I should get a 403 response code')

    @mongo.clean_collections_after('fleets')
    @mongo.insert_collection_data('fleets', fleets)
    @authenticate('usermanager', 'one')
    def test_get_fleet_that_is_registered_in_the_system_with_invalid_permissions(self, session):
        rsp = session.get('fleet')
        self.assertEqual(403, rsp.status_code, 'I should get a 403 response code')
        # It is not possible to obtain a fleet-id with invalid permissions.
        # fleet_id = rsp.json()[0]['id']
        # rsp = session.get('fleet/' + fleet_id)
        # self.assertEqual(403, rsp.status_code, 'I should get a 403 response code')

if __name__ == '__main__':
    unittest.main()
