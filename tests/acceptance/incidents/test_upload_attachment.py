import os
import config
import unittest
from utilities import mongo
from utilities.session import authenticate

class UploadAttachment(unittest.TestCase):

    @mongo.clean_collections_after('incidents')
    @authenticate('incident', 'manager')
    def test_upload_word_document_to_an_open_incident(self, session):
        incident = self.createIncident(session)
        filename1 = 'c2_uploadattachmentincident.docx'
        filename2 = self.uploadFile(session, filename1, incident['id'])
        self.assertTrue(filename1, filename2)

    @mongo.clean_collections_after('incidents')
    @authenticate('incident', 'manager')
    def test_upload_excel_spreadsheet_to_an_open_incident(self, session):
        incident = self.createIncident(session)
        filename1 = 'OmniCom Solar Requirements [ST]_31AUG2016 (1).xlsx'
        filename2 = self.uploadFile(session, filename1, incident['id'])
        self.assertTrue(filename1, filename2)

    @mongo.clean_collections_after('incidents')
    @authenticate('incident', 'manager')
    def test_upload_pdf_to_an_open_incident(self, session):
        incident = self.createIncident(session)
        filename1 = 'pdf.pdf'
        filename2 = self.uploadFile(session, filename1, incident['id'])
        self.assertTrue(filename1, filename2)

    @mongo.clean_collections_after('incidents')
    @authenticate('incident', 'manager')
    def test_upload_jpeg_to_an_open_incident(self, session):
        incident = self.createIncident(session)
        filename1 = 'happy-fingers.jpg'
        filename2 = self.uploadFile(session, filename1, incident['id'])
        self.assertTrue(filename1, filename2)

    @mongo.clean_collections_after('incidents')
    @authenticate('incident', 'manager')
    def test_upload_json_document_to_an_open_incident(self, session):
        incident = self.createIncident(session)
        filename1 = 'vessels_data.json'
        filename2 = self.uploadFile(session, filename1, incident['id'])
        self.assertTrue(filename1, filename2)

    @mongo.clean_collections_after('incidents')
    @authenticate('incident', 'manager')
    def test_upload_png_document_to_an_open_incident(self, session):
        incident = self.createIncident(session)
        filename1 = 'happyface.png'
        filename2 = self.uploadFile(session, filename1, incident['id'])
        self.assertTrue(filename1, filename2)

    def uploadFile(self, session, filename, incident_id):
        script_dir = os.path.dirname(__file__) # <-- absolute dir the script is in
        rel_path = "files"
        abs_file_path = os.path.join(script_dir, rel_path, filename)
        multipart_form_data = {
            'file': open(abs_file_path, 'rb'),
        }
        rsp = session.request('POST', config.tms+'/file', files=multipart_form_data)
        body = rsp.json()
        self.assertIn('id', body)
        attachment_id = body['id']

        rsp = session.post('incident/'+incident_id+'/log', {
        	"type":"a100",
        	"note":"!!!!",
        	"attachment":{ "id":attachment_id }
        })
        incident = rsp.json()
        incident_log_entry = incident['log'][1]
        filename2 = incident_log_entry['attachment']['name']
        return filename2

    def createIncident(self, session):
        rsp = session.post('incident', {
            "name": "sar",
            "assignee": "incident",
            "commander": "incident",
            "phase": 2,
            "type": "Unlawful"
        })
        self.assertEqual(201, rsp.status_code, rsp.text)
        body = rsp.json()
        self.assertIn('id', body)
        return body

if __name__ == '__main__':
    unittest.main()
