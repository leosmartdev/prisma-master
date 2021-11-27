import json
import ssl

import requests
import websockets

import config


class Session(requests.Session):

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.user = None

    def login(self, name, password):
        response = self.request('POST', config.auth + '/session', data=json.dumps({
            'userName': name,
            'token': password
        }),
            headers={
            'Content-Type': 'application/json',
        })
        self.user, status_code = json.loads(
            response.text), response.status_code
        if response.status_code != 200:
            raise Exception("Unable to authenticate as the {} user".format(name))
        return self

    def create_user(self, user_id, password, roles):
        self.delete_user(user_id, if_exists=True)
        response = self.request('POST', config.auth + '/user', data=json.dumps({
            'userId': user_id,
            'password': password,
            'roles': roles
        }), headers={'Content-Type': 'application/json'})
        if response.status_code != 201:
            raise Exception("Unable to create user {}: {}".format(
                user_id, response.status_code))

    def delete_user(self, user_id, if_exists=False):
        response = self.request('DELETE',
                                config.auth + '/user/{}'.format(user_id),
                                headers={'Content-Type': 'application/json'})
        if not if_exists and response.status_code != 200:
            raise Exception("Unable to delete user {}".format(user_id))

    def request(self, method, url, **kwargs):
        return super().request(method, url, verify=False, **kwargs)

    def get(self, uri, *args, **kwargs):
        new_url = "{0}/{1}".format(config.tms, uri)
        return super().get(new_url, *args, headers=config.json, **kwargs)

    def get_auth(self, uri, *args, **kwargs):
        new_url = "{0}/{1}".format(config.auth, uri)
        return super().get(new_url, *args, headers=config.json, **kwargs)

    def get_simulator(self, uri, *args, **kwargs):
        newUrl = "{0}/{1}".format(config.simulator, uri)
        return super().get(newUrl, *args, headers=config.json, **kwargs)

    def post(self, uri, data, *args, **kwargs):
        new_url = "{0}/{1}".format(config.tms, uri)
        return super().post(new_url, json=data, headers=config.json, **kwargs)

    def post_auth(self, uri, data, *args, **kwargs):
        new_url = "{0}/{1}".format(config.auth, uri)
        return super().post(new_url, json=data, headers=config.json, **kwargs)

    def post_simulator(self, uri, data, *args, **kwargs):
        newUrl = "{0}/{1}".format(config.simulator, uri)
        return super().post(newUrl, json=data, headers=config.json, **kwargs)

    def put(self, uri, data, *args, **kwargs):
        new_url = "{0}/{1}".format(config.tms, uri)
        return super().put(new_url, json=data, headers=config.json, **kwargs)

    def put_auth(self, uri, data, *args, **kwargs):
        new_url = "{0}/{1}".format(config.auth, uri)
        return super().put(new_url, json=data, headers=config.json, **kwargs)

    def delete(self, uri, *args, **kwargs):
        new_url = "{0}/{1}".format(config.tms, uri)
        return super().delete(new_url, headers=config.json, **kwargs)

    def delete_auth(self, uri, *args, **kwargs):
        new_url = "{0}/{1}".format(config.auth, uri)
        return super().delete(new_url, headers=config.json, **kwargs)

    def websocket_connect(self, uri, *args, **kwargs):
        """
        Opens a websocket connection to the provided URI. Use this function in a `async with` block.
        For example:

            async with session.websocket_connect('/') as websocket:
                msg = websocket.recv()
                ...

        Base URL is provided in the config.py file, this URI should be only after the
        `wss://host:port/ws/v2`. Eg, for map socket, uri is `/view/stream` for
        `wss://localhost/ws/v2/view/stream` Accepts all args are kwargs for `websockets.connect()`
        and will pass them through. See http://websockets.readthedocs.io/en/stable/ for more info.

        @param string uri: The URI of the websocket. Must contain leading /
        """
        headers = {'Cookie': 'id={0}'.format(self.cookies.get_dict()['id'])}
        return websockets.connect('{0}{1}'.format(config.websocket_base, uri),
            ssl=get_ssl_context(),
            extra_headers=headers,
            timeout=0
        )


class authenticate:

    def __init__(self, username, password):
        self.username = username
        self.password = password

    def __call__(self, func):
        def wrapped(them, *args, **kwargs):
            session = Session().login(self.username, self.password)
            try:
                ret = func(them, session, *args, **kwargs)
            finally:
                session.close()
            return ret
        return wrapped


def get_ssl_context():
  """
  Returns an ssl context for use with a remote connection.
  This context turns off certificate validation and hostname checks.
  WARNING: ONLY TO BE USED FOR TESTING PURPOSES. This is incredibly insecure.

  This is mainly used to allow ssl to work with the websocket package.

  @return ssl.SSLContext The context to use.
  """
  ssl_ctx = ssl.create_default_context()
  ssl_ctx.check_hostname = False
  ssl_ctx.verify_mode = ssl.CERT_NONE

  return ssl_ctx
