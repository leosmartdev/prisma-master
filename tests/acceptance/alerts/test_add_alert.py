import unittest, config, time
from utilities import mongo
from utilities.session import authenticate


class AddAlert(unittest.TestCase):

    @mongo.clean_collections_after('zones', 'notices')
    @authenticate('admin', 'admin')
    def test_add_alert_to_the_system_with_valid_input(self, session):
        # create a zone
        rsp = session.post('zone', {
            "name": "Zone",
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
                            },
                            {
                                "latitude": 0.6234618171050528,
                                "longitude": 104.54383850097658
                            },
                            {
                                "latitude": 1.256460562411604,
                                "longitude": 104.54109191894531
                            },
                            {
                                "latitude": 1.2530281581017135,
                                "longitude": 103.79642486572263
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
        # get all new notices
        rsp = session.get('notice/new?limit=100')
        self.assertEqual(200, rsp.status_code, 'I should get a 200 response code')
        body = rsp.json()
        result = False
        imei1 = '999234010030455'

        for v in body:
            isOmnicom = ('target' in v) and ('type' in v['target']) and ('registry_id' in v['target']) and (v['target']['type'] == 'OmnicomVMS')
            if not isOmnicom:
                continue

            registry_id = v['target']['registry_id']
            # get IMEI
            rsp = session.get('registry/%s' % registry_id)
            registry = rsp.json()
            self.assertEqual(200, rsp.status_code, rsp.text)
            self.assertIn('target', registry)
            self.assertIn('imei', registry['target'])
            imei2 = registry['target']['imei']
            if imei1 == imei2:
                result = True
                break

        self.assertTrue(result, 'OmnicomVMS IMEI(%s) not found in Notices(%s)' % (imei1, body))

if __name__ == '__main__':
    unittest.main()
