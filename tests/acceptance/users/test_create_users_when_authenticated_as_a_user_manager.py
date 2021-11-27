import config
import json
import requests
import unittest
from utilities import mongo
from utilities.session import authenticate


class CreateUsers(unittest.TestCase):

    @authenticate('usermanager', 'one')
    def test_create_standard_user_when_logged_in_as_a_user_manager(self, session):
        response = session.post_auth('user', {
            "userId": "standardone",
            "password": "password",
            "roles": ["StandardUser"]
        })
        self.assertEqual(201, response.status_code)
        body = response.json()
        self.assertEqual('standardone', body['userId'])
        self.assertEqual('initialized', body['state'])
        self.assertEqual(['StandardUser'], body['roles'])
        response = session.get_auth("user/standardone")
        self.assertEqual(200, response.status_code)
        body = response.json()
        self.assertEqual('standardone', body['userId'])
        usersCollection = mongo.get_aaa_collection('users')
        usersCollection.find_one_and_delete({'userId': 'standardone'})

    @authenticate('usermanager', 'one')
    def test_create_incident_manager_when_logged_in_as_a_user_manager(self, session):
        response = session.post_auth('user', {
            "userId": "inc",
            "password": "mgr",
            "roles": ['StandardUser', 'IncidentManager']
        })
        self.assertEqual(201, response.status_code)
        body = response.json()
        self.assertEqual('inc', body['userId'])
        self.assertEqual('initialized', body['state'])
        self.assertEqual(['StandardUser', 'IncidentManager'], body['roles'])
        response = session.get_auth("user/inc")
        self.assertEqual(200, response.status_code)
        body = response.json()
        usersCollection = mongo.get_aaa_collection('users')
        usersCollection.find_one_and_delete({'userId': 'inc'})

    @authenticate('usermanager', 'one')
    def test_create_fleet_manager_when_logged_in_as_a_user_manager(self, session):
        pass

    @authenticate('usermanager', 'one')
    def test_create_user_manager_when_logged_in_as_a_user_manager(self, session):
        pass

if __name__ == '__main__':
    unittest.main()
