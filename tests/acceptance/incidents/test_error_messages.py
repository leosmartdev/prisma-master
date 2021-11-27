import config
import json
import requests
import unittest
from utilities.session import authenticate


class ErrorMessages(unittest.TestCase):

    @authenticate('incident', 'manager')
    def test_with_invalid_character_length_for_incident_name(self, session):
        r = session.post('incident', {
            'incidentId': '',
            'name': 'S',
            'assignee': 'incident',
            'commander': 'incident',
            'phase': 2,
            'type': 'Unlawful',
        })
        self.assertEqual(400, r.status_code)
        body = r.json()
        self.assertEqual('name', body[0]['property'])
        self.assertEqual('MinLength', body[0]['rule'])
        self.assertEqual('Out of range.', body[0]['message'])

    @authenticate('incident', 'manager')
    def test_with_type_field_blank_in_incident_form(self, session):
        r = session.post('incident', {
            'incidentId': '',
            'name': 'ST',
            'assignee': 'incident',
            'commander': 'incident',
            'phase': 1,
            'type': '',
        })
        self.assertEqual(400, r.status_code)
        body = r.json()
        self.assertEqual('type', body[0]['property'])
        self.assertEqual('Required', body[0]['rule'])
        self.assertEqual('Required non-empty property', body[0]['message'])


if __name__ == '__main__':
    unittest.main()
