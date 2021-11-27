import config
import json
import requests
import unittest
from utilities import mongo
from utilities.session import authenticate


class AuditLogEdgeCases(unittest.TestCase):

    @authenticate('admin', 'admin')
    def test_that_you_are_unable_to_write_to_the_audit_log(self, session):
        response = session.post_auth('audit', {})
        self.assertEqual(405, response.status_code)

    @authenticate('admin', 'admin')
    def test_that_you_are_unable_delete_the_audit_log(self, session):
        response = session.delete_auth('audit', {})
        self.assertEqual(405, response.status_code)

    @authenticate('admin', 'admin')
    def test_that_you_are_unable_update_the_audit_log(self, session):
        response = session.put_auth('audit', {
            'id': '5a40492e3055b43bbecb7d18'
        })
        self.assertEqual(405, response.status_code)

if __name__ == '__main__':
    unittest.main()
