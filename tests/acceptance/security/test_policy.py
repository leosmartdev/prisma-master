import config
import json
import unittest
from utilities.session import authenticate


class Policy(unittest.TestCase):

    @authenticate('admin', 'admin')
    def setUp(self, session):
        pass

    @authenticate('admin', 'admin')
    def tearDown(self, session):
        pass

    @authenticate("admin", "admin")
    def test_complex_password(self, session):
        pass

    @authenticate("admin", "admin")
    def test_reusing_old_password(self, session):
        pass

    @authenticate('admin', 'admin')
    def test_logging_in_with_your_updated_password(self, session):
        pass

if __name__ == '__main__':
    unittest.main()
