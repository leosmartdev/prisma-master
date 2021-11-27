import unittest

from utilities import mongo
from utilities.session import authenticate


class CreateVessel(unittest.TestCase):

    @mongo.clean_collections_after('vessels')
    @authenticate('fleetmanager', 'one')
    def test_create_a_device_add_to_vessel_check_vessel_endpoint(self, session):
        response = session.post('device', {
            "type": "omnicom-vms",
            "deviceId": "900000134234",
            "networks": [
                {
                    "subscriberId": "356738035649999",
                    "type": "satellite-data",
                    "providerId": "Iridium"
                }
            ]
        })
        self.assertEqual(201, response.status_code, 'I should get a 201 response code')
        device = response.json()
        self.assertIn('id', device)
        device_id = device['id']
        rsp = session.post('vessel', {
            "name": "integration",
            "devices": [
                {
                    "id": device_id,
                }
            ],
            "crew": [
                {
                    "name": "Andrew Razorsharp"
                }]
        })
        self.assertEqual(201, rsp.status_code, 'I should get a 201 response code')
        vessel = rsp.json()
        self.assertIn('id', vessel)
        vessel_id = vessel['id']
        self.assertEqual(1, len(vessel['devices']))
        self.assertEqual(device_id, vessel['devices'][0]['id'], str(vessel))
        # get vessel
        response = session.get('vessel/' + vessel_id)
        self.assertEqual(200, response.status_code, 'I should get a 200 response code')
        # get device
        response = session.get('device/' + device_id)
        self.assertEqual(200, response.status_code, 'I should get a 200 response code')

    @mongo.clean_collections_after('vessels')
    @authenticate('fleetmanager', 'one')
    def test_create_a_vessel_device_check_device_endpoint(self, session):
        rsp = session.post('vessel', {
            "name": "integration",
            "devices": [
                {
                    "type": "omnicom-vms",
                    "deviceId": "900000134234",
                    "networks": [
                        {
                            "subscriberId": "356738035649999",
                            "type": "satellite-data",
                            "providerId": "Iridium"
                        }
                    ]
                },
                {
                    "type": "omnicom-solar",
                    "deviceId": "900000111111",
                    "networks": [
                        {
                            "subscriberId": "356738035641111",
                            "type": "satellite-data",
                            "providerId": "Iridium"
                        }
                    ]
                }
            ],
            "crew": [
                {
                    "name": "Alexander Ramon"
                },
                {
                    "name": "Andrew Razorsharp"
                }]
        })
        self.assertEqual(201, rsp.status_code, 'I should get a 201 response code')
        body = rsp.json()
        self.assertIn('id', body)
        self.assertEqual(2, len(body['devices']))
        response = session.get('device/' + body['devices'][0]['id'])
        self.assertEqual(200, response.status_code, 'I should get a 200 response code')
        response = session.get('device/' + body['devices'][1]['id'])
        self.assertEqual(200, response.status_code, 'I should get a 200 response code')

    @mongo.clean_collections_after('vessels', 'devices')
    @authenticate('fleetmanager', 'one')
    def test_create_a_vessel_with_valid_input_returns_success(self, session):
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
                    "id": "5a392446e1b2f9c0bab3b8b4",
                    "name": "Adalberto Zetta"
                },
                {
                    "id": "5a392446e1b2f9c0bab3b8b5",
                    "name": "Adrian Yuri"
                }
            ]
        })
        vessel = rsp.json()
        self.assertEqual(201, rsp.status_code, str(vessel))
        self.assertIn('id', vessel)

    @mongo.clean_collections_after('vessels')
    @authenticate('fleetmanager', 'one')
    def test_create_a_vessel_that_has_more_than_one_device(self, session):
        rsp = session.post('vessel', {
            "name": "integration",
            "devices": [
                {
                    "type": "omnicom-vms",
                    "deviceId": "900000134234",
                    "networks": [
                        {
                            "subscriberId": "356738035643809",
                            "type": "satellite-data",
                            "providerId": "Iridium"
                        }
                    ]
                }
            ],
            "crew": [
                {
                    "name": "Alexander Ramon"
                },
                {
                    "name": "Andrew Razorsharp"
                }]
        })
        self.assertEqual(201, rsp.status_code, 'I should get a 201 response code')
        body = rsp.json()
        self.assertIn('id', body)

    @authenticate('admin', 'admin')
    def test_create_a_vessel_with_invalid_input_returns_error(self, session):
        rsp = session.post('vessel', {
            "crew": []
        })

        self.assertEqual(400, rsp.status_code, 'I should get a 400 response code')

    @authenticate('usermanager', 'one')
    def test_create_a_vessel_with_invalid_permissions(self, session):
        rsp = session.post('vessel', {
            "devices": [{}]
        })
        self.assertEqual(403, rsp.status_code, 'I should get a 403 response code')

    @authenticate('usermanager', 'one')
    def test_create_a_vessel_with_invalid_permissions_that_has_more_than_one_device(self, session):
        response = session.post('vessel', {
            "devices": [
                {
                    "networks": []
                },
                {
                    "networks": []
                }
            ],
            "crew": []
        })
        self.assertEqual(403, response.status_code, 'I should get a 403 response code')

    @authenticate('usermanager', 'one')
    def test_create_a_vessel_with_valid_input_as_a_user_manager(self, session):
        rsp = session.post('vessel', {
            "devices": [{}]
        })
        self.assertEqual(403, rsp.status_code, 'I should get a 403 response code')

    @authenticate('incident', 'manager')
    def test_create_a_vessel_with_valid_input_as_a_incident_manager(self, session):
        rsp = session.post('vessel', {
            "devices": [{}]
        })
        self.assertEqual(403, rsp.status_code, 'I should get a 403 response code')


if __name__ == '__main__':
    unittest.main()
