import unittest
from utilities.session import authenticate
from utilities import mongo


class SearchVessel(unittest.TestCase):

    @mongo.clean_collections_after('vessels')
    @authenticate('admin', 'admin')
    def test_search(self, session):
        pass


if __name__ == '__main__':
    unittest.main()
