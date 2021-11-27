from multiprocessing import Process

import requests
import config

from utilities import session

session = session.Session()
rsp = session.request('GET', config.config)
session.close()
body = rsp.json()

url_request = body['service']['tms']['request']
url_activity = body['service']['tms']['activity']

activity_message = '{"registryId":"aa6cdba6a484dc775a70cd17c155de8b","activityId":"a88b8051fb6c21b4ccadc1222eea5961","requestId":"1","body":{"@type":"type.googleapis.com/prisma.tms.iridium.Iridium","mtc":{"IEI":"RA==","MTCL":25,"UCMIF":1127362609,"IMEI":"111234010030457","MTMessageStatus":-2}},"time":"2018-01-29T14:25:28.329708909Z","type":"OmnicomVMS"}'

request_message_history = '{"registryId":"100","requestId":"2","cmd":{"@type":"type.googleapis.com/prisma.tms.omnicom.Omni","rmh":{"Header":"MQ==","Date":{"Year":18,"Month":1,"Day":24,"Minute":809},"DateInterval":{"Start":{"Year":18,"Month":1,"Day":24,"Minute":749},"Stop":{"Year":18,"Month":1,"Day":24,"Minute":808}},"IDMsg":1992}},"destinations":[{"type":"IMEI","value":"119234910030457"}],"time":"2018-01-24T13:29:16.330456048Z","type":"OmnicomVMS"}'


# get_activities will listen on a activity stream, cache all the activity message during a limited amount of time.
# it return a json array with activity objects. for more info on activity strcture lookup activity_request.proto
def get_activities(url):
    response = requests.get(url)
    print("received answer for get_activities")
    print(response.status_code)
    print(response.text)


# get_requests will listen on a request stream, cache all the request messages during a limited amount of time.
# it return a json array with request objects. for more info on requests structure lookup activity_request.proto
def get_requests(url):
    response = requests.get(url)
    print(response.status_code)
    print(response.text)


# get_activity takes an endpoint and request_id, it will listen on a request stream for a specific request to show up.
# it will return a json request object or timeout.
def get_activity(url, Id):
    response = requests.get(url + "/" + Id)
    print(response.status_code)
    print(response.text)


# get_request takes an endpoint and request_id,
# it will listen on an activity stream for a specific activity with the associated request id specified to show up.
# it will return a json activity object or timeout.
def get_request(url, Id):
    response = requests.get(url + "/" + Id)
    print(response.status_code)
    print(response.text)


# post_request takes and endpoint and a json formated request object
# that will be sent to a specific daemon configured in tporald.
def post_requests(url, msg):
    print(msg)
    print("sending request...")
    response = requests.post(url, data=msg)
    print(response.status_code)


# post_activity takes and endpoint and a json formated activity object
# that will be sent to a specific daemon configured in tportald.
def post_activity(url, msg):
    response = requests.post(url, data=msg)
    print(response.status_code)


if __name__ == '__main__':
    process_activity = Process(target=get_activities, args=(url_activity,))
    process_request = Process(target=post_requests, args=(url_request, request_message_history))
    process_activity.start()
    process_request.start()
    process_activity.join()
    process_request.join()
