import unittest

from utilities.session import authenticate


class GetDevices(unittest.TestCase):

    @authenticate('admin', 'admin')
    def test_get_all_devices_valid_success(self, session):
        rsp = session.get('device')
        self.assertEqual(200, rsp.status_code, 'I should get a 200 response code')


if __name__ == '__main__':
    unittest.main()
