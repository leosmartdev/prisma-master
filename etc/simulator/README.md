# Quick Guide

This directory has a JSON file, ais_invalid_mid.json, that simulates a 
ship with an invalid MID. To simulate this track for testing, ensure that
the simulator and a tnoid is configured in tmsd.conf:

```text
{tsimulator -fstations /vagrant/vagrant/stations_data.json -fvessels /vagrant/vagrant/vessels_data.json --webaddr :8089}
{tnoid --address :9011}
```

Start the TMS system normally. The simulated track will be plotted at 
latitude 1, longitude 1. Start c2 at that location:

```
npm run local -- --center=1,1
```

Add the target with:

```
curl -X POST -H "Content-type: application/json" -d @./etc/simulator/ais_invalid_mid.json http://localhost:8089/v1/target
```


