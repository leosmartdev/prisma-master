import unittest
import time
from utilities.session import authenticate
from utilities import mongo


def wait():
    time.sleep(2)

class TestManualTrack(unittest.TestCase):

    @mongo.clean_collections_after('tracks', 'registry', 'trackex')
    @authenticate('admin', 'admin')
    def test_create_a_track_with_valid_input_returns_success(self, session):
        track = self.createTrack(session, 10, 10, '')
        wait()
        registry_id = track['registryId']
        rsp = session.get('registry/%s' % registry_id)
        self.assertEqual(200, rsp.status_code, rsp.text)
        registry = rsp.json()
        self.assertIsNotNone(registry, rsp.text)
        self.assertIn('target', registry, rsp.text)
        self.assertIsNotNone(registry['target'], rsp.text)
        self.assertIn('position', registry['target'], rsp.text)
        self.assertIsNotNone(registry['target']['position'], rsp.text)
        self.assertIn('latitude', registry['target']['position'], rsp.text)
        self.assertIn('longitude', registry['target']['position'], rsp.text)
        self.assertEqual(registry_id, registry['registryId'], rsp.text)
        self.assertEqual('Manual', registry['target']['type'], rsp.text)

    @mongo.clean_collections_after('tracks', 'registry', 'trackex', 'activity')
    @authenticate('admin', 'admin')
    def test_update_a_track_with_valid_input_returns_success(self, session):
        track = self.createTrack(session, 10, 10, '')
        wait()
        registry_id = track['registryId']
        self.createTrack(session, 20, 30, registry_id) # update the track
        wait()
        rsp = session.get('registry/%s' % registry_id)
        self.assertEqual(200, rsp.status_code, rsp.text)
        registry = rsp.json()
        self.assertIsNotNone(registry, rsp.text)
        self.assertIn('target', registry, rsp.text)
        self.assertIsNotNone(registry['target'], rsp.text)
        self.assertIn('position', registry['target'], rsp.text)
        self.assertIsNotNone(registry['target']['position'], rsp.text)
        self.assertIn('latitude', registry['target']['position'], rsp.text)
        self.assertIn('longitude', registry['target']['position'], rsp.text)
        self.assertEqual(20, registry['target']['position']['latitude'], rsp.text)
        self.assertEqual(30, registry['target']['position']['longitude'], rsp.text)

    @unittest.skip("tracks needs to be fixed to handle timeout messages in the 15 minute window")
    @mongo.clean_collections_after('tracks', 'registry', 'trackex')
    @authenticate('admin', 'admin')
    def test_delete_a_track_with_valid_input_returns_success(self, session):
        track = self.createTrack(session, '')
        wait()
        registry_id = track['registryId']
        rsp = session.delete('track/%s' % registry_id)
        self.assertEqual(200, rsp.status_code, rsp.text)
        wait()
        rsp = session.get('track/%s' % registry_id)
        self.assertEqual(404, rsp.status_code, rsp.text)

    def createTrack(self, session, latitude, longitude, registry_id):
        rsp = session.post('track', {
            'registryId': registry_id,
            'name': 'name',
            'latitude': latitude,
            'longitude': longitude,
        })
        self.assertEqual(200, rsp.status_code, rsp.text)
        track = rsp.json()
        self.assertIn('registryId', track)
        return track