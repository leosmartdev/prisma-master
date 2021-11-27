import config
import json
import requests
import unittest
from utilities.session import authenticate


class UserManagerPermissions(unittest.TestCase):

    def get_actions(self, perms, classId):
        for perm in perms:
            if perm["classId"] == classId:
                return perm["actions"]
        self.fail("unable to find permissions with this classification " + classId)

    @authenticate('usermanager', 'one')
    def test_check_if_user_manager_has_an_activated_state_when_logged_in(self, session):
        self.assertEqual("usermanager", session.user["user"]["userId"])
        self.assertEqual('activated', session.user["user"]["state"])

    @authenticate('usermanager', 'one')
    def test_user_profile_privileges_for_a_user_manager(self, session):
        self.assertEqual('usermanager', session.user['user']['userId'])
        perms = session.user['permissions']
        actions = self.get_actions(perms, 'Profile')
        self.assertListEqual(actions, ['READ', 'UPDATE', 'UPDATE_PASSWORD'])

    @authenticate('usermanager', 'one')
    def test_(self, session):
        self.assertEqual('usermanager', session.user['user']['userId'])
        perms = session.user['permissions']
        actions = self.get_actions(perms, 'Zone')
        self.assertListEqual(actions, ['CREATE', 'READ', 'UPDATE', 'DELETE'])

    @authenticate('usermanager', 'one')
    def test_incident_permissions_for_a_user_manager(self, session):
        self.assertEqual('usermanager', session.user['user']['userId'])
        perms = session.user['permissions']
        actions = self.get_actions(perms, 'Incident')
        self.assertListEqual(actions, ['READ'])

    @authenticate('usermanager', 'one')
    def test_notice_permissions_for_user_manager(self, session):
        self.assertEqual('usermanager', session.user['user']['userId'])
        perms = session.user['permissions']
        actions = self.get_actions(perms, 'Notice')
        self.assertListEqual(actions, ['GET', 'ACK', 'TIMEOUT', 'ACK_ALL'])

    @authenticate('usermanager', 'one')
    def test_map_view_permissions_for_user_manager(self, session):
        self.assertEqual('usermanager', session.user['user']['userId'])
        perms = session.user['permissions']
        actions = self.get_actions(perms, 'View')
        self.assertListEqual(actions, ['CREATE', 'STREAM'])

    @authenticate('usermanager', 'one')
    def test_track_priveleges_for_user_manager(self, session):
        perms = session.user['permissions']
        actions = self.get_actions(perms, 'Track')
        self.assertListEqual(actions, ['GET', 'UPDATE', 'DELETE'])

if __name__ == '__main__':
    unittest.main()
