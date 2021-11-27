import config
import json
import requests
import unittest
from utilities.session import authenticate


class IridiumMessages(unittest.TestCase):

    @authenticate('admin', 'admin')
    def test_error_message_in_activity_request_00x0(self, session):
        pass

    @authenticate('admin', 'admin')
    def test_known_error_messages_from_iridium_3G(self, session):
        pass

    @authenticate('admin', 'admin')
    def test_unknown_error_messages_from_iridium_3G(self, session):
        pass

if __name__ == '__main__':
    unittest.main()