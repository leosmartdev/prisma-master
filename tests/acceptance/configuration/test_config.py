import unittest

import config
from utilities.session import authenticate


class Configuration(unittest.TestCase):
    @authenticate('admin', 'admin')
    def test_get_config_put_config(self, session):
        response = session.request('GET', config.config)
        self.assertEqual(200, response.status_code, str(response))
        configuration = response.json()
        configuration['brand']['name'] = 'PRISMA Test'
        response = session.put('config', configuration)
        self.assertEqual(200, response.status_code, str(response))
        newConfiguration = response.json()
        self.assertEqual(configuration['brand']['name'], newConfiguration['brand']['name'], str(newConfiguration))
        # get again
        response = session.request('GET', config.config)
        self.assertEqual(200, response.status_code, str(response))
        configuration = response.json()
        self.assertEqual(newConfiguration['brand']['name'], configuration['brand']['name'], str(configuration))

    @authenticate('admin', 'admin')
    def test_get_config_put_config_create_incident(self, session):
        response = session.request('GET', config.config)
        self.assertEqual(200, response.status_code, str(response))
        configuration = response.json()
        configuration['site']['incidentIdPrefix'] = 'TEST-INC-'
        response = session.put('config', configuration)
        self.assertEqual(200, response.status_code, str(response))
        newConfiguration = response.json()
        self.assertEqual(configuration['site']['incidentIdPrefix'], newConfiguration['site']['incidentIdPrefix'],
                         str(newConfiguration))
        # create incident
        response = session.post('incident', {
            'incidentId': '',
            'name': 'SAR exercise',
            'assignee': 'incident',
            'commander': 'incident',
            'phase': 2,
            'type': 'Unlawful',
        })
        self.assertEqual(201, response.status_code)
        incident = response.json()
        self.assertIn(configuration['site']['incidentIdPrefix'], incident['incidentId'], str(incident))

    @authenticate('standard', 'user')
    def test_put_config_unauthorized(self, session):
        response = session.request('GET', config.config)
        self.assertEqual(200, response.status_code, str(response))
        configuration = response.json()
        response = session.put('config', configuration)
        self.assertEqual(403, response.status_code, str(response))

    @authenticate('admin', 'admin')
    def test_put_config_notfound(self, session):
        response = session.request('GET', config.config)
        self.assertEqual(200, response.status_code, str(response))
        configuration = response.json()
        configuration['id'] = ''
        response = session.put('config', configuration)
        self.assertEqual(404, response.status_code, str(response))

    @authenticate('admin', 'admin')
    def test_put_config_invalid(self, session):
        configuration = 'garbage'
        response = session.put('config', configuration)
        self.assertEqual(400, response.status_code, str(response))
