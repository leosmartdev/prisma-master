import config
import json
import requests
import unittest
from utilities import mongo
from utilities.session import authenticate


class AuditLog(unittest.TestCase):

    @authenticate('admin', 'admin')
    def test_check_audit_log_to_validate_fleet_was_created(self, session):
        pass

    @authenticate('admin', 'admin')
    def test_check_audit_log_to_validate_fleet_was_deleted(self, session):
        pass

    @authenticate('admin', 'admin')
    def test_check_audit_log_to_validate_incident_was_created(self, session):
        pass

    @authenticate('admin', 'admin')
    def test_check_audit_log_to_validate_incident_was_deleted(self, session):
        pass

if __name__ == '__main__':
    unittest.main()
