import unittest
from utilities.session import authenticate

class GetGeoFences(unittest.TestCase):

    @authenticate('admin', 'admin')
    def test_crud(self, session):
        resp = session.get('geofence')
        self.assertEqual(200, resp.status_code, resp.text)
        length_answer = len(resp.json()) if resp.json() is not None else 0
        resp = session.post('geofence', {
            'database_id': '5a57734677f94b6c1ece8eb6'
        })
        self.assertEqual(201, resp.status_code, resp.text)
        resp = session.get('geofence')
        self.assertEqual(200, resp.status_code, resp.text)
        self.assertEqual(1, len([x for x in resp.json() if x.get('database_id') == '5a57734677f94b6c1ece8eb6']))
        # update
        resp = session.put('geofence/5a57734677f94b6c1ece8eb6', {})
        self.assertEqual(201, resp.status_code, resp.text)
        resp = session.put('geofence/0000000077f94b6c1ece8eb6', {})
        self.assertEqual(404, resp.status_code, resp.text)
        resp = session.put('geofence/1', {})
        self.assertEqual(400, resp.status_code, resp.text)

        resp = session.delete('geofence/5a57734677f94b6c1ece8eb6')
        self.assertEqual(200, resp.status_code, resp.text)
        resp = session.delete('geofence/5a57734677f94b6c1ece8eb6')
        self.assertEqual(404, resp.status_code, resp.text)

        # protect from invalid id's
        resp = session.delete('geofence/1')
        self.assertEqual(400, resp.status_code, resp.text)
        resp = session.post('geofence', {
            'database_id': '1'
        })
        self.assertEqual(400, resp.status_code, resp.text)


if __name__ == '__main__':
    unittest.main()
