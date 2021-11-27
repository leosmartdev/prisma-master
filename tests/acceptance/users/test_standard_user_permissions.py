import config
import json
import requests
import unittest
from utilities.session import authenticate


class StandardUserCapabilities(unittest.TestCase):

    @authenticate('admin', 'admin')
    def setUp(self, session):
        session.create_user('standard', 'user', ['StandardUser'])

    @authenticate('admin', 'admin')
    def tearDown(self, session):
        session.delete_user('standard')

    def get_actions(self, perms, classId):
        for perm in perms:
            if perm["classId"] == classId:
                return perm["actions"]
        self.fail("unable to find permissions with this classification " + classId)

    @authenticate('standard', 'user')
    def test_validate_state(self, session):
        self.assertEqual("standard", session.user["user"]["userId"])
        self.assertEqual("activated", session.user["user"]["state"])

    @authenticate('standard', 'user')
    def test_profile_permission(self, session):
        self.assertEqual("standard", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Profile")
        self.assertListEqual(actions, ['READ', 'UPDATE', 'UPDATE_PASSWORD'])

    @authenticate('standard', 'user')
    def test_zone_permission(self, session):
        self.assertEqual("standard", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Zone")
        self.assertListEqual(actions, ['CREATE', 'READ', 'UPDATE', 'DELETE'])

    @authenticate('standard', 'user')
    def test_incident_permission(self, session):
        self.assertEqual("standard", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Incident")
        self.assertListEqual(actions, ['READ'])

    @authenticate('standard', 'user')
    def test_history_permission(self, session):
        self.assertEqual("standard", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "History")
        self.assertListEqual(actions, ['GET'])

    @authenticate('standard', 'user')
    def test_registry_permission(self, session):
        self.assertEqual("standard", session.user["user"]["userId"])

        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Registry")
        self.assertListEqual(actions, ['GET'])

if __name__ == '__main__':
    unittest.main()
