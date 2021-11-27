# Introduction
tselfsign generates own certificates.

## Command line usage

tselfsign is a cli to generate SSL certificates for C2 server. Furthermore, it can be used to generate X509 certificates for mongodb cluster authentification. The use of tselfsign is only recommended for development or test machines. For production or demo machines, we recommended getting a signed certificate from a Certificate Authority.

## Usage

Below is the output of -help for tselfsign

```bash
 -CACertificate string
   	write out CA file to the referenced file (default "/etc/trident/mongoCA.crt")
 -CAKey string
   	write out CA PK to referenced file (default "/etc/trident/mongoCA.key")
 -certificate string
   	write out certificate to this file (default "/etc/trident/certificate.pem")
 -dns string
   	comma delimted list of DNS names to include
 -f	force to regenerate CA key and crt even if they already exist
 -generateCA
   	flag to generate CA crt and key default is false
 -generateMongoCertificate
   	flag to generate mongo certs default is false
 -ip string
   	comma delimited list of IP addresses to include
 -key string
   	write out key to this file (default "/etc/trident/key.pem")
 -mongoCertificatePath string
   	write out mongo pem file to the referenced file (default "/etc/trident/")
```

## Generate SSL C2 server certificates

To generate use tselfsign as below:

```bash
tselfsign -ip 127.0.0.1
```

tselfsign can take dns instead of ip by using the -dns argument instead, also generated certificate and key pem files default paths can be changed by using -certificate && -key arguments.

## Generate X509 mongodb cluster certificates

- To generate certificates for your mongo cluster from scratch you can use tselfsign as below:

```bash
tselfsign  --generateCA --generateMongoCertificate -ip 127.0.0.4
```

Without --generateMongoCertificate or --generateCA flag tselfsign will generate SSL C2 certificates instead. The output of the below above command will result in 6 files under /etc/trident/ by default.

```bash
mongo.crt      mongo.key      mongo.pem
mongoCA.crt    mongoCA.key    mongoCA.srl
```

If /etc/trident is not the target needed repository, --CAKey, --CAFile, and --mongoCertificate flags should be altered.

- Also if mongoCA.crt and mongoCA.key already exist in the local machine, you can generate new cluster node cert by running:

```bash
tselfsign  --generateMongoCertificate -ip 127.0.0.5
```

This command will use the existing mongoCA.crt and mongoCA.key, and generate locally the below files:

```bash
mongo.crt      mongo.key      mongo.pem
```

If mongoCA.crt and mongoCA.key files are not in the default path then you can use -CAFile, and -CAkey argument to point to them.

Finally, if you want tselfsign to regenerate new mongoCA.crt and mongoCA.key instead of using the existing ones, you can run the below command:

```bash
tselfsign  -f --generateCA --generateMongoCertificate -ip 127.0.0.5
```

This command will force mongoCA.crt and mongoCA.key to regenerate. The command will generate the below files locally:

```bash
mongo.crt      mongo.key      mongo.pem
mongoCA.crt    mongoCA.key    mongoCA.srl
```

-f does no have to be used  --generateCA and generateMongoCertificate, for example:

```bash
tselfsign  -f --generateCA
```

will regenerate just these files:

```bash
mongoCA.crt    mongoCA.key    mongoCA.srl
```

and

```bash
tselfsign  -f --generateMongoCertificate -ip 127.0.0.5
```

will just regenerat these files:

```bash
mongo.crt      mongo.key      mongo.pem
```
