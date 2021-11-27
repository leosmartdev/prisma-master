import config
import json
import requests
import unittest
from utilities import mongo
from utilities.session import authenticate

class CreateIncident(unittest.TestCase):

    @authenticate('admin', 'admin')
    def test_valid_input_returns_201_response(self, session):
        response = session.post('incident', {
            'incidentId': '',
            'name': 'SAR exercise',
            'assignee': 'incident',
            'commander': 'incident',
            'phase': 2,
            'type': 'Unlawful',
        })
        self.assertEqual(201, response.status_code)
        body = response.json()
        self.assertIn('id', body)
        self.assertIn('incidentId', body)
        self.assertEqual('SAR exercise', body['name'])  
        self.assertEqual(2, body['phase'])  
        self.assertEqual('Unlawful', body['type'])

    @authenticate('standard', 'user')
    def test_create_incident_as_a_standard_user(self, session):
        response = session.post('incident', {
            'incidentId': '',
            'name': 'test403',
            'assignee': 'incident',
            'commander': 'incident',
            'phase': 1,
            'type': 'Unlawful'
        })
        self.assertEqual(403, response.status_code)

if __name__ == '__main__':
    unittest.main()
