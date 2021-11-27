import unittest

from utilities import mongo
from utilities.session import authenticate


class GetDevice(unittest.TestCase):

    @mongo.clean_collections_after('devices')
    @authenticate('admin', 'admin')
    def test_get_a_device_valid_success(self, session):
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
        device = rsp.json()
        self.assertIn('id', device)
        device_id = device['id']
        # get device
        rsp = session.get('device/' + device_id)
        self.assertEqual(200, rsp.status_code, rsp.json())
        body = rsp.json()
        self.assertIn('id', body)
        self.assertEqual(body['id'], device_id)


if __name__ == '__main__':
    unittest.main()
