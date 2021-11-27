import config
import json
import requests
import unittest
from utilities.session import authenticate


class IncidentManagerPermissions(unittest.TestCase):

    def get_actions(self, perms, classId):
        for perm in perms:
            if perm["classId"] == classId:
                return perm["actions"]
        self.fail("unable to find permissions with this classification " + classId)

    @authenticate('incident', 'manager')
    def test_check_if_incident_manager_has_an_activated_state_when_logged_in(self, session):
        self.assertEqual("incident", session.user["user"]["userId"])
        self.assertEqual('activated', session.user["user"]["state"])

    @authenticate('incident', 'manager')
    def test_user_profile_privileges_for_incident_manager(self, session):
        self.assertEqual('incident', session.user['user']['userId'])
        perms = session.user['permissions']
        actions = self.get_actions(perms, 'Profile')
        self.assertListEqual(actions, ['READ', 'UPDATE', 'UPDATE_PASSWORD'])

    @authenticate('incident', 'manager')
    def test_zone_priveleges_for_an_incident_manager(self, session):
        self.assertEqual('incident', session.user['user']['userId'])
        perms = session.user['permissions']
        actions = self.get_actions(perms, 'Zone')
        self.assertListEqual(actions, ['CREATE', 'READ', 'UPDATE', 'DELETE'])

    @unittest.skip("defect: CONV-1797-classId: for incident needs to be de-duped")
    @authenticate('incident', 'manager')
    def test_incident_permissions_for_an_incident_manager(self, session):
        self.assertEqual('incident', session.user['user']['userId'])
        perms = session.user['permissions']
        actions = self.get_actions(perms, 'Incident')
        self.assertListEqual(actions, ['CREATE', 'READ', 'UPDATE', 'DELETE', 'OPEN', 'CLOSE', 'ASSIGN', 'UNASSIGN', 'ARCHIVE', 'ADD_NOTE', 'ADD_NOTE_FILE', 'ADD_NOTE_ENTITY', 'DELETE_NOTE'])

    @authenticate('incident', 'manager')
    def test_notice_permissions_for_an_incident_manager(self, session):
        self.assertEqual('incident', session.user['user']['userId'])
        perms = session.user['permissions']
        actions = self.get_actions(perms, 'Notice')
        self.assertListEqual(actions, ['GET', 'ACK', 'TIMEOUT', 'ACK_ALL'])

    @authenticate('incident', 'manager')
    def test_map_view_permissions_for_an_incident_manager(self, session):
        self.assertEqual('incident', session.user['user']['userId'])
        perms = session.user['permissions']
        actions = self.get_actions(perms, 'View')
        self.assertListEqual(actions, ['CREATE', 'STREAM'])

    @authenticate('incident', 'manager')
    def test_history_permissions_for_an_incident_manager(self, session):
        self.assertEqual('incident', session.user['user']['userId'])
        perms = session.user['permissions']
        actions = self.get_actions(perms, 'History')
        self.assertListEqual(actions, ['GET'])

    @authenticate('incident', 'manager')
    def test_registry_permissions_for_an_incident_manager(self, session):
        self.assertEqual('incident', session.user['user']['userId'])
        perms = session.user['permissions']
        actions = self.get_actions(perms, 'Registry')
        self.assertListEqual(actions, ['GET'])

if __name__ == '__main__':
    unittest.main()
