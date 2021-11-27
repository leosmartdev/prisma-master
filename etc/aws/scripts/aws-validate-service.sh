#!/bin/bash
cd /opt/prisma
export PYTHONWARNINGS="ignore:Unverified HTTPS request"
echo Integration tests started on `date` &> tests/acceptance/integration.log
./tests/acceptance/test-tmsd &>> tests/acceptance/integration.log
sleep 5
cd tests/acceptance
# tmsd -info &>> integration.log
# id -un &>> integration.log
# netstat -vatn &>> integration.log
./run-tests $(/opt/prisma/tests) &>> integration.log
python3 slack.py 'integration.log' $? $(cat /opt/prisma/etc/aws.txt)
