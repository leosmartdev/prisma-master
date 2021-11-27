# PRISMA Changelog

For client changes see [https://github.com/orolia/prisma-ui/blob/develop/CHANGELOG.md](https://github.com/orolia/prisma-ui/blob/develop/CHANGELOG.md)

More release information and downloads are at [https://github.com/orolia/prisma/releases](https://github.com/orolia/prisma/releases)

## 1.8.0 release 2 (Oct 19, 2020)

This second version of 1.8.0 fixes a critical ADS-B bug.

### TMS Changes

[FIX] Create track and registry ID using callsign or Code S. [PR#110](https://github.com/mcmsar/prisma/pull/110)

### TMS Changes

* [NEW] Add support for TEST USER PROTOCOL decoding. [PR#98](https://github.com/mcmsar/prisma/pull/98)
* [NEW] Add vts support. [PR#100](https://github.com/mcmsar/prisma/pull/100)
* [NEW] Add support for beacon ID with orbitography Protocol data. [PR#103](https://github.com/mcmsar/prisma/pull/103)
* [NEW] SIT 915 support. [PR#99](https://github.com/mcmsar/prisma/pull/99)
* [NEW] Added support for the AirNav RadarBox data ingestion. [PR#97](https://github.com/mcmsar/prisma/pull/97)
* [NEW] Add RLS support. [PR#95](https://github.com/mcmsar/prisma/pull/95)

## 1.8.0 release 1 (Oct 19, 2020)

This is primarily a release for delivering RCC system for Qatar to replace legacy RCC.

### TMS Changes

* [NEW] Add support for TEST USER PROTOCOL decoding. [PR#98](https://github.com/mcmsar/prisma/pull/98)
* [NEW] Add vts support. [PR#100](https://github.com/mcmsar/prisma/pull/100)
* [NEW] Add support for beacon ID with orbitography Protocol data. [PR#103](https://github.com/mcmsar/prisma/pull/103)
* [NEW] SIT 915 support. [PR#99](https://github.com/mcmsar/prisma/pull/99)
* [NEW] Added support for the AirNav RadarBox data ingestion. [PR#97](https://github.com/mcmsar/prisma/pull/97)
* [NEW] Add RLS support. [PR#95](https://github.com/mcmsar/prisma/pull/95)

## 1.8 release candidate 1 (Oct 1, 2020)

This is primarily a release for delivering RCC system for Qatar to replace legacy RCC.

### TMS Changes

* [NEW] Add support for TEST USER PROTOCOL decoding. [PR#98](https://github.com/mcmsar/prisma/pull/98)
* [NEW] Add vts support. [PR#100](https://github.com/mcmsar/prisma/pull/100)
* [NEW] Add support for beacon ID with orbitography Protocol data. [PR#103](https://github.com/mcmsar/prisma/pull/103)
* [NEW] SIT 915 support. [PR#99](https://github.com/mcmsar/prisma/pull/99)
* [NEW] Added support for the AirNav RadarBox data ingestion. [PR#97](https://github.com/mcmsar/prisma/pull/97)
* [NEW] Add RLS support. [PR#95](https://github.com/mcmsar/prisma/pull/95)

### Developer Changes.

* [NEW] Add configuration for ubuntu bionic64 to the Vagrantfile. [PR#101](https://github.com/mcmsar/prisma/pull/101)
* [CHG] Changing proto-gen-go version from latest to v1.3.0 and protoc version to v3.11.4. [PR#96](https://github.com/mcmsar/prisma/pull/96)


## 1.7.7 (Jan 29, 2020)

This is primarily a release for delivering couple new features to Singapore RCC.

### TMS Changes.

* [NEW] Add parameter to allow deleted log entries for incidents. [PR#83](https://github.com/mcmsar/prisma/pull/83)
* [NEW] Add policy to prohibit user id as password. [PR#84](https://github.com/mcmsar/prisma/pull/84)

### Developer Changes.

* [NEW] add parameter "first" to track request to get the first target instead of the last. [PR#82](https://github.com/mcmsar/prisma/pull/82)
* [FIX] Corrected longitude validation in twebd. [PR#87](https://github.com/mcmsar/prisma/pull/87)
* [FIX] Add new endpoint, history-database, that has the behavior of the hold history endpoint used for tooltip and info panels. [PR#88](https://github.com/mcmsar/prisma/pull/88)
* [CHG]: Update TrackID hash function for SARSAT beacons to use only HexID as a seed. [PR#89](https://github.com/mcmsar/prisma/pull/89)

## 1.7.6 (May 24, 2019)

This is primarily a technical debt release which includes many dependency updates, documentation changes, and build system updates. This release is not intended to be released to customers. 

### TMS Changes

* [NEW] When a track is added to an incident, it will not longer be removed from the map until the incident is closed. [PR #48](https://github.com/orolia/prisma/pull/48)
* [NEW] Can now set time to live indexes on data stored in mongo so data can automatically be purged after an amount of time. [PR #38](https://github.com/orolia/prisma/pull/38)
* [CHG] Revert back to `token` property from `password` for tauthd session login. [PR #36](https://github.com/orolia/prisma/pull/36)
* [CHG] Sites can now be retrieved by mongo `_id` and site id. [PR #29](https://github.com/orolia/prisma/pull/29), [PR #41](https://github.com/orolia/prisma/pull/41)
* [CHG] Disabled insecure ciphers for TLS configurations [PR #42](https://github.com/orolia/prisma/pull/42)
* [FIX] When certificates are rolled over, daemons no longer need to be restarted to pick up the new certificates. [PR #40](https://github.com/orolia/prisma/pull/40)
* [FIX] Fixed critical issue where systems not using x509 certificates to authenticate mongo wouldn't start. [PR #35](https://github.com/orolia/prisma/pull/35)
* [FIX] Fixed issue with NAF parsing failing due to empty voltage on systems deployed for the Ministry of Fisheries. [PR #10](https://github.com/orolia/prisma/pull/10)
* [FIX] Fixed some links in the documentation site and added new documentation for docker, replication, and other updates. [PR #45](https://github.com/orolia/prisma/pull/45), [PR #46](https://github.com/orolia/prisma/pull/46), [PR #47](https://github.com/orolia/prisma/pull/47)
* [FIX] Fixed an issue where trackex wasn't loading correctly and causing tracks to disappear before they should. [PR #43](https://github.com/orolia/prisma/pull/43)

### Developer Changes

* [NEW] Finished moving the repo to github. Documentation now links to github repos. [PR #2](https://github.com/orolia/prisma/pull/2), [PR #17](https://github.com/orolia/prisma/pull/17) 
* [NEW] Added golinter [PR #31](https://github.com/orolia/prisma/pull/31)
* [NEW] Code coverage is now reported during builds [PR #30](https://github.com/orolia/prisma/pull/30)
* [NEW] Added link to upcoming prisma-ui documentation site [PR #28](https://github.com/orolia/prisma/pull/28)
* [NEW] Added codecov to the builds for code coverage reports. [PR #54](https://github.com/orolia/prisma/pull/54)
* [CHG] Added new docker images for builds and updated builds in AWS. [PR #44](https://github.com/orolia/prisma/pull/44)
  - Documentation site is now automatically deployed on builds of master, develop, and release builds. 
  - Docker updated to include more dependencies and shorten build times. 
  - Godoc documentation is now automatically deployed to `documentation.mcmurdo.io/godoc/nightly/pkg/`
* [FIX] Glide now references correct proto and assert library versions. [PR #23](https://github.com/orolia/prisma/pull/23)
* [FIX] tms-recycle is now added to tms-dev properly, added README for the tool and fixed some other issues with misconfigurations. [PR #8](https://github.com/orolia/prisma/pull/8)
* [FIX] Tests now run correctly. [PR #37](https://github.com/orolia/prisma/pull/37)
* [FIX] ejdb is now pulled from AWS since ejdb2 was released and old debians were pulled. [PR #39](https://github.com/orolia/prisma/pull/39)
* [FIX] Update Version documentation page updated [PR #52](https://github.com/orolia/prisma/pull/52)
* [RM] Client is now removed from the `orolia/prisma` respository. [PR #27](https://github.com/orolia/prisma/pull/27)

## 1.7.6 Release Candidate 2 (May 22, 2019)

* [NEW] Added codecov to the builds for code coverage reports. [PR #54](https://github.com/orolia/prisma/pull/54)
* [CHG] Updated links to PRISMA UI documentation site and moved some content around. [PR #53](https://github.com/orolia/prisma/pull/53)
* [FIX] Fixed issue with some track queries taking a long time. [PR #56](https://github.com/orolia/prisma/pull/56)
* [FIX] Godoc is now publishing to documentation site properly [PR #51](https://github.com/orolia/prisma/pull/51)
* [FIX] Update Version documentation page updated [PR #52](https://github.com/orolia/prisma/pull/52)
* [FIX] Fixed broken documentation links. [PR #50](https://github.com/orolia/prisma/pull/50)

## 1.7.6 Release Candidate 1 (May 13, 2019)

### TMS Changes

* [NEW] When a track is added to an incident, it will not longer be removed from the map until the incident is closed. [PR #48](https://github.com/orolia/prisma/pull/48)
* [NEW] Can now set time to live indexes on data stored in mongo so data can automatically be purged after an amount of time. [PR #38](https://github.com/orolia/prisma/pull/38)
* [CHG] Revert back to `token` property from `password` for tauthd session login. [PR #36](https://github.com/orolia/prisma/pull/36)
* [CHG] Sites can now be retrieved by mongo `_id` and site id. [PR #29](https://github.com/orolia/prisma/pull/29), [PR #41](https://github.com/orolia/prisma/pull/41)
* [CHG] Disabled insecure ciphers for TLS configurations [PR #42](https://github.com/orolia/prisma/pull/42)
* [FIX] When certificates are rolled over, daemons no longer need to be restarted to pick up the new certificates. [PR #40](https://github.com/orolia/prisma/pull/40)
* [FIX] Fixed critical issue where systems not using x509 certificates to authenticate mongo wouldn't start. [PR #35](https://github.com/orolia/prisma/pull/35)
* [FIX] Fixed issue with NAF parsing failing due to empty voltage on systems deployed for the Ministry of Fisheries. [PR #10](https://github.com/orolia/prisma/pull/10)
* [FIX] Fixed some links in the documentation site and added new documentation for docker, replication, and other updates. [PR #45](https://github.com/orolia/prisma/pull/45), [PR #46](https://github.com/orolia/prisma/pull/46), [PR #47](https://github.com/orolia/prisma/pull/47)
* [FIX] Fixed an issue where trackex wasn't loading correctly and causing tracks to disappear before they should. [PR #43](https://github.com/orolia/prisma/pull/43)

### Developer Changes

* [NEW] Finished moving the repo to github. Documentation now links to github repos. [PR #2](https://github.com/orolia/prisma/pull/2), [PR #17](https://github.com/orolia/prisma/pull/17) 
* [NEW] Added golinter [PR #31](https://github.com/orolia/prisma/pull/31)
* [NEW] Code coverage is now reported during builds [PR #30](https://github.com/orolia/prisma/pull/30)
* [NEW] Added link to upcoming prisma-ui documentation site [PR #28](https://github.com/orolia/prisma/pull/28)
* [CHG] Added new docker images for builds and updated builds in AWS. [PR #44](https://github.com/orolia/prisma/pull/44)
  - Documentation site is now automatically deployed on builds of master, develop, and release builds. 
  - Docker updated to include more dependencies and shorten build times. 
  - Godoc documentation is now automatically deployed to `documentation.mcmurdo.io/godoc/nightly/pkg/`
* [FIX] Glide now references correct proto and assert library versions. [PR #23](https://github.com/orolia/prisma/pull/23)
* [FIX] tms-recycle is now added to tms-dev properly, added README for the tool and fixed some other issues with misconfigurations. [PR #8](https://github.com/orolia/prisma/pull/8)
* [FIX] Tests now run correctly. [PR #37](https://github.com/orolia/prisma/pull/37)
* [FIX] ejdb is now pulled from AWS since ejdb2 was released and old debians were pulled. [PR #39](https://github.com/orolia/prisma/pull/39)
* [RM] Client is now removed from the `orolia/prisma` respository. [PR #27](https://github.com/orolia/prisma/pull/27)

## 1.7.5

* [NEW] Added documentation on Incident Transfer. [PR #19](https://github.com/orolia/prisma/pull/19)
* [FIX] Fixed `/auth/policy` endpoint not using the mongo connection with certificates. Now when authentication is turned on in mongo the policy endpoint will function. [PR #1](https://github.com/orolia/prisma/pull/1)
* [FIX] Dates with two digits are now properly parsed in SIT185 messages. [PR #7](https://github.com/orolia/prisma/pull/7)

## 1.7.4

This release focuses on adding authentication and ecryption to mongo connections.

* [NEW] `tmongo-cli`: Command line tool for more easily connecting to mongo command line when authentication is enabled. It is just a light wrapper around `mongo` client to pass certificates and authentication flags. [PR-588](https://gitlab.com/orolia/prisma/merge_requests/588)
* [NEW] `tselfsign` added flags for creating mongo CA files and instance ssl certificates. [PR-591](https://gitlab.com/orolia/prisma/merge_requests/591)
* [NEW] Mongo can now be authenticated using X509 certificates. [PR-581](https://gitlab.com/orolia/prisma/merge_requests/581)
* [CHG] Default paths for mongo authentication and certificates were changed to well know filenames in `/etc/trident` [PR-589](https://gitlab.com/orolia/prisma/merge_requests/589)
* [CHG] Updated sit185 template parsing to handle changes made to the format as of February 2018. [PR-580](https://gitlab.com/orolia/prisma/merge_requests/580)
* [CHG] Mongo schema updates are now run on daemon start. [PR-578](https://gitlab.com/orolia/prisma/merge_requests/578)
* [RM] mongo.service file has been removed. Installations should now use normal mongod.service installed by mongo and the mongo.conf file in etc for daemon lifecycle and configuration. [PR-579](https://gitlab.com/orolia/prisma/merge_requests/579)

## 1.7.4-rc2

* [NEW] `tmongo-cli`: Command line tool for more easily connecting to mongo command line when authentication is enabled. It is just a light wrapper around `mongo` client to pass certificates and authentication flags. [PR-588](https://gitlab.com/orolia/prisma/merge_requests/588)
* [NEW] `tselfsign` added flags for creating mongo CA files and instance ssl certificates. [PR-591](https://gitlab.com/orolia/prisma/merge_requests/591)
* [CHG] Default paths for mongo authentication and certificates were changed to well know filenames in `/etc/trident` [PR-589](https://gitlab.com/orolia/prisma/merge_requests/589)

## 1.7.4-rc1

* [NEW] Mongo can now be authenticated using X509 certificates. [PR-581](https://gitlab.com/orolia/prisma/merge_requests/581)
* [CHG] Updated sit185 template parsing to handle changes made to the format as of February 2018. [PR-580](https://gitlab.com/orolia/prisma/merge_requests/580)
* [CHG] Mongo schema updates are now run on daemon start. [PR-578](https://gitlab.com/orolia/prisma/merge_requests/578)
* [RM] mongo.service file has been removed. Installations should now use normal mongod.service installed by mongo and the mongo.conf file in etc for daemon lifecycle and configuration. [PR-579](https://gitlab.com/orolia/prisma/merge_requests/579)

## 1.7.3

There are no backend changes in this release.

## 1.7.3-rc2

There are no backend changes in this release.

## 1.7.3-rc1

There are no backend changes in this release.

## 1.7.2

This update provides bug fixes for Incident PDF generation, Spidertracks ingestion, SARMAP support, and a few other critical crashing bugs.

* [NEW] Installer now removes prisma repository after installation.
* [FIX] SARMAP endpoint now exports the proper GeoJSON Feature cooridinate format. [PR 557](https://gitlab.com/orolia/prisma/merge_requests/557)
* [FIX] Fixed issue that could cause the tanalyzed to crash on startup [PR 566](https://gitlab.com/orolia/prisma/merge_requests/566)
* [FIX] Added missing incident-processing-form.html file to the tms install script to be installed into /etc/trident where the incident rest calls were looking for it. [PR-562](https://gitlab.com/orolia/prisma/merge_requests/562)
* [FIX] tspiderd now properly initiates the `<body>` time so the request gets all new data for the last 15 minutes then time since last request after that. [PR-560](https://gitlab.com/orolia/prisma/merge_requests/560)

## 1.7.2-rc2

* [FIX] SARMAP endpoint now exports the proper GeoJSON Feature cooridinate format. [PR 557](https://gitlab.com/orolia/prisma/merge_requests/557)
* [FIX] Fixed issue that could cause the tanalyzed to crash on startup [PR 566](https://gitlab.com/orolia/prisma/merge_requests/566)

## 1.7.2-rc1

* [NEW] Installer now removes prisma repository after installation.
* [FIX] Added missing incident-processing-form.html file to the tms install script to be installed into /etc/trident where the incident rest calls were looking for it. [PR-562](https://gitlab.com/orolia/prisma/merge_requests/562)
* [FIX] tspiderd now properly initiates the `<body>` time so the request gets all new data for the last 15 minutes then time since last request after that. [PR-560](https://gitlab.com/orolia/prisma/merge_requests/560)

## 1.7.1

This release adds a few new features and bug fixes, specifically for the releases to Chile and Singapore. These include support for running mongo on a completely separate server on a different subnet, fixes to Incident Transfer, and a brand new documentation site and release process.

Also in this release is a single command installer for installing the system on a clean server. The Installer runs both online and completely offline and includes all dependencies needed to install the system.

* [NEW] Added new documentation site and markdown documentation to the doc directory. [PR 520](https://gitlab.com/orolia/prisma/merge_requests/520)
* [NEW] Installer Tarball is now created with dependencies and debian packages for online and offline installations in a single command. [PR 520](https://gitlab.com/orolia/prisma/merge_requests/520)
* [CHG] Spaces are no longer allowed in usernames. [PR 488](https://gitlab.com/orolia/prisma/merge_requests/488)
* [CHG] When configuration is created on install twebd will try to resolve the IP of the machine instead of just localhost for all service entries. [PR 479](https://gitlab.com/orolia/prisma/merge_requests/479)
* [FIX] Fixed properties missing on the config.json endpoint. Lat,lng, zoom, service.map, showSpidertracksLayerHide, version, git, and policy description are now correctly shown in the configuration file. [PR 525](https://gitlab.com/orolia/prisma/merge_requests/525), [PR 484](https://gitlab.com/orolia/prisma/merge_requests/484), [PR 485](https://gitlab.com/orolia/prisma/merge_requests/485)
* [FIX] When incident transfer fails, incident is now re-opened [PR 519](https://gitlab.com/orolia/prisma/merge_requests/519)
* [FIX] Fixed a bug where excluding vessels from a zone would cause a crash [PR 504](https://gitlab.com/orolia/prisma/merge_requests/504)
* [FIX] Fixed infinite loop that could occur decoding values from mongo [PR 522](https://gitlab.com/orolia/prisma/merge_requests/522)
* [FIX] Incident transfer expiration no longer causing a failure. [PR 531](https://gitlab.com/orolia/prisma/merge_requests/531)
* [FIX] tspiderd now included in tms debian package [PR 506](https://gitlab.com/orolia/prisma/merge_requests/506)
* [FIX] Fixed an issue where system could not be configured with the database running on a separate system on another subnet. [PR 500](https://gitlab.com/orolia/prisma/merge_requests/500)
* [FIX] Fixed an issue where user could not change their password. [PR 487](https://gitlab.com/orolia/prisma/merge_requests/487)
* [FIX] Fixed an instance where `twebd` would crash. [PR 490](https://gitlab.com/orolia/prisma/merge_requests/490)
* [FIX] Fixed issue where tdatabased would slow and timeout when track collection is large. [PR 454](https://gitlab.com/orolia/prisma/merge_requests/454)
* [RM] Removed misleading notification for failed Incident Transfers. [PR 526](https://gitlab.com/orolia/prisma/merge_requests/526)


## 1.7.1-rc7

* [FIX] Incident transfer expiration no longer causing a failure. [PR 531](https://gitlab.com/orolia/prisma/merge_requests/531)

## 1.7.1-rc6

This release is mostly bug fixes to hopefully complete the final release candidate for `1.7.1`. There is one new feature that is new documentation static site that reads the markdown files in `prisma/doc` and can be hosted directly in S3.

To view the site locally, run `make serve-docs` in the top level of the git repo (see `doc/index.md` for more information and how to get mkdocs installed) or run `make docs` to build the new static site that can be hosted in S3 bucket `prisma-documentation`.

* [NEW] Added new documentation site and markdown documentation to the doc directory. [PR 520](https://gitlab.com/orolia/prisma/merge_requests/520)
* [NEW] Added code for the new PRISMA installer tarball. [PR 520](https://gitlab.com/orolia/prisma/merge_requests/520)
* [FIX] Fixed properties missing on the config.json endpoint. Lat,lng, zoom, version, git, and policy description are now correctly shown in the configuration file. [PR 525](https://gitlab.com/orolia/prisma/merge_requests/525)
* [FIX] When incident transfer fails, incident is now correctly re-opened [PR 519](https://gitlab.com/orolia/prisma/merge_requests/519)
* [FIX] Fixed a bug where excluding vessels from a zone would cause a crash [PR 504](https://gitlab.com/orolia/prisma/merge_requests/504)
* [FIX] Fixed infinite loop that could occur decoding values from mongo [PR 522](https://gitlab.com/orolia/prisma/merge_requests/522)
* [RM] Removed misleading notification for failed Incident Transfers. [PR 526](https://gitlab.com/orolia/prisma/merge_requests/526)

## 1.7.1-rc5

* [FIX] tms and tms-db debian packages now have their dependencies with versions listed. This includes, redis, mongo, consul, ejdb, etc... [PR 506](https://gitlab.com/orolia/prisma/merge_requests/506)
* [FIX] tspiderd was missing from tms debian package [PR 506](https://gitlab.com/orolia/prisma/merge_requests/506)

## 1.7.1-rc4

* [FIX] Reversed the twebd crash websocket bugfix as it prevented all streaming data. tms-recycle mongo port changed. [PR 502](https://gitlab.com/orolia/prisma/merge_requests/502)

## 1.7.1-rc3

* [FIX] Fixed an issue where system could not be configured with the database running on a separate system on another subnet. [PR 500](https://gitlab.com/orolia/prisma/merge_requests/500)
* [FIX] Fixed an issue where user could not change their password. [PR 487](https://gitlab.com/orolia/prisma/merge_requests/487)
* [FIX] Fixed an instance where `twebd` would crash. [PR 490](https://gitlab.com/orolia/prisma/merge_requests/490)
* [CHG] Spaces are no longer allowed in usernames. [PR 488](https://gitlab.com/orolia/prisma/merge_requests/488)

## 1.7.1-rc2

* [NEW] Added service.map and showSpidertracksLayerHide to the configuration protos to match the client configuration. [PR 484](https://gitlab.com/orolia/prisma/merge_requests/484), [PR 485](https://gitlab.com/orolia/prisma/merge_requests/485)
* [CHG] When configuration is created on install twebd will try to resolve the IP of the machine instead of just localhost for all service entries. [PR 479](https://gitlab.com/orolia/prisma/merge_requests/479)

## 1.7.1-rc1

* [FIX] Fixed issue where tdatabased would slow and timeout when track collection is large. [PR 454](https://gitlab.com/orolia/prisma/merge_requests/454)


## 1.7.0

* [NEW] tmongo tool for interacting and testing mongo. [PR 380](https://gitlab.com/orolia/prisma/merge_requests/380)
* [NEW] Added capabilities to use Amazon CodeBuild for building in the cloud as well as integration with build reports to Slack. [PR 387](https://gitlab.com/orolia/prisma/merge_requests/387)
* [NEW] Tracks now read speed and course values from OmniCom position reports [PR 399](https://gitlab.com/orolia/prisma/merge_requests/399)
* [NEW] Added GET `/zone` that returns GeoJSON collection of zones. [PR 406](https://gitlab.com/orolia/prisma/merge_requests/406)
* [NEW] Added support for Spidertracks [PR 404](https://gitlab.com/orolia/prisma/merge_requests/404), [PR 408](https://gitlab.com/orolia/prisma/merge_requests/408), [PR 415](https://gitlab.com/orolia/prisma/merge_requests/415)
* [NEW] Added support for uploading geofences to an OmniCom beacon. [PR 414](https://gitlab.com/orolia/prisma/merge_requests/418), [PR 420](https://gitlab.com/orolia/prisma/merge_requests/420), [PR 421](https://gitlab.com/orolia/prisma/merge_requests/421)

* [CHG] Upgraded to golang protobuf v1.10
* [CHG] Upgraded to protobuf c65a0412e71e8b9b3bfd22925720d23c0f054237 [PR 399](https://gitlab.com/orolia/prisma/merge_requests/399)
* [CHG] Upgraded to Consul 1.0.7 [PR 380](https://gitlab.com/orolia/prisma/merge_requests/380)
* [CHG] Upgraded to support final OmniCom Beacon firmware and API [PR 399](https://gitlab.com/orolia/prisma/merge_requests/399)
* [CHG] Switched Websocket stream format from protobuf to json. Now all websockets are in json format. [PR 407](https://gitlab.com/orolia/prisma/merge_requests/407)
* [CHG] Moved Envelope proto from websocket proto files into it's own location. [PR 393](https://gitlab.com/orolia/prisma/merge_requests/393)

* [FIX] Issues with vagrant make protobuf failing [PR 380](https://gitlab.com/orolia/prisma/merge_requests/380)
* [FIX] Updates to Multicast API to match the documented API in Postman. [PR 381](https://gitlab.com/orolia/prisma/merge_requests/381)
* [FIX] Websockets will now close properly. [PR 392](https://gitlab.com/orolia/prisma/merge_requests/392)
* [FIX] twebd leaks [PR 391](https://gitlab.com/orolia/prisma/merge_requests/391)
* [FIX] [PR 399](https://gitlab.com/orolia/prisma/merge_requests/399)
* [FIX] Locked accounts can no longer login. [PR 440](https://gitlab.com/orolia/prisma/merge_requests/440)
