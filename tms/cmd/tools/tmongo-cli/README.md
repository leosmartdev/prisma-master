# tmongo-cli

`tmongo-cli` is a command like helper for connection to mongo. It was created to assist with connection to mongo that is locked down with 
X509 certificate authentication, since there are a number of command line flags. Since the key files are in a well known location with 
common names, we can easily reduce the hassle of passing those flags to `mongo` and use this script like we would regular `mongo` command. 

!!! note 
    This script will not work with non authenticated mongo instances, or instances using username/password authentication. 
    
In the future, this script can be expanded as needed to provide helper functions to assist in connecting mongo instances.

## Usage 

```
tmongo-cli <mongourl> [mongo-options]
```

This script can be used exactly like `mongo` where the first argument is the hostname or `hostname:port/db-name`. Then you can pass any other mongo flags.
The output of the script will be calling `mongo` command with:

```
mongo <mongourl> --ssl --sslPEMKeyFile /etc/trident/mongo.pem  --sslCAFile /etc/trident/mongoCA.crt --authenticationDatabase '$external' --authenticationMechanism MONGODB-X509
```

Any additional flags you pass to `tmongo-cli` will
be passed directly to the `mongo` command. 

The mongourl is a normal mongo url such as `localhost/trident`, or `10.20.32.20:27017`.

### Examples

To login to mongo trident database:
```bash
tmongo-cli localhost:27017/trident
```

Get the client configuration:
```
tmongo-cli localhost/trident --eval "db.config.find().pretty()" > configuration.js
```

Run the schema updates for aaa: 
```
tmongo-cli localhost/aaa /usr/share/tms-db/aaa.js
```
