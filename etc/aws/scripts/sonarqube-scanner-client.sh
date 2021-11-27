#!/usr/bin/env bash
#set -xe
cd /opt/go/src/prisma
sudo chmod 0777 -R client
#ls -la client/yarn.lock
( cd client ; yarn install ) > yarn_install_stdout.txt 2> yarn_install_stderr.txt
( yarn global add jest ) > yarn_jest_stdout.txt 2> yarn_jest_stderr.txt
( cd client ; yarn build ) > yarn_build_stdout.txt 2> yarn_build_stderr.txt
( cd client/packages/prisma-map ; yarn test --coverage ) > yarn_coverage_map.txt 2> yarn_coverage_map_stderr.txt
( cd client/packages/prisma-ui ; yarn test --coverage ) > yarn_coverage_ui.txt 2> yarn_coverage_ui_stderr.txt
( cd client/packages/prisma-electron ; rm -f src/bundle*.js ) > electron_bundle_stdout.txt 2> electron_bundle_stderr.txt
( cd client/packages/prisma-electron ; yarn test:coverage --ci --no-color ) > yarn_coverage_test.txt 2> yarn_coverage_stderr.txt

export PATH=$PATH:/opt/sonar-scanner/bin
export SONAR_RUNNER_HOME=/opt/sonar-scanner

login=$(aws ssm get-parameters --region us-east-1 --names SonarQubeUser --with-decryption --query Parameters[0].Value)
login=`echo $login | sed -e 's/^"//' -e 's/"$//'`
password=$(aws ssm get-parameters --region us-east-1 --names SonarQubePassword --with-decryption --query Parameters[0].Value)
password=`echo $password | sed -e 's/^"//' -e 's/"$//'`

sonar-scanner \
    -Dsonar.host.url=https://sonar.mcmurdo.io \
    -Dsonar.exclusions=**/bundle.js,**/*_pb.js \
    -Dsonar.javascript.lcov.reportPaths=client/packages/prisma-map/coverage/lcov.info,client/packages/prisma-ui/coverage/lcov.info,client/packages/prisma-electron/coverage/lcov.info \
    -Dsonar.login=$login \
    -Dsonar.password=$password \
    -Dsonar.projectBaseDir=/opt/go/src/prisma \
    -Dsonar.projectName=prisma-client-1 \
    -Dsonar.projectKey=prisma-client-1 \
    -Dsonar.projectVersion=1.6 \
    -Dsonar.sourceEncoding=UTF-8 \
    -Dsonar.sources=client \
    -Dsonar.working.directory=/opt/go/src/prisma/.scannerwork > sonar-scanner-client-stdout.txt 2> sonar-scanner-client-stderr.txt
