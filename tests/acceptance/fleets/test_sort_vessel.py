import config
import json
import requests
import unittest
from utilities.session import authenticate
from utilities import mongo


class SortVesselsByFleet(unittest.TestCase):

    @authenticate('admin', 'admin')
    def test_create_a_vessel_with_valid_input_returns_success(self, session):
        pass 

if __name__ == '__main__':
    unittest.main()
