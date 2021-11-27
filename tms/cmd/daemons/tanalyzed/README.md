# Introduction
tanalyzed is a daemon, which allows to analyze and process any updates
as a streamer


### Sharding data
Any updates which passes a rule(An action for the rule should be forward)
can be shared between data centers.
##### How does it work?
- Each data centers has consul servers and clients - services.
- tanalyzed is ran with next params:
    - datacenter(string): a name of the datacenter where tanalyzed is ran.
    - consul-server(ip): tanalyzed should be able to connect to the server.
- So a module forwarder is able to connect to the server and discover
any data centers by using [gossip](https://www.consul.io/docs/internals/gossip.html).
- The module runs several gourutines for sync with the server
- So a client has an opportunity to create a rule with a forward action.
- Each updates are passed the rule will be shared.
##### How test it
1. The Vagrantfile has settings for launching multiple machines, please
read the readme file in the root of this project.
2. Launch all machines:
    - The first machine:
    ```bash
    vagrant up
    vagrant ssh
    tmsd -start &
    ```
    - The second machine:
    ```bash
    vagrant up cc2
    vagrant ssh cc2
    tmsd -start -config /etc/trident/tmsd_cc2.conf
    ```
3. On the host machine run the next snippet:
    - The first tab:
    ```bash
    cd $GOPATH/src/prisma/c2
    npm run -- --host=localhost
    ```
    - The second tab:
    ```bash
        cd $GOPATH/src/prisma/c2
        npm run -- --host=110.0.0.11
    ```
You have to be able to see several windows with vessels.

4. Create a vessel by using tsimulator
    ```bash
      curl -X POST \
      http://127.0.0.1:8089/v1/target \
      -H 'cache-control: no-cache' \
      -H 'content-type: application/json' \
      -H 'postman-token: af371843-c53e-26c9-3968-5d97808d5aa0' \
      -d '{
          "device": "ais",
          "mmsi": 563004370,
          "destination": "RU",
          "type": 30,
          "name": "TESTFORWARD",
          "eta": "02091504",
          "pos": [
            {
              "latitude": 1,
              "longitude": 1,
              "speed": 20
            },
            {
              "latitude": 1.162931,
              "longitude": 102.511706
            }
          ]
        }'
    ```
5. Create a rule by using the rule engine
    ```bash
    curl -X POST \
      http://110.0.0.11:8181/api/v1/auth/session \
      -H 'cache-control: no-cache' \
      -H 'content-type: application/json' \
      -d '{
      "userName":"admin",
      "token":"admin"
    }' &&
    curl -X POST \
      http://110.0.0.11:8080/api/v2/rule \
      -H 'cache-control: no-cache' \
      -H 'content-type: application/json' \
      -d '{"id":"testforward","name":"testforward","all":{"operandType":"target","operation":"EQUAL", "target":{"nmea":{"vdm": {"m1371":{"mmsi":563004370}}}}, "checkFields":["nmea.vdm.m1371.mmsi"]},  "forward": {"dc": ["dc1"]}}'
    ```
6. Open a c2 instance for localhost and take a look
If you cannot see the vessel try the follow tips:
    - Be sure you tsimulator has a station for a point [1,1] with type 'ais'
    - Try to reinstall dependencies for frontend and launch it again


