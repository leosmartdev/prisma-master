#!/bin/bash
set -e -x
# create a user in EC2
adduser vagrant &&

# System dependencies
yum update
yum install -y \
    build-essential \
    python-minimal \
    python3-pip \
    mongodb-org \
    vsftpd \
    apache2-utils \
    libpam-pwdfile \
	libfontconfig1 \
	libxrender1 \
    unzip
( \
	cd /opt ; \
	curl --silent -O -L https://github.com/wkhtmltopdf/wkhtmltopdf/releases/download/0.12.4/wkhtmltox-0.12.4_linux-generic-amd64.tar.xz ; \
	tar xpvf wkhtmlto*.tar.xz ; \
	rm -rf wkhtmltox-0.12.4_linux-generic-amd64.tar.xz ; \
	rm -rf /opt/wkhtmltopdf ; \
	mv wkhtmltox /opt/wkhtmltopdf ; \
	cd /opt/wkhtmltopdf ; \
	find /opt/wkhtmltopdf -type d -exec chmod 755 {} \; ;\
    find /opt/wkhtmltopdf -type f -exec chmod 644 {} \; ;\
	chmod 755 /opt/wkhtmltopdf/bin/wkhtmltopdf ; \
)

( \
    cd /opt ; \
    curl --silent -O -L https://releases.hashicorp.com/consul/1.0.7/consul_1.0.7_linux_amd64.zip ; \
    unzip consul*.zip ; \
    rm -rf consul_1.0.7_linux_amd64.zip ; \
    rm -rf /usr/local/bin/consul ; \
    mv consul /usr/local/bin
)

# install redis
cd /opt
wget http://download.redis.io/releases/redis-4.0.9.tar.gz
tar xzf redis-4.0.9.tar.gz
rm -rf redis-*.tar.gz*
( cd redis-4.0.9 && make && make install)

# Config and runtime files
cp /opt/prisma/vagrant/tms-dev.sh /etc/profile.d
cp /opt/prisma/vagrant/go.sh /etc/profile.d
cp /opt/prisma/vagrant/protoc.sh /etc/profile.d
cp /opt/prisma/vagrant/wkhtmltopdf.sh /etc/profile.d
install -o vagrant -g vagrant -d /etc/trident
install -o vagrant -g vagrant -d /etc/redis
install -o vagrant -g vagrant -d /var/lib/redis
install -o vagrant -g vagrant -d /var/trident/db
install -o vagrant -g vagrant -d /var/trident/db-it
install -o vagrant -g vagrant /opt/prisma/vagrant/tmsd.conf /etc/trident/tmsd.conf
install -o vagrant -g vagrant /opt/prisma/vagrant/sit185-template.json /etc/trident/sit185-template.json
install -o vagrant -g vagrant /opt/prisma/vagrant/00-tmsd.conf /etc/rsyslog.d/00-tmsd.conf
install -o vagrant -g vagrant /opt/prisma/etc/tls/intermediate/certs/localhost-ca-chain.cert.pem /etc/trident/certificate.pem
install -o vagrant -g vagrant /opt/prisma/etc/tls/intermediate/private/localhost.key.pem /etc/trident/key.pem
install -o vagrant -g vagrant /opt/prisma/vagrant/services/mongodb/mongo.service /lib/systemd/system/mongo.service
install -o vagrant -g vagrant /opt/prisma/vagrant/services/redis/redis.service /lib/systemd/system/redis.service
install -o vagrant -g vagrant /opt/prisma/vagrant/services/redis/redis.conf /etc/redis/redis.conf
install -o vagrant -g vagrant /opt/prisma/etc/reports/incident-processing-form.html /etc/trident/incident-processing-form.html

# FTP configs for MCC
install /opt/prisma/vagrant/vsftpd.conf /etc/vsftpd.conf
install /opt/prisma/vagrant/vsftpd.pam /etc/pam.d/vsftpd
install -o vagrant -g vagrant /opt/prisma/vagrant/vsftpd.passwd /etc/trident/vsftpd.passwd
install -o vagrant -g vagrant -d /srv/ftp/test

# Acceptance test requirements
pip3 install --user -r /opt/prisma/tests/acceptance/requirements.txt
pip3 install --user -r /opt/prisma/tms/cmd/tools/tdemo/requirements.txt

# Link mongodb scripts and install
ln -s /opt/prisma/etc/mongodb/ /usr/share/tms-db &&

# Restart modified services
systemctl daemon-reload
systemctl restart mongo.service
systemctl enable mongo.service
systemctl enable redis
systemctl restart vsftpd.service
systemctl restart rsyslog.service
systemctl start redis &

# test integration
rm -rf /etc/profile
echo "export PATH=$PATH:/opt/prisma/bin" > /etc/profile
source /etc/profile