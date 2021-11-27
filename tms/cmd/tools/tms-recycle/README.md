tms-recycle(1) -- clean and restart tms environment
====================================================

## SYNOPSIS

`tms-recycle` [--rebuild] [--config=file]

## DESCRIPTION

The tms-recycle program clean and restarts tms environments.
It uses makefile to rebuild tms daemons.
It uses makefile to rebuild protobuf files.
By default it takes /etc/trident/tmsd.conf config to pass to tmsd daemon.
It removes all data from /var/trident/db/ to clean database, but before it
tries to stop mongodb service.
It uses redis-cli to clean redis db.
It sleeps to allow mongodb finishing on each steps.
The tms-recycle set up mongodb using js scripts which are places at /vagrant/etc/mongodb/
Scripts are:

    replication.js - setup replset
    trident.js - setup trident schema
    aaa.js - setup aaa schema

It also validates the schemas after the setups.
It runs tmsd in background using passed configs.

## OPTIONS

The following options are available

* `--config`=file:
passing different config files allows to run tmsd daemon with different config files
* `--rebuild`:
allows to rebuild protobuf files, daemons

## NOTE

The tms-recycle uses "case" to parse flags. In the future it can be replaced on get opts.
--config flag requires an absolute or a relative path to a config file.
You can use flags using any order.
It doesn't have auto completion.

## EXAMPLES

stop tmsd daemon. parse flags. clean redis. clean mongodb. setup mongodb. run tmsd with default config file.

    $ tms-recycle

stop tmsd daemon. parse flags. rebuild protobuf. rebuild daemons. clean redis. clean mongodb. setup mongodb. run tmsd with default config file.

    $ tms-recycle --rebuild

stop tmsd daemon. parse flags. clean redis. clean mongodb. setup mongodb. run tmsd with myconfig.conf file.

    $ tms-recycle --config=myconfig.conf

stop tmsd daemon. rebuild protobuf. rebuild daemons. parse flags. clean redis. clean mongodb. setup mongodb. run tmsd with myconfig.conf file.

    $ tms-recycle --rebuild --config=myconfig.conf

## EXIT STATUS

    0   Successful program execution.

## HISTORY

    14 Aug 2017 - was originally written by Mike McGann
    13 Nov 2017 - added mongodb cleaner by Mike McGann
    27 Mar 2018 - added redis flushing by Aleksandr Rassanov

## COPYRIGHT
COPYRIGHT © 2018 OROLIA GROUP AND/OR ITS AFFILIATES. ALL RIGHTS ARE STRICTLY RESERVED.
