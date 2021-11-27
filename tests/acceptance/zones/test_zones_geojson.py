import time
import unittest
from utilities import mongo
from utilities.session import authenticate

class TestZonesGeoJson(unittest.TestCase):

    @mongo.clean_collections_after('zones', 'notices')
    @authenticate('admin', 'admin')
    def test_zones_geojson(self, session):
        name = "zone name"
        rsp = session.post('zone', {
            "name": name,
            "description": "",
            "poly": {
                "lines": [
                    {
                        "points": [
                            {
                                "latitude": 1.2530281581017135,
                                "longitude": 103.79642486572263
                            },
                            {
                                "latitude": 0.6234618171050528,
                                "longitude": 103.79711151123044
                            }
                        ]
                    }
                ]
            },
            "fill_color": {"r": 255, "g": 0, "b": 0, "a": 0.25},
            "fill_pattern": "solid",
            "stroke_color": {"r": 0, "g": 0, "b": 0},
            "create_alert_on_enter": True,
            "create_alert_on_exit": False
        })
        self.assertEqual(201, rsp.status_code, 'I should get a 201 response code')
        time.sleep(1)
        rsp = session.get('/zone/geo')
        self.assertEqual(200, rsp.status_code)
        body = rsp.json()
        self.assertIn('type', body)
        self.assertIn('polygons', body)
        self.assertIn('points', body)
        self.assertEqual(1, len(body['polygons']['features']))
        self.assertEqual(name, body['polygons']['features'][0]['properties']['name'])

if __name__ == '__main__':
    unittest.main()
