import unittest

from utilities.session import authenticate


class TestForbiddenResponse(unittest.TestCase):

    @authenticate('standard', 'user')
    def test_retrieve_audit_log_as_a_standard_user(self, session):
        pass


if __name__ == '__main__':
    unittest.main()
