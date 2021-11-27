import config
import json
import unittest
from utilities.session import authenticate
from utilities import mongo
from .data_search import data, fleet, vessels


class Search(unittest.TestCase):

    @mongo.insert_collection_data('registry', data)
    @authenticate('admin', 'admin')
    def test_search_imo(self, session):
        response = session.get("search/registry?query=9257888")
        self.assertEqual(200, response.status_code)
        body = response.json()
        self.assertEqual(1, len(body))
        self.assertEqual("WAVEMASTER 5", body[0]["label"])

    @mongo.insert_collection_data('registry', data)
    @authenticate('admin', 'admin')
    def test_search_mmsi(self, session):
        response = session.get("search/registry?query=63601")
        self.assertEqual(200, response.status_code)
        body = response.json()
        self.assertEqual(1, len(body))
        self.assertEqual("636011907", body[0]["label"])

    @mongo.insert_collection_data('registry', data)
    @authenticate("admin", "admin")
    def test_search_name_exact(self, session):
        response = session.get("/search/registry?query=WAVEMASTER 3")
        self.assertEqual(200, response.status_code)
        body = response.json()
        self.assertEqual(1, len(body))
        self.assertEqual("WAVEMASTER 3", body[0]["label"])

    @mongo.insert_collection_data('registry', data)
    @authenticate("admin", "admin")
    def test_search_name_fuzzy(self, session):
        response = session.get("/search/registry?query=WAVEMASTER")
        self.assertEqual(200, response.status_code)
        body = response.json()
        self.assertEqual(2, len(body))
        names = [body[0]["label"], body[1]["label"]]
        self.assertIn("WAVEMASTER 3", names)
        self.assertIn("WAVEMASTER 5", names)

    @mongo.insert_collection_data('fleets', fleet)
    @authenticate("admin", "admin")
    def test_search_fleets(self, session):
        resp = session.get("/search/tables/fleets?query=gold")
        self.assertEqual(200, resp.status_code, str(resp.json()))
        self.assertEqual('5a5df1d577f94b46842d8ade', resp.json()[0]['id'])

    @mongo.insert_collection_data('vessels', vessels)
    @authenticate("admin", "admin")
    def test_search_vessels(self, session):
        resp = session.get("/search/tables/vessels?query=zuby")
        self.assertEqual(200, resp.status_code)
        self.assertEqual('5a5df66077f94b5466787de6', resp.json()[0]['id'])

    @mongo.insert_collection_data('fleets', fleet)
    @authenticate("admin", "admin")
    def test_search_fleets_and_vessels(self, session):
        resp = session.get("/search/tables/fleets,vessels?query=gold")
        self.assertEqual(200, resp.status_code, str(resp.json()))
        self.assertEqual('5a5df1d577f94b46842d8ade', resp.json()[0]['id'])

if __name__ == '__main__':
    unittest.main()
