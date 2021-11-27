import unittest

from utilities import mongo
from utilities.session import authenticate
from device.data_vessel import data as vessels_data


class CreateDevice(unittest.TestCase):

    @mongo.clean_collections_after('vessels', 'devices')
    @authenticate('fleetmanager', 'one')
    def test_create_a_device_with_valid_input_returns_success(self, session):
        # create a vessel without devices
        rsp = session.post('vessel', {
            "name": "zuby",
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
        vessel = rsp.json()
        self.assertIn('id', vessel)
        vessel_id = vessel['id']
        #  add the device to the vessel
        rsp = session.post('device/vessel/' + vessel_id, {
            "type": "omnicom-vms",
            "deviceId": "1234",
            "networks": [
                {
                    "subscriberId": "356938035643809",
                    "type": "satellite-data",
                    "providerId": "Iridium"
                }
            ]
        })
        self.assertEqual(201, rsp.status_code, 'I should get a 201 response code')
        body = rsp.json()
        self.assertIn('id', body)
        self.assertEqual('omnicom-vms', body['type'])
        # get vessel
        response = session.get('/vessel/' + vessel_id)
        self.assertEqual(200, response.status_code, str(response))
        vessel = response.json()
        self.assertEqual('omnicom-vms', vessel['devices'][0]['type'], str(vessel))

    @mongo.clean_collections_after('devices')
    @authenticate('fleetmanager', 'one')
    def test_create_a_device_without_vessel_with_valid_input_returns_success(self, session):
        rsp = session.post('device', {
            "type": "omnicom-vms",
            "deviceId": "900000234234",
            "networks": [
                {
                    "subscriberId": "356938035643809",
                    "type": "satellite-data",
                    "providerId": "Iridium"
                }
            ]
        })
        self.assertEqual(201, rsp.status_code, 'I should get a 201 response code')
        body = rsp.json()
        self.assertIn('id', body)

    @mongo.clean_collections_after('vessel', 'devices')
    @mongo.insert_collection_data('vessels', vessels_data)
    @authenticate('fleetmanager', 'one')
    def test_create_a_device_with_invalid_input(self, session):
        vessel_id = '0a000a00f0ffcb00fa000e0f'
        rsp = session.post('device/vessel/' + vessel_id, {
            "type": "omnicom-vms",
            "deviceId": "900000234234",
            "networks": [
                {
                    "subscriberId": "356938035643809",
                    "type": "satellite-data",
                    "providerId": "Iridium"
                }
            ]
        })
        self.assertEqual(404, rsp.status_code, 'I should get a 404 response code')


if __name__ == '__main__':
    unittest.main()
