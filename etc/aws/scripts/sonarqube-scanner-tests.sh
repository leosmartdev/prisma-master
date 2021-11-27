#!/usr/bin/env bash
#set -xe
cd /opt/go/src/prisma
#sudo chown -R ec2-user tests
#sudo chmod 0777 -R tests

export PATH=$PATH:/opt/sonar-scanner/bin
export SONAR_RUNNER_HOME=/opt/sonar-scanner

login=$(aws ssm get-parameters --region us-east-1 --names SonarQubeUser --with-decryption --query Parameters[0].Value)
login=`echo $login | sed -e 's/^"//' -e 's/"$//'`
password=$(aws ssm get-parameters --region us-east-1 --names SonarQubePassword --with-decryption --query Parameters[0].Value)
password=`echo $password | sed -e 's/^"//' -e 's/"$//'`

sonar-scanner -X \
    -Dsonar.host.url=https://sonar.mcmurdo.io \
    -Dsonar.login=$login \
    -Dsonar.password=$password \
    -Dsonar.projectName=prisma-tests-1 \
    -Dsonar.projectKey=prisma-tests-1 \
    -Dsonar.projectVersion=1.0 \
    -Dsonar.sourceEncoding=UTF-8 \
    -Dsonar.sources=tests/acceptance \
    -Dsonar.exclusions=**/*.go,**/*.xml,**/*.js,vendor/**/*.go \
    -Dsonar.projectBaseDir=/opt/go/src/prisma \
    -Dsonar.scm.disabled=true \
    -Dsonar.scm.enabled=false \
    -Dsonar.language=py \
    -Dsonar.working.directory=/opt/go/src/prisma/.scannerwork > sonar-scanner-tests-stdout.txt 2> sonar-scanner-tests-stderr.txt
