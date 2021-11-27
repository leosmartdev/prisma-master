#!/usr/bin/env bash

cd /opt/go/src/prisma

export GOPATH=/opt/go/
export PATH=$PATH:/opt/go/bin:/opt/sonar-scanner/bin
export WORKDIR=/opt/go/src
export SONAR_RUNNER_HOME=/opt/sonar-scanner

glide install > glide_stdout.txt 2> glide_stderr.txt
( cd vendor/github.com/golang/protobuf/protoc-gen-go ; go install )
make protobuf > protobuf_stdout.txt 2> protobuf_stderr.txt

gocov test ./... | gocov-xml > coverage.xml 2> gocov_test_stderr.txt
gometalinter --install > gometalinter_install_stdout.txt 2> gometalinter_install_stderr.txt
gometalinter ./... --deadline 10m --checkstyle > report.xml 2> gometalinter_stderr.txt

login=$(aws ssm get-parameters --region us-east-1 --names SonarQubeUser --with-decryption --query Parameters[0].Value)
login=`echo $login | sed -e 's/^"//' -e 's/"$//'`
password=$(aws ssm get-parameters --region us-east-1 --names SonarQubePassword --with-decryption --query Parameters[0].Value)
password=`echo $password | sed -e 's/^"//' -e 's/"$//'`

sonar-scanner \
    -Dsonar.host.url=https://sonar.mcmurdo.io \
    -Dsonar.login=$login \
    -Dsonar.password=$password \
    -Dsonar.projectName=prisma-tms-1 \
    -Dsonar.projectKey=prisma-tms-1 \
    -Dsonar.projectVersion=1.0 \
    -Dsonar.sourceEncoding=UTF-8 \
    -Dsonar.sources=tms \
    -Dsonar.exclusions=**/*.py,**/*.pb.go,**/*.xml,**/*.js,vendor/**/*.go \
    -Dsonar.projectBaseDir=/opt/go/src/prisma \
    -Dsonar.working.directory=/opt/go/src/prisma/.scannerwork > sonar-scanner-tms-stdout.txt 2> sonar-scanner-tms-stderr.txt
