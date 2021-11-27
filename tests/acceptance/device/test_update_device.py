import unittest

from utilities import mongo
from utilities.session import authenticate


class UpdateDevice(unittest.TestCase):

    @mongo.clean_collections_after('devices')
    @authenticate('admin', 'admin')
    def test_update_device_valid_success(self, session):
        # create a device
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
        device = rsp.json()
        self.assertIn('id', device)
        device_id = device['id']
        device_type = 'qwe'
        device['type'] = device_type
        # update the device
        rsp = session.put('device/' + device_id, device)
        self.assertEqual(200, rsp.status_code, rsp.text)
        device2 = rsp.json()
        self.assertIn('id', device2)
        self.assertEqual(device2['id'], device_id)
        # get the device
        rsp = session.get('device/' + device_id)
        self.assertEqual(200, rsp.status_code, rsp.text)
        device3 = rsp.json()
        self.assertIn('id', device3)
        self.assertEqual(device3['id'], device_id)
        self.assertEqual(device3['type'], device_type)


if __name__ == '__main__':
    unittest.main()
