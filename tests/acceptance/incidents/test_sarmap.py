import json
import time
import unittest

import config
from utilities import mongo
from utilities.session import authenticate


class GetSARMAP(unittest.TestCase):

    def createTrack(self, session, name, latitude, longitude):
        rsp = session.post('track', {
            "name": name,
            "latitude": latitude,
            "longitude": longitude
        })
        body = rsp.json()
        self.assertEqual(200, rsp.status_code)
        self.assertIn('registryId', body)
        # send another update
        time.sleep(0.1)
        rsp = session.post('track', body)
        self.assertEqual(200, rsp.status_code)
        return body['registryId']

    def createIncident(self, session, name):
        rsp = session.post('incident', {
            'incidentId': '',
            'name': name,
            'assignee': 'incident',
            'commander': 'incident',
            'phase': 2,
            'type': 'Unlawful',
        })
        self.assertEqual(201, rsp.status_code)
        body = rsp.json()
        self.assertIn('id', body)
        self.assertIn('incidentId', body)
        incident_id = body['id']
        return incident_id

    def createIncidentLogEntry(self, session, incident_id, track_id):
        rsp = session.post('incident/' + incident_id + '/log', {
            "type": "ENTITY",
            "entity": {
                "id": track_id,
                "type": "registry",
            }
        }, timeout=10)
        self.assertEqual(200, rsp.status_code)
        return rsp

    def getRegistry(self, session, registry_id):
        rsp = session.get('registry/' + registry_id)
        if rsp.status_code == 404:
            time.sleep(2)
            rsp = session.get('registry/' + registry_id)
        self.assertEqual(200, rsp.status_code)
        body = rsp.json()
        self.assertIn('target', body)
        return rsp

    @mongo.clean_collections_after('tracks', 'registry', 'trackex', 'activity')
    @authenticate('admin', 'admin')
    def test_sarmap(self, session):
        client = mongo.connect()
        db = client[config.MONGO_DB_TRIDENT]
        incidents_collection = db["incidents"]
        incidents_collection.delete_many({})
        tracks_collection = db["tracks"]
        tracks_collection.delete_many({})
        client.close()
        # sanity check
        sarmap_url = 'http://localhost:8081/api/v2/sarmap.json'
        rsp = session.request('GET', sarmap_url)
        self.assertEqual('[]', rsp.text, 'Expected an empty array.')
        # two tracks add to one incident
        registry_id1 = self.createTrack(session, 'name of track 1', 10.3, 40.5)
        self.getRegistry(session, registry_id1)
        registry_id2 = self.createTrack(session, 'name of track 2', 20.2, 70.8)
        self.getRegistry(session, registry_id2)
        incident_id1 = self.createIncident(session, 'Incident 47 - MOB')
        self.createIncidentLogEntry(session, incident_id1, registry_id1)
        self.createIncidentLogEntry(session, incident_id1, registry_id2)
        # one track to two incidents
        registry_id3 = self.createTrack(session, 'name of track 3', 30.5, 90.9)
        self.getRegistry(session, registry_id3)
        incident_id2 = self.createIncident(session, 'Incident 112 lost contact')
        self.createIncidentLogEntry(session, incident_id2, registry_id3)
        incident_id3 = self.createIncident(session, 'Incident 234 unknown')
        self.createIncidentLogEntry(session, incident_id3, registry_id3)
        time.sleep(5)
        # sarmap has 4 elements
        rsp = session.request('GET', sarmap_url)
        body = json.loads(rsp.text)
        self.assertTrue(isinstance(body, list), 'Expected an array.')
        self.assertEqual(4, len(body), 'Expected a array of 4.')

        for geo_json_item in body:
            self.assertIn('properties', geo_json_item)
            properties = geo_json_item['properties']
            self.assertIn('deviceType', properties)
            self.assertIn('label', properties)

if __name__ == '__main__':
    unittest.main()
