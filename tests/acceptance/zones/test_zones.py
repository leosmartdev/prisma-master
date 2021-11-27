import config
import json
import requests
import unittest
from utilities.session import authenticate


class CreateZoneWithAdminPermissions(unittest.TestCase):

    @authenticate('admin', 'admin')
    def test_with_valid_input_returns_proper_response(self, session):
        pass

if __name__ == '__main__':
    unittest.main()
