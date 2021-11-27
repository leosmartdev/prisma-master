# TCleaner mongo

tcleaner_mongo.py is used to flush mongodb collections to reduce disk usage.

## Usage
```
tcleaner_mongo.py --collections <collection names> <options>
```

`tcleaner_mongo.py` takes a list of collections and removes data from those collections. This is a
developer utility to help in maintinaing servers so they do not fill up disk space.

For production systems, it is expected that specific removal and archival setup will be put in place
to handle data as required for that installation.

This script can take an optional flag `--older-than` so you can mark only data written to the db
before a specified time instead of clearing the entire collection. This is useful as you can run
the script on a cron job and consistently remove data after it has been in the system for a given
amount of time.

## Flags

* `--collections` This is required. Collections takes a comma separated list of collection names to remove data from. `--collections incidents,tracks`
* `--mongodb-url` Specify the connection URL to mongo. Takes the full url: `mongodb://127.0.0.1:27017`
* `--older-than` Specify a time period where data older than that period will be deleted. `--older-than 3m`
    - Valid time values:
        * `<num>d` Days. Eg. `7d`
        * `<num>h` Hours. Eg. `24h`
        * `<num>m` Minutes. Eg. `60m`
        * `<num>s` Seconds. Eg. `60s`

## Examples

See help and options for using the script.
```
tcleaner_mongo.py --help
```

You can pass different mongodb ip:port
```
tcleaner_mongo.py --collections tracks --mongodb-url mongodb://8.8.8.8:27017/
```

You can pass different mongodb domain:port
```
tcleaner_mongo.py --collections tracks --mongodb-url mongodb://mydomains.ru:27017/
```

You must pass --collections option
```
tcleaner_mongo.py --collections tracks,zones
```

Point out "older-than" to set time how long ago since now records should be removed from the collections
```
tcleaner_mongo.py --collections tracks,zones --older-than 3m
```


## Additional information
The script has a dictionary called "collection_utime_field" you can pass a collection name as a key and a dictionary as value to provide maintaining a collection.
Where the value should have a key name - it is a name of time field, format is a lambda to get time.

The script has a dictionary called "time_shortname" you can pass different short name for time to pass value to datetime.
