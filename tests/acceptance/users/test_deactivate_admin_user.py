import config
import json
import requests
import unittest
from utilities import mongo
from utilities.session import authenticate


class DeactivateAdminUser(unittest.TestCase):

    @authenticate('admin', 'admin')
    def test_deacitvate_admin_user(self, session):
        response = session.post_auth('user', {
            "userId": "admintest",
            "password": "admintest",
            "roles": ["StandardUser", "Administrator"]
        })
        self.assertEqual(201, response.status_code)
        body = response.json()
        self.assertEqual('admintest', body['userId'])
        response = session.get_auth("user/admintest")
        self.assertEqual(200, response.status_code)
        body = response.json()
        self.assertEqual('admintest', body['userId'])
        response = session.delete_auth("user/admintest")
        self.assertEqual(200, response.status_code)
        body = response.json()
        user_id = body['userId']
        self.assertEqual(user_id, body['userId'])
        self.assertEqual(1000, body['state'])
        #remove document with admintest user
        usersCollection = mongo.get_aaa_collection('users')
        usersCollection.find_one_and_delete({'userId': user_id})
        #removeDocument = mongo.remove_aaa_document_from_collection('users', {'userId': user_id})
        response = session.get_auth("user")
        self.assertEqual(200, response.status_code)
        #self.assertNotIn('fleet', vessel, str(vessel))


if __name__ == '__main__':
    unittest.main()
