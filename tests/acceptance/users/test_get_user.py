import config
import json
import requests
import unittest
from utilities.session import authenticate


class GetUser(unittest.TestCase):

    @authenticate('admin', 'admin')
    def test_get_a_user_that_is_not_in_the_system_when_authenticated_as_an_admin_user(self, session):
        user_id = '123'
        rsp = session.get_auth('user/' + user_id)
        self.assertEqual(404, rsp.status_code)
        body = rsp.json()
        self.assertTrue(isinstance(body, list), 'Expected an array.')
        self.assertEqual(1, len(body), 'Expected a array of 1.')
        self.assertDictEqual(body[0], {
            "property": "User",
            "rule": "NotFound",
            "message": "Not found"
        })

if __name__ == '__main__':
    unittest.main()
