# Packer

## Packer for AWS

#### Build image
```
packer build amd64-ubuntu-build.json
see https://console.aws.amazon.com/ecs/home?region=us-east-1#/repositories
aws ecr get-login --no-include-email --region us-east-1
docker login (from above)
docker tag orolia/amd64-ubuntu-build (from above)
docker push (from above)
see https://gitlab.com/orolia/prisma/container_registry
```

#### Build prisma
```
GOPATH=/root/go
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/opt/go/bin:/root/go/bin
docker cp prisma/ 70f024b8a7a8:/root/go/src
cd /root/go/src/prisma
glide i
make protobuf
make

```

#### Build c2
```
https://www.electron.build/multi-platform-build#docker
```

## Docker for GitLab
docker tag orolia/electron-build:latest registry.gitlab.com/orolia/prisma/electron-build:latest
see https://gitlab.com/orolia/prisma/container_registry
docker login registry.gitlab.com
docker push registry.gitlab.com/orolia/prisma/electron-build

## Packer for Pico

#### PicoCluster
picocluster@pc0:~$ lsb_release -a
No LSB modules are available.
Distributor ID:	Ubuntu
Description:	Ubuntu 16.04.3 LTS
Release:	16.04
Codename:	xenial
picocluster@pc0:~$ uname -a
Linux pc0 3.10.106-148 #1 SMP PREEMPT Sat Sep 23 04:20:40 UTC 2017 armv7l armv7l armv7l GNU/Linux



mkdir go
mkdir go/src
docker cp prisma/ af7bed64ba16:/root/go/src
~/go/src/prisma# make protobuf compile

docker cp af7bed64ba16:/root/go/bin .
tar -cvzf orolia-prisma-armhf-ubuntu-bin.tar.gz bin

ex.
docker cp foo.txt mycontainer:/foo.txt
docker cp mycontainer:/foo.txt foo.txt

### install go
cd /opt
curl --silent -O -L <url-to-golang>
tar xf go*.tar.gz

### install protobuf
cd /opt
curl --silent -O -L https://github.com/google/protobuf/archive/v3.5.0.tar.gz
mkdir protobuf
tar xf v3*.tar.gz
mkdir /opt/protoc ; \
cd /opt/protoc ; \
unzip ../protoc*.zip ; \
find /opt/protoc -type d -exec chmod 755 {} \; ;\
find /opt/protoc -type f -exec chmod 644 {} \; ;\
chmod 755 /opt/protoc/bin/protoc ; \

### Config and runtime files
cp /vagrant/vagrant/tms-dev.sh /etc/profile.d
cp /vagrant/vagrant/go.sh /etc/profile.d
cp /vagrant/vagrant/protoc.sh /etc/profile.d


#### EJDB
https://github.com/Softmotions/ejdb
apt-get install software-properties-common python-software-properties

libejdb-1.a
libejdb.so -> libejdb.so.1
libejdb.so.1 -> libejdb.so.1.2.12
libejdb.so.1.2.12

/usr/lib/x86_64-linux-gnu/libejdb-1.a
-rw-r--r--  1 root root  1711718 Mar 21  2017 libejdb-1.a
lrwxrwxrwx  1 root root       12 Mar 21  2017 libejdb.so -> libejdb.so.1
lrwxrwxrwx  1 root root       17 Mar 21  2017 libejdb.so.1 -> libejdb.so.1.2.12
-rw-r--r--  1 root root  1301440 Mar 21  2017 libejdb.so.1.2.12