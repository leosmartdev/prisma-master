import config
import json
import requests
import unittest
from utilities.session import authenticate


class FleetManagerPermissions(unittest.TestCase):

    def get_actions(self, perms, classId):
        for perm in perms:
            if perm["classId"] == classId:
                return perm["actions"]
        self.fail("unable to find permissions with this classification " + classId)

    @authenticate('fleetmanager', 'one')
    def test_check_active_state_for_a_fleet_manager_on_login(self, session):
        self.assertEqual("fleetmanager", session.user["user"]["userId"])
        self.assertEqual("activated", session.user["user"]["state"])

    @authenticate('fleetmanager', 'one')
    def test_fleet_permission(self, session):
        self.assertEqual("fleetmanager", session.user["user"]["userId"])
        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Fleet")
        self.assertListEqual(actions, ['CREATE', 'READ', 'UPDATE', 'DELETE', 'ADD_VESSEL', 'REMOVE_VESSEL'])

    @authenticate('fleetmanager', 'one')
    def test_vessel_permission_for_a_fleet_manager(self, session):
        self.assertEqual("fleetmanager", session.user["user"]["userId"])
        perms = session.user["permissions"]
        actions = self.get_actions(perms, "Vessel")
        self.assertListEqual(actions, ['CREATE', 'READ', 'UPDATE', 'DELETE'])


if __name__ == '__main__':
    unittest.main()
