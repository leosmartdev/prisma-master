#!/usr/bin/python3

"""
 tcleaner_mongo is used to truncate collections in mongodb for test environments
"""

import argparse
import datetime
import os
import shutil

from pymongo import MongoClient

collection_utime_field = {
    'activity': {'name': 'time', 'format': lambda dt: dt.time().microsecond},
    'config': None,
    'devices': {'name': 'time', 'format': lambda dt: dt},
    'fleets': None,
    'incidents': {'name': 'utime', 'format': lambda dt: dt},
    'multicasts': None,
    'notices': {'name': 'utime', 'format': lambda dt: dt},
    'referenceSequence': None,
    'registry': {'name': 'me.target.time.seconds', 'format': lambda dt: dt.time().second},
    'request': None,
    'sites': None,
    'trackex': {'name': 'Updated', 'format': lambda dt: dt},
    'tracks': {'name': 'update_time', 'format': lambda dt: dt},
    'transmissions': None,
    'vessels': None,
    'zones': {'name': 'utime', 'format': lambda dt: dt},
}

time_shortname = {
    'd': 'days',
    's': 'seconds',
    'h': 'hours',
    'm': 'minutes',
}

args = argparse.ArgumentParser()
args.add_argument('--mongodb-url', default='mongodb://localhost:8201/', help='mongo db url')
args.add_argument('--collections', help='separated collections by "," for removing')
args.add_argument('--older-than', default='', help='how old data should be removed')


def main():
    arg_values = args.parse_args()
    client = MongoClient(arg_values.mongodb_url)
    print('gonna remove collection: [%s]' % arg_values.collections)
    since_time = None
    if arg_values.older_than:
        try:
            time_options = {
                time_shortname[arg_values.older_than[-1:]]: int(arg_values.older_than[:-1])
            }
            since_time = datetime.datetime.now() - datetime.timedelta(**time_options)
        except TypeError as err:
            print('error setting older-than argument %s' % err)
            return
    db_trident = client.trident
    for collection in arg_values.collections.split(','):
        del_query = {}
        if collection_utime_field.get(collection) is not None and since_time is not None:
            del_query[collection_utime_field.get(collection)['name']] = \
                {"$gte": collection_utime_field.get(collection)['format'](since_time)}
        result = db_trident[collection].delete_many(del_query)
        print(collection, del_query, result.deleted_count)
    client.close()
    try:
        shutil.rmtree('/srv/ftp')
    finally:
        os.mkdir('/srv/ftp')


if __name__ == '__main__':
    main()
