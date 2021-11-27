## Docker

```bash
docker network create my-mongo-cluster

docker run -d -p 27017:27017 --name mongo-node1 --net my-mongo-cluster mongo:3.6 --replSet "rs0"
docker run -d -p 27018:27017 --name mongo-node2 --net my-mongo-cluster mongo:3.6 --replSet "rs0"
docker run -d -p 27019:27017 --name mongo-node3 --net my-mongo-cluster mongo:3.6 --replSet "rs0"

docker exec -it mongo-node1 mongo
```

Mongo shell
```javascript
config = {
    "_id" : "rs0",
    "members" : [
        {
            "_id" : 0,
            "host" : "mongo-node1:27017"
        },
        {
            "_id" : 1,
            "host" : "mongo-node2:27017"
        },
        {
            "_id" : 2,
            "host" : "mongo-node3:27017"
        }
    ]
}

rs.initiate(config)

rs.status()
```

```bash
cd etc/mongodb/

mongo localhost/aaa schema/aaa.js 
mongo localhost/trident schema/trident.js
```