import logging, json, os, sys, requests, base64, boto3

from urllib.request import Request, urlopen
from urllib.error import URLError, HTTPError
# The Slack channel to send a message to stored in the slackChannel environment variable
HOOK_URL = 'https://hooks.slack.com/services/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'
channel_id = 'conv-dev-aws'
encrypedToken = 'AQICAHjAJPXeFo/LNgYEcN07kmlSTAbbNkm+rp4448bo6E6QkQHN6P0S0UC+vIMwfWR15GqhAAAAljCBkwYJKoZIhvcNAQcGoIGFMIGCAgEAMH0GCSqGSIb3DQEHATAeBglghkgBZQMEAS4wEQQMFBEW6tNQz+k83EU6AgEQgFAzXLXnwfe4EAVi4pyQTVs5NjkEEta6+rX7EPVtSgbGtlopWLI63RUwpgbkcYPjANILi6dnX2Zxfxdyumo00RxtrDp3MM2Xf2l5P0J0IG80yQ=='
region_name = 'us-east-1'
# slackUrlPostMessage = 'https://slack.com/api/chat.postMessage'
slackUrlFilesUpload = "https://slack.com/api/files.upload"

logger = logging.getLogger()
logger.setLevel(logging.INFO)

def encrypt(client, secret, alias):
    ciphertext = client.encrypt(
        KeyId=alias,
        Plaintext=bytes(secret, 'utf8'),
    )
    return base64.b64encode(ciphertext["CiphertextBlob"])


def decrypt(client, secret):
    plaintext = client.decrypt(
        CiphertextBlob=bytes(base64.b64decode(secret))
    )
    return plaintext["Plaintext"]

def sendToSlackByHook():
    data = "data"
    slack_message = {
        'attachments': [{
            'fallback': 'eventText',
            'pretext': "Job ID:",
            'color': "#D00000",
            'fields': [{
                "value": data,
                "short": False
            }]
        }]
    }

    req = Request(HOOK_URL, json.dumps(slack_message).encode('utf-8'))
    try:
        response = urlopen(req)
        response.read()
        logger.info("Message posted to %s", 'channel')
    except HTTPError as e:
        logger.error("Request failed: %d %s", e.code, e.reason)
    except URLError as e:
        logger.error("Server connection failed: %s", e.reason)

def sendToSlack():
    # logger.info("Event: " + str(event))
    kms = boto3.client('kms', region_name=region_name)
    token = decrypt(kms, encrypedToken).decode('utf-8')
    # slackUrlPostMessageTC = slackUrlPostMessage + "?token={token}".format(token=token)
    aws_pipeline = sys.argv[3]
    if aws_pipeline == 'codepipeline/prisma-pc2-294':
        channel_id = 'conv-dev-aws'

    if sys.argv[2] == '0':
        initial_comment = ':heavy_check_mark: *' + aws_pipeline + '* Integration tests is ok.'
    else:
        initial_comment = ':x: *' + aws_pipeline + '* Integration tests failed.'
    # rsp = requests.post(slackUrlPostMessageTC, {'channel': channel_id, 'text': initial_comment})
    fh = open(sys.argv[1], "r")
    # data = fh.read()
    slackUrlFilesUploadTC = slackUrlFilesUpload + "?token={token}&channels={channel_id}&initial_comment={initial_comment}".format(token=token, channel_id=channel_id, initial_comment=initial_comment)
    rsp = requests.post(slackUrlFilesUploadTC, files={'file': fh})
    fh.close()
    # print(rsp.content)

def sendToSlack2():
    slack_token = sys.argv[1]
    kms = boto3.client('kms', region_name=region_name)
    secret = encrypt(kms, slack_token, 'alias/slack')
    print(secret.decode('utf-8'))

if __name__ == '__main__':
    sendToSlack()
