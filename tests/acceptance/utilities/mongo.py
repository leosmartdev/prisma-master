from pymongo import MongoClient
import config
import requests


def connect():
    return MongoClient("mongodb://{host}:{port}".format(
        host=config.MONGO_HOST,
        port=config.MONGO_PORT,
    ))

def get_trident_collection(collection):
    """
    """
    client = connect()
    db = client[config.MONGO_DB_TRIDENT]
    return db[collection]

def get_aaa_collection(collection):
    """
    """
    client = connect()
    db = client[config.MONGO_DB_AAA]
    return db[collection]

def clear_collections(*collections):
    client = connect()
    db = client[config.MONGO_DB_TRIDENT]
    for collection in collections:
        db[collection].delete_many({})
    client.close()

"""
def remove_document_from_trident_collection(collection, query):
    client = connect()
    db = client[config.MONGO_DB_TRIDENT]
    find = db[collection].find_one_and_delete(query)
    return find
"""

class clean_collections_after:
    """
    Usage:
        @clean_collections_after('collection1', 'collection2', ...)

    This decorator will take a list of collection names and
    clear those collections using mongo.delete_many() when the
    wrapped function returns.

    Example:
     * Clear fleets collection
        @clean_collections_after('fleets')

     * clear fleets and vessels
        @clean_collections_after('fleets', 'vessels')

    """

    def __init__(self, *args):
        self.collections = args

    def __call__(self, func):
        def wrapped(*args, **kwargs):
            client = connect()
            ret = None
            try:
                ret = func(*args, **kwargs)
            finally:
                db = client[config.MONGO_DB_TRIDENT]
                for collection in self.collections:
                    db[collection].delete_many({})
                client.close()
            return ret
        return wrapped


class insert_collection_data:
    """
    Usage:
        @clean_collections_after('collection1', 'collection2', ...)

    This decorator will take a list of collection names and
    clear those collections using mongo.delete_many() when the
    wrapped function returns.

    Example:
     * Clear fleets collection
        @clean_collections_after('fleets')

     * clear fleets and vessels
        @clean_collections_after('fleets', 'vessels')

    """

    def __init__(self, collection, data):
        self.collection = collection
        self.data = data

    def __call__(self, func):
        def wrapped(*args, **kwargs):
            client = connect()
            db = client[config.MONGO_DB_TRIDENT]
            db[self.collection].delete_many({})
            db[self.collection].insert_many(self.data)
            ret = None
            try:
                ret = func(*args, **kwargs)
            finally:
                db[self.collection].delete_many({})
                client.close()
            return ret
        return wrapped
