import config
import json
import unittest
from utilities.session import authenticate


class AdminUserCapabilities(unittest.TestCase):

    def get_actions(self, perms, classId):
        for perm in perms:
            if perm["classId"] == classId:
                return perm["actions"]
        self.fail("unable to find permissions with this classification " + classId)

    @authenticate('admin', 'admin')
    def test_validate_state(self, session):
        self.assertEqual("admin", session.user["user"]["userId"])
        self.assertEqual("activated", session.user["user"]["state"])

    @authenticate('admin', 'admin')
    def test_user_permission(self, session):
        self.assertEqual("admin", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "User")
        self.assertListEqual(
            actions, ['CREATE', 'READ', 'UPDATE', 'DELETE', 'UPDATE_ROLE', 'DEACTIVATE'])

    @authenticate('admin', 'admin')
    def test_fleet_permission(self, session):
        self.assertEqual("admin", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Fleet")
        self.assertListEqual(
            actions, ['CREATE', 'READ', 'UPDATE', 'DELETE', 'ADD_VESSEL', 'REMOVE_VESSEL'])

    @authenticate('admin', 'admin')
    def test_rule_permission(self, session):
        self.assertEqual("admin", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Rule")
        self.assertListEqual(
            actions, ['CREATE', 'READ', 'UPDATE', 'DELETE', 'STATE'])

    @authenticate('admin', 'admin')
    def test_vessel_permission(self, session):
        self.assertEqual("admin", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Vessel")
        self.assertListEqual(
            actions, ['CREATE', 'READ', 'UPDATE', 'DELETE'])

    @authenticate('admin', 'admin')
    def test_device_permission(self, session):
        self.assertEqual("admin", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Device")
        self.assertListEqual(actions, ['CREATE', 'READ', 'UPDATE', 'DELETE'])

    @authenticate('admin', 'admin')
    def test_audit_permission(self, session):
        self.assertEqual("admin", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Audit")
        self.assertListEqual(actions, ['READ'])

    @authenticate('admin', 'admin')
    def test_incident_permission(self, session):
        self.assertEqual("admin", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Incident")
        self.assertListEqual(actions, ['CREATE', 'READ', 'UPDATE', 'DELETE', 'CLOSE', 'OPEN', 'ASSIGN',
                                       'UNASSIGN', 'ARCHIVE', 'ADD_NOTE', 'ADD_NOTE_FILE', 'ADD_NOTE_ENTITY', 'DELETE_NOTE', 'TRANSFER_SEND', 'TRANSFER_RECEIVE'])

    @authenticate('admin', 'admin')
    def test_profile_permission(self, session):
        self.assertEqual("admin", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Profile")
        self.assertListEqual(actions, ['READ', 'UPDATE', 'UPDATE_PASSWORD'])

    @authenticate('admin', 'admin')
    def test_zone_permission(self, session):
        self.assertEqual("admin", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Zone")
        self.assertListEqual(actions, ['CREATE', 'READ', 'UPDATE', 'DELETE'])

    @authenticate('admin', 'admin')
    def test_alert_permission(self, session):
        self.assertEqual("admin", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Notice")
        self.assertListEqual(actions, ['GET', 'ACK', 'TIMEOUT', 'ACK_ALL'])

    @authenticate('admin', 'admin')
    def test_file_permission(self, session):
        self.assertEqual("admin", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "File")
        self.assertListEqual(actions, ['CREATE', 'READ', 'DELETE'])

    @authenticate('admin', 'admin')
    def test_view_permission(self, session):
        self.assertEqual("admin", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "View")
        self.assertListEqual(actions, ['CREATE', 'STREAM'])

    @authenticate('admin', 'admin')
    def test_track_permission(self, session):
        self.assertEqual("admin", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Track")
        self.assertListEqual(actions, ['GET', 'UPDATE', 'DELETE'])

    @authenticate('admin', 'admin')
    def test_history_permission(self, session):
        self.assertEqual("admin", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "History")
        self.assertListEqual(actions, ['GET'])

    @authenticate('admin', 'admin')
    def test_registry_permission(self, session):
        self.assertEqual("admin", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Registry")
        self.assertListEqual(actions, ['GET'])

    @authenticate('admin', 'admin')
    def test_config_permission(self, session):
        self.assertEqual("admin", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Config")
        self.assertListEqual(actions, ['UPDATE'])

    @authenticate('admin', 'admin')
    def test_policy_permission(self, session):
        self.assertEqual("admin", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Policy")
        self.assertListEqual(actions, ['READ', 'UPDATE'])


if __name__ == '__main__':
    unittest.main()
