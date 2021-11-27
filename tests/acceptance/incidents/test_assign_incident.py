import config
import json
import requests
import unittest
from utilities.session import authenticate

class AssignIncident(unittest.TestCase):

    @authenticate('admin', 'admin')
    def test_assign_incident_to_incident_manager(self, session):
        pass

    @authenticate('admin', 'admin')
    def test_assign_incident_to_user_with_invalid_permissions(self, session):
        pass

    @authenticate('admin', 'admin')
    def test_assign_incident_and_leave_assignee_field_blank(self, session):
        pass

if __name__ == '__main__':
    unittest.main()
