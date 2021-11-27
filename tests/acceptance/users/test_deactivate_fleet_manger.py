import config
import json
import requests
import unittest
from utilities import mongo
from utilities.session import authenticate


class DeactivateFleetManager(unittest.TestCase):

    @authenticate('admin', 'admin')
    def test_deacitvate_fleet_manager(self, session):
        pass


if __name__ == '__main__':
    unittest.main()
