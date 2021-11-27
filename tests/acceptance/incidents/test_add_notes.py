import config
import json
import requests
import unittest
from utilities.session import authenticate

class AddNotesToAnOpenIncident(unittest.TestCase):

    @authenticate('incident', 'manager')
    def test_add_a_note_to_an_open_incident(self, session):
        pass

if __name__ == '__main__':
    unittest.main()
