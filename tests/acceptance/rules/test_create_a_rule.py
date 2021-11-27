import config
import json
import requests
import unittest
from utilities import mongo
from utilities.session import authenticate


class CreateRule(unittest.TestCase):

    @authenticate('admin', 'admin')
    def test_create_a_rule_with_valid_input(self, session):
        pass

    @authenticate('admin', 'admin')
    def test_create_a_rule_with_invalid_input(self, session):
        pass

    @authenticate('admin', 'admin')
    def test_list_all_rules_created(self, session):
        pass

    @authenticate('admin', 'admin')
    def test_retrieve_sepcific_rule_that_is_created_in_the_system(self, session):
        pass

    @authenticate('admin', 'admin')
    def test_update_a_rule_that_is_already_in_the_system(self, session):
        pass

    @authenticate('admin', 'admin')
    def test_update_a_rule_that_is_not_in_the_system(self, session):
        pass

    @authenticate('admin', 'admin')
    def test_delete_a_rule_that_is_already_in_the_system(self, session):
        pass

    @authenticate('admin', 'admin')
    def test_delete_a_rule_that_is_not_in_the_system(self, session):
        pass

if __name__ == '__main__':
    unittest.main()
