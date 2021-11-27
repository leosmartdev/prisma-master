import unittest

from utilities import mongo
from utilities.session import authenticate
from .data_mongo import zones, devices


class MulticastMessagingService(unittest.TestCase):

    @unittest.skip("this capability will be available in the next release")
    @authenticate('fleetmanager', 'one')
    def test_broadcast_message_to_a_fleet_with_no_vessels(self, session):
        pass

    @unittest.skip("this capability will be available in the next release")
    @authenticate('user', 'manager')
    def test_broadcast_message_to_a_fleet_with_no_vessels_invalid_permissions(self, session):
        pass

    @authenticate('admin', 'admin')
    @mongo.clean_collections_after('multicasts', 'devices')
    @mongo.insert_collection_data('zones', zones)
    @mongo.insert_collection_data('transmissions', [{}])
    def test_broadcast_message_to_a_vessel_geofence_upload(self, session):
        resp = session.post('device', {
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
        self.assertEqual(201, resp.status_code, str(resp))
        device = resp.json()
        self.assertIn('id', device)
        # test twebd - should be able to record a transmission and to send to tgwad
        resp = session.post("multicast/device/%s" % device["id"], {
            "configuration": {
                "@type": "type.googleapis.com/prisma.tms.moc.DeviceConfiguration",
                "configuration": {
                    "@type": "type.googleapis.com/prisma.tms.omnicom.OmnicomConfiguration",
                    "action": "Geofence",
                    "zoneId": str(zones[0]["_id"]),
                }
            }
        })
        self.assertEqual(resp.status_code, 400, resp.text)
        pass
        return
        # TODO remove when functionality added
        self.assertEqual(resp.status_code, 201, str(resp))
        rdata = resp.json()
        self.assertListEqual(rdata["destinations"], [device["deviceid"]])
        self.assertEqual(rdata["deviceType"], "omnicom-vms")
        transmission = mongo.get_trident_collection("transmissions").find_one({
            "me.messageId": rdata["transmissions"][0]["messageId"]
        })
        self.assertIsNotNone(transmission)
        # okay, so we should be able to see two records in the "live" collection
        # 1: new
        # 2: update - it means we received data from a beacon
        # Sleep some time, cause the process takes different time this depends on host machine
        records = mongo.get_trident_collection("transmissions").find({"objid": transmission["_id"]})
        for i, action in enumerate(records):
            self.assertLess(i, 3)
            self.assertIn(action, ["New", "Update"])

    @authenticate('fleetmanager', 'one')
    def test_broadcast_message_to_a_vessel_with_valid_permissions(self, session):
        pass

    @authenticate('usermanager', 'one')
    def test_broadcast_message_to_a_vessel_as_a_user_manager(self, session):
        pass

    @authenticate('standard', 'user')
    def test_broadcast_message_to_a_vessel_as_a_standard_user(self, session):
        pass

    @authenticate('fleetmanager', 'one')
    def test_broadcast_message_to_a_fleet_with_updated_configuration_as_a_binary(self, session):
        pass


if __name__ == '__main__':
    unittest.main()
