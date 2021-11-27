"""
Tests device multicast configuration using multicast and websocket connections.
"""
import asyncio
import json
import logging  # Uncomment to see websocket logging on stdout.
import unittest
from datetime import datetime

from utilities import mongo
from utilities.session import authenticate

# Uncomment to see websocket logging on stdout.
logger = logging.getLogger('websockets')
logger.setLevel(logging.DEBUG)
logger.addHandler(logging.StreamHandler())


def elapsedTime(startTime):
    """ Returns the timedelta between the start time and current time."""
    time_delta = datetime.now() - startTime
    return time_delta.total_seconds()


class TestDeviceMulticast(unittest.TestCase):
    def setUp(self):
        self.device = self.get_device(deviceId="4")

    @authenticate('admin', 'admin')
    def get_device(self, session, deviceId="1"):
        response = session.post('device', {
            "type": "omnicom-vms",
            "deviceId": deviceId,
            "networks": [
                {
                    "subscriberId": "333234010030454",
                    "type": "satellite-data",
                    "providerId": "iridium",
                }
            ]
        })
        self.assertEqual(201, response.status_code, response.text)
        device = response.json()
        return device

    @mongo.clean_collections_after('multicasts', 'transmissions', 'devices')
    @authenticate('admin', 'admin')
    def test_device_multicast_get_configuration(self, session):
        """ 	Tests given a device in the system, you can query it's configuration using device multicast	API.        """
        self.assertNotEqual(self.device, None, "device not set")

        # Everything needs to be run inside a coroutine so that the asyncio
        # loop works (and because unittest wont run async code directly)
        async def run_test():
            async with session.websocket_connect('/') as websocket:
                # Send multicast request.
                response = session.post('multicast/device/{0}'.format(self.device['id']), {
                    "payload": {
                        "@type": "type.googleapis.com/prisma.tms.moc.DeviceConfiguration",
                        "id": self.device['id'],
                        "configuration": {
                            "@type": "type.googleapis.com/prisma.tms.omnicom.OmnicomConfiguration",
                            "action": "RequestGlobalParameters"
                        }
                    }
                })
                self.assertEqual(response.status_code, 201,
                                 "Multicast config POST failed with incorrect status code. RESPONSE: {0}".format(
                                     response.text))
                multicast = response.json()
                self.assertEqual(len(multicast['transmissions']), 1,
                                 "Multicast response should contain exactly 1 transmission message")

                transmissionId = multicast['transmissions'][0]['id']
                transmission_count = 0
                expected_transmission_count = 1
                device_update_count = 0
                startTime = datetime.now()

                # Read all websocket responses and find the transmission we want. we are expecting many
                # transmissions, so if we don't get those within about 10 seconds fail.
                while transmission_count < expected_transmission_count and elapsedTime(startTime) < 10:
                    message = await asyncio.wait_for(websocket.recv(), timeout=20)
                    envelope = json.loads(message)
                    if envelope['type'] == 'Device/UPDATE' and envelope['device']['id'] == self.device['id']:
                        device_update_count += 1

                    # we only care about transmission update messages
                    if envelope['type'] == 'Transmission/UPDATE' and envelope['transmission']['parentId'] == multicast['id']:
                        transmission = envelope['transmission']
                        transmission_count += 1
                        logger.info(transmission_count)
                        # transmission asserts
                        self.assertEqual(transmission['id'], transmissionId, 'Transmission ID is required.')
                        self.assertEqual(transmission['destination']['type'], 'iridium',
                                         'Transmission destination is expected to be a device object')
                        self.assertEqual(transmission['destination']['id'], '333234010030454',
                                         'Transmission destination id is expected to be the device id')
                        # First transmission received
                        if transmission_count == 1:
                            self.assertEqual(transmission['state'], 'Pending',
                                             'Expect first transmission to be Pending (STATE==1)')
                        # Second transmission received
                        elif transmission_count == 2:
                            self.assertEqual(transmission['state'], 'Partial',
                                             'Expect second transmission to be Partial (STATE==4)')
                            self.assertEqual(transmission['status']['code'], 211,
                                             'Expect second transmission status object to have 211 error code')
                            self.assertEqual(transmission['status']['message'], 'transmission sent')
                        # Last transmission received
                        elif transmission_count == expected_transmission_count:
                            self.assertEqual(transmission['state'], 'Success',
                                             'Expect final transmission to be Success (STATE==2)')
                            self.assertEqual(transmission['status']['code'], 200,
                                             'Expect final transmission status object to have 200 error code')
                            self.assertEqual(transmission['status']['message'], 'transmission successful')
                self.assertEqual(transmission_count, expected_transmission_count,
                                 "Expected transmissions over the socket.")

        # Run the actual test
        asyncio.get_event_loop().run_until_complete(run_test())


if __name__ == "__main__":
    unittest.main()
