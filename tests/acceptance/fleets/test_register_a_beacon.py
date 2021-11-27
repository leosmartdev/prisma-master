import unittest

from utilities.session import authenticate


class RegisterABeacon(unittest.TestCase):

    @authenticate('fleetmanager', 'one')
    def test_register_a_beacon_valid_input(self, session):
        pass

    @authenticate('fleetmanager', 'one')
    def test_register_a_beacon_with_invalid_input(self, session):
        pass

    @authenticate('usermanager', 'one')
    def test_register_a_beacon_with_invalid_permissions(self, session):
        pass

    @authenticate('fleetmanager', 'one')
    def test_remove_a_beacon_valid_input(self, session):
        pass

    @authenticate('usermanager', 'one')
    def test_remove_a_beacon_with_invalid_permissions(self, session):
        pass

    @authenticate('fleetmanager', 'one')
    def test_add_a_beacon_to_a_vessel_that_is_already_registered_to_another_vessel(self, session):
        pass

    @authenticate('fleetmanager', 'one')
    def test_broadcast_a_message_to_a_fleet(self, session):
        pass

    @authenticate('fleetmanager', 'one')
    def test_broadcast_a_message_to_a_fleet_with_valid_permissions(self, session):
        pass

    @authenticate('usermanager', 'one')
    def test_broadcast_a_message_to_a_fleet_with_invalid_permissions(self, session):
        pass


if __name__ == '__main__':
    unittest.main()
