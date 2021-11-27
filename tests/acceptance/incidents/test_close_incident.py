import config
import json
import requests
import unittest
from utilities import mongo
from utilities.session import authenticate


class CloseIncident(unittest.TestCase):

    @authenticate('incident', 'manager')
    def test_close_incident_without_selecting_an_outcome(self, session):
        response = session.post('incident', {
            'incidentId': '',
            'name': 'incmanage',
            'assignee': 'incident',
            'commander': 'incident',
            'phase': 2,
            'type': 'Unlawful'
        })
        self.assertEqual(201, response.status_code)
        body = response.json()
        self.assertIn('id', body)
        incident_id = body['id']
        self.assertIn('incidentId', body)
        self.assertEqual(incident_id, body['id'])
        self.assertEqual('incmanage', body['name'])
        self.assertEqual(2, body['phase'])
        self.assertEqual('Unlawful', body['type'])
        response = session.put('incident/' + incident_id, {
            'id': incident_id,
            'incidentId': '1',
            'name': 'incmanage',
            'type': 'Unlawful',
            'phase': 2,
            'commander': 'incident',
            'state': 2,
            'assignee': 'incident',
            'log': [
                    {
                        'id': incident_id,
                        'timestamp': {
                            'seconds': 1507308442,
                            'nanos': 3764296
                        }
                    }
            ],
            'synopsis': 'this is a closed incident now'
        })
        self.assertEqual(400, response.status_code)
        body = response.json()
        self.assertEqual('outcome', body[0]['property'])
        self.assertEqual('Required', body[0]['rule'])
        self.assertEqual('Required non-empty property', body[0]['message'])

    @authenticate('incident', 'manager')
    def test_close_incident_without_entering_a_synopsis(self, session):
        response = session.post('incident', {
            'incidentId': '',
            'name': 'inc1',
            'assignee': 'incident',
            'commander': 'incident',
            'phase': 1,
            'type': 'Unlawful'
        })
        self.assertEqual(201, response.status_code)
        body = response.json()
        self.assertIn('id', body)
        incident_id = body['id']
        self.assertEqual(incident_id, body['id'])
        self.assertEqual('inc1', body['name'])
        self.assertEqual(1, body['phase'])
        self.assertEqual('Unlawful', body['type'])
        response = session.put('incident/' + incident_id, {
            'id': incident_id,
            'incidentId': '1',
            'name': 'incmanage',
            'type': 'Unlawful',
            'phase': 2,
            'commander': 'incident',
            'state': 2,
            'assignee': 'incident',
            'log': [
                    {
                        'id': incident_id,
                        'timestamp': {
                            'seconds': 1507308442,
                            'nanos': 3764296
                        }
                    }
            ],
            'outcome': 'Fatality'
        })
        self.assertEqual(400, response.status_code)
        body = response.json()
        self.assertEqual('synopsis', body[0]['property'])
        self.assertEqual('Required', body[0]['rule'])
        self.assertEqual('Required non-empty property', body[0]['message'])

    @authenticate('standard', 'user')
    def test_close_incident_as_a_standard_user(self, session):
        pass

if __name__ == '__main__':
    unittest.main()
