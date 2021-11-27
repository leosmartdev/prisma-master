import config
import json
import requests
import unittest
from utilities.session import authenticate


class CreateUsers(unittest.TestCase):

    @authenticate('admin', 'admin')
    def test_create_standard_user_when_authenticated_as_an_admin_user(self, session):
        session.delete_user('standard', if_exists=True)
        user = {
            'userId': 'standard',
            'password': 'user',
            'roles': ['StandardUser']
        }
        resp = session.request('POST', config.auth + '/user', json=user)

        status_code = resp.status_code
        self.assertEqual(201, status_code)
        self.assertDictEqual(resp.json(), {
            'userId': 'standard',
            'state': 'initialized',
            'roles': ['StandardUser'],
            'profile': {},
        })

    @authenticate('admin', 'admin')
    def test_create_incident_manager_when_authenticated_as_an_admin_user(self, session):
        session.delete_user('Im', if_exists=True)
        user = {
            'userId': 'Im',
            'password': 'user',
            'roles': ['StandardUser', 'IncidentManager']
        }
        resp = session.request('POST', config.auth + '/user', json=user)

        status_code = resp.status_code
        self.assertEqual(201, status_code)
        self.assertDictEqual(resp.json(), {
            'userId': 'Im',
            'state': 'initialized',
            'roles': ['StandardUser', 'IncidentManager'],
            'profile': {}
        })

    @authenticate('admin', 'admin')
    def test_create_fleet_manager_when_authenticated_as_an_admin_user(self, session):
        session.delete_user('fm', if_exists=True)
        user = {
            'userId': 'fm',
            'password': 'user',
            'roles': ['StandardUser', 'FleetManager']
        }
        resp = session.request('POST', config.auth + '/user', json=user)

        status_code = resp.status_code
        self.assertEqual(201, status_code)
        self.assertDictEqual(resp.json(), {
            'userId': 'fm',
            'state': 'initialized',
            'roles': ['StandardUser', 'FleetManager'],
            'profile': {}
        })

    @authenticate('admin', 'admin')
    def test_create_user_manager_when_authenticated_as_an_admin_user(self, session):
        session.delete_user('user', if_exists=True)
        user = {
            'userId': 'user',
            'password': 'mgr',
            'roles': ['StandardUser', 'UserManager']
        }
        resp = session.request('POST', config.auth + '/user', json=user)

        status_code = resp.status_code
        self.assertEqual(201, status_code)
        self.assertDictEqual(resp.json(), {
            'userId': 'user',
            'state': 'initialized',
            'roles': ['StandardUser', 'UserManager'],
            'profile': {}
        })

if __name__ == '__main__':
    unittest.main()
