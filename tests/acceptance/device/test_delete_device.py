import config
import json
import requests
import unittest
from utilities.session import authenticate
from utilities import mongo


class DeleteDevice(unittest.TestCase):

    @mongo.clean_collections_after('devices')
    @authenticate('admin', 'admin')
    def test_delete_a_devices_valid_success(self, session):
        # create a device, then delete it
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
        self.assertEqual(201, rsp.status_code, rsp.text)
        body = rsp.json()
        self.assertIn('id', body)
        device_id = body['id']
        print('device_id', body['id'])
        rsp = session.delete('device/' + device_id)
        self.assertEqual(204, rsp.status_code, rsp.text)


if __name__ == '__main__':
    unittest.main()
