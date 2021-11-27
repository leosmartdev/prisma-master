# AWS CodePipeline (Deprecated)

We are not currently using code pipeline, but looking at ways to use it in the
future since we shouldn't be cramming everything we are in CodeBuild. But, because code pipeline
doesnt let us start a pipeline from a merge request or tag without a lambda function easily,
code build is the way to go. - John April 2019

## Previous document

We run a number of code pipelines currently for building the application.

[Our AWS CodePipeline Management Console](https://console.aws.amazon.com/codesuite/codepipeline/pipelines?region=us-east-1)

## release-prisma-tms

This pipeline builds tms and schema, run sonar scanner and integration tests.

[See in AWS](https://console.aws.amazon.com/codesuite/codepipeline/pipelines/release-prisma-tms/view?region=us-east-1)

* Source action that get a code from repository and places it in MyApp as an artifact.
* Build action named CodeBuild that build TMS binary files from Golang source and places it in MyAppBuild as an artifact.
* SonarScanner-PrismaClient action that places a appspec-sonar-scanner-client.yml file as an appspec.yml in a root folder and places it in SonarScanner-PrismaClient as an artifact.
* SonarScanner-PrismaTest action that places appspec-sonar-scanner-tests.yml as an appspec.yml in a root folder and places it in SonarScanner-PrismaTests as an artifact.
* SonarScanner-PrismaTMS action that places appspec-sonar-scanner-tms.yml as an appspec.yml in a root folder and places it in SonarScanner-PrismaTMS as an artifact.
* Client-Tests action run a yarn test:coverage use a MyApp artifact and places a stdout to PrismaReleaseClient-Tests as an artifact.
* TMS-Test action run a golang toot:cover use a MyApp artifact and place a stdout to PrismaReleaseTMS-UnitTests as an artifact.
* SonarScanner-PrismaTests action run a sonar scanner use a SonarScanner-PrismaTests artifact on a EC2 inctance tagged prisma-test-release.
* SonarScanner-Client action run a sonar scanner use a SonarScanner-PrismaClient artifact on a EC2 inctance tagged prisma-test-release.
* SonarScanner-PrismaTMS action run a sonar scanner use a SonarScanner-PrismaTMS artifact on EC2 inctance tagged prisma-test-release.
* integration-test action run a integration tests use a MyAppBuild artifact on EC2 inctance tagged prisma-test-release.

## develop-prisma-tms

Run integration tests.

[See in AWS](https://console.aws.amazon.com/codesuite/codepipeline/pipelines/develop-prisma-tms/view?region=us-east-1)

* Source action that get a code from repository and places it in MyApp as an artifact.
* Build action named CodeBuild that build TMS binary files from Golang source and places it in MyAppBuild as an artifact.
* Integration_Test action run a integration tests use a MyAppBuild artifact on EC2 inctance tagged prisma-test.
* prisma-dg action deploy prisma on EC2 inctances tagged version:1.5, environment:production, stack-level:application.

## prisma-pc2-294

it's used for debugging.

[See in AWS](https://console.aws.amazon.com/codesuite/codepipeline/pipelines/prisma-pc2-294/view?region=us-east-1)

* Source action that get a code from repository and places it in MyApp as an artifact.
* Build action named CodeBuild that build TMS binary files from Golang source and places it in MyAppBuild as an artifact.
* test action run a integration tests use a MyAppBuild artifact on EC2 inctance tagged prisma-test.
* prisma-dg action deploy prisma on EC2 inctances tagged version:1.5, environment:production, stack-level:application.
