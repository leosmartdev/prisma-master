import config
import json
import requests
import unittest
from utilities import mongo
from utilities.session import authenticate


class OmnicomDistressAlert(unittest.TestCase):

    @authenticate('admin', 'admin')
    def test_omnicom_distress_alert_with_simulator_data(self, session):
        pass

if __name__ == '__main__':
    unittest.main()
