"""
Tests incident forwarding using multicast and websocket connections.

NOTE: This is the first of the async tests that tests the websocket connections, so some of the
code in here will definitely need to be migrated into common utility code, but it's a proof of
concept just to get things working for now.
"""
import asyncio
import logging  # Uncomment to see websocket logging on stdout.
import unittest
from datetime import datetime

from incidents.test_upload_attachment import UploadAttachment
from prisma.tms.envelope.envelope_pb2 import Envelope
from utilities.session import authenticate

# Uncomment to see websocket logging on stdout.
logger = logging.getLogger('websockets')
logger.setLevel(logging.DEBUG)
logger.addHandler(logging.StreamHandler())


def elapsedTime(startTime):
    """ Returns the timedelta between the start time and current time."""
    time_delta = datetime.now() - startTime
    return time_delta.total_seconds()


class TestIncidentForwardingMulticast(unittest.TestCase):
    def setUp(self):
        """ create the site and incident to start tests with """
        self.incident = {}
        self.site = self.get_site(2)
        self.create_incident(name='Incident to Forward')

    @authenticate('incident', 'manager')
    def create_incident(self, session, name="Incident", phase=2, incidentType="Unknown"):
        """
        Creates an incident to forward in the system. Uses POST /incidents to create.

        @param {utils.session.Session} session Session provided from @authenticate decorator.
        @keyword {string} name Name of the incident
        @keyword {string} incidentType Type incident
        @keyword {int} phase Int version of the phase enum for the incident.
        @return {dictionary} The incident that was created in the system.
        @throws self.fail If the incident was not created.
        """
        response = session.post('incident', {
            "type": incidentType,
            "phase": phase,
            "commander": "incident",
            "assignee": "incident",
            "name": name,
        })

        self.incident = response.json()
        if response.status_code != 201:
            self.fail("Unable to create test incident for tests. STATUS={0}, RESPONSE={1}".format(response.status_code,
                                                                                                  self.incident))
        # upload files
        uploader = UploadAttachment()
        filename1 = 'c2_uploadattachmentincident.docx'
        filename2 = uploader.uploadFile(session, filename1, self.incident['id'])
        self.assertTrue(filename1, filename2)
        filename1 = 'OmniCom Solar Requirements [ST]_31AUG2016 (1).xlsx'
        filename2 = uploader.uploadFile(session, filename1, self.incident['id'])
        self.assertTrue(filename1, filename2)
        filename1 = 'pdf.pdf'
        filename2 = uploader.uploadFile(session, filename1, self.incident['id'])
        self.assertTrue(filename1, filename2)
        filename1 = 'happy-fingers.jpg'
        filename2 = uploader.uploadFile(session, filename1, self.incident['id'])
        self.assertTrue(filename1, filename2)
        filename1 = 'vessels_data.json'
        filename2 = uploader.uploadFile(session, filename1, self.incident['id'])
        self.assertTrue(filename1, filename2)
        filename1 = 'happyface.png'
        filename2 = uploader.uploadFile(session, filename1, self.incident['id'])
        self.assertTrue(filename1, filename2)
        return self.incident

    @authenticate('admin', 'admin')
    def get_site(self, session, siteId=2):
        response = session.get('site')
        self.assertEqual(200, response.status_code, str(response))
        sites = response.json()
        for site in sites:
            if siteId == site['siteId']:
                return site
        self.create_site()

    @authenticate('admin', 'admin')
    def create_site(self, session, name="Site", siteId=2):
        """
        Creates a site in the system to forward an incident to. Uses POST /sites to create.

        @param {utils.session.Session} session Session provided from @authenticate decorator.
        @keyword {string} name Name of the site
        @keyword {int} sideId Site ID to use for the site.
        @return {dictionary} The site that was created in the system.
        @throws self.fail If the site was not created.
        """
        response = session.post('site', {
            "name": name,
            "siteId": siteId
        })

        site = response.json()
        if response.status_code == 400:
            return  # already there
        if response.status_code != 201:
            self.fail("Unable to create test site for tests. STATUS={0}, RESPONSE={1}".format(response.status_code,
                                                                                              site))
        return site

    # @mongo.clean_collections_after('multicasts', 'transmissions')
    @authenticate('admin', 'admin')
    def test_multicast_incident_forward_with_no_log_entries_when_two_sites_pass(self, session):
        """
        Tests basic incident forwarding (with no log entries) skips correctly when the destination
        site is not running.
        """
        if self.site == None or 'connectionStatus' not in self.site or self.site['connectionStatus'] != "Ok":
            return unittest.skip("multiple site configuration required")

        # Everything needs to be run inside a coroutine so that the asyncio
        # loop works (and because unittest wont run async code directly)
        async def run_test():
            async with session.websocket_connect('/') as websocket:
                # Send multicast request.
                response = session.post('multicast/site/{0}'.format(self.site['id']), {
                    "payload": {
                        "@type": 'prisma.tms.moc.Incident',
                        "id": self.incident['id'],
                    }
                })

                multicast = response.json()
                self.assertEqual(response.status_code, 201,
                                 "Multicast config POST failed with incorrect status code. RESPONSE: {0}".format(
                                     multicast))
                self.assertEqual(len(multicast['transmissions']), 1,
                                 "Multicast response should contain exactly 1 transmission message")

                transmissionId = multicast['transmissions'][0]['id']
                transmissionCount = 0
                expected_transmission_count = 17
                incidentUpdateCount = 0
                startTime = datetime.now()

                # Read all websocket responses and find the transmission we want. we are expecting many
                # transmissions, so if we don't get those within about 10 seconds fail.
                while transmissionCount < expected_transmission_count and elapsedTime(startTime) < 10:
                    message = await asyncio.wait_for(websocket.recv(), timeout=20)
                    envelope = Envelope()  # parse message. It's a serialized protobuf
                    envelope.ParseFromString(message)
                    logger.info(envelope)
                    # check incident update
                    if envelope.type == 'Incident/CLOSE' and envelope.incident.id == self.incident['id']:
                        incidentUpdateCount += 1

                    # we only care about transmission update messages
                    if envelope.type == 'Transmission/UPDATE' and envelope.transmission.parentId == multicast['id']:
                        transmission = envelope.transmission
                        transmissionCount += 1
                        logger.info(transmissionCount)
                        # transmission asserts
                        self.assertEqual(transmission.id, transmissionId, 'Transmission ID is required.')
                        self.assertEqual(transmission.destination.type, 'prisma.tms.moc.Site',
                                         'Transmission destination is expected to be a site object')
                        self.assertEqual(transmission.destination.id, self.site['id'],
                                         'Transmission destination id is expected to be the site id')
                        self.assertEqual(transmission.parentId, multicast['id'],
                                         'Transmission parentId is expected to be the multicast ID')
                        # First transmission received
                        if transmissionCount == 1:
                            self.assertEqual(transmission.state, 1,
                                             'Expect first transmission to be Pending (STATE==1)')
                        # Second transmission received
                        elif transmissionCount == 2:
                            self.assertEqual(transmission.state, 4,
                                             'Expect second transmission to be Partial (STATE==4)')
                            self.assertEqual(transmission.status.code, 211,
                                             'Expect second transmission status object to have 211 error code')
                            self.assertEqual(transmission.status.message, 'transmission sent')
                        # Last transmission received
                        elif transmissionCount == expected_transmission_count:
                            self.assertEqual(transmission.state, 2,
                                             'Expect final transmission to be Success (STATE==2)')
                            self.assertEqual(transmission.status.code, 200,
                                             'Expect final transmission status object to have 200 error code')
                            self.assertEqual(transmission.status.message, 'transmission successful')
                self.assertEqual(transmissionCount, expected_transmission_count,
                                 "Expected transmissions over the socket.")
                self.assertEqual(incidentUpdateCount, 1,
                                 "Expected exactly 1 incident update over the socket.")

        # Run the actual test
        asyncio.get_event_loop().run_until_complete(run_test())


if __name__ == "__main__":
    unittest.main()
