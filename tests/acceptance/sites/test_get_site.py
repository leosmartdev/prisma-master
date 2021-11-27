import unittest, config, time
from utilities import mongo
from utilities.session import authenticate
from .site_data import data

class GetSite(unittest.TestCase):

    @mongo.insert_collection_data('sites', data)
    @authenticate('admin', 'admin')
    def test_read_site_using_id_or_siteid(self, session):

        rsp = session.get('site/27')
    
        self.assertEqual(200, rsp.status_code,"Unable to read test site for tests. STATUS={0}, RESPONSE={1}".format(rsp.status_code,rsp.json()))
        
        site1 = rsp.json()

        rsp = session.get('site/' + site1["id"])

        self.assertEqual(200, rsp.status_code,"Unable to read test site using hexid for tests. STATUS={0}, RESPONSE={1}".format(rsp.status_code,rsp.json()))
        
        site2 = rsp.json()

        self.assertEqual(27, site2["siteId"],"Unable to read the correct site using hex id. Site={0}". format(site2))

        self.assertEqual(site1, site2, "site 1 and 2 should be the same. {0}!={1}".format(site1,site2))



if __name__ == '__main__':
    unittest.main()