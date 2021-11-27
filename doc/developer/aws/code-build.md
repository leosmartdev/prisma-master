# AWS CodeBuild

We are using code build and docker to build and release our application in AWS. We have a number of
builds that run depending on where in the repo the code is pushed, a pull request is opened, etc...

This document will go over the builds, how they are setup, and how the environments are created.
This will apply to prisma and prisma-ui repositories.

## Build Projects

For our main projects we build on the following conditions:

* Pull Request Opened/Updated
* Push to develop
* Push to master
* (TODO) Push to release/*
* Tag created

For each of these conditions, we have a separate project in code build that runs based on a buildspec file with the same name as the project in `.aws` directory at the top level of the repo. The
buildspec files all follow the same pattern of `<projectname>.buildspec.yml`.

Project name follows a convention as well. For the develop and master branch pushes the project
names are just the names of the repo `prisma`/`prisma-ui` as these are the basic builds on every push to those branches.
Release and tag created builds are called `release-prisma-tms`/`release-prisma-ui`.

The builds per pull request are called `<projectname>-pull-request`, so `prisma-pull-request`/`prisma-ui-pull-request`.

### Pull Request Builds

On every pull request opened, reopened, or updated, the pull request build is run. This will compile all code, run unittests, linters and if available, push code coverage to codecov. The pull requests in github are setup to only allow merging when the builds succeed.

!!! note
    If a build fails and you re-run the build in CodeBuild, it will not report back to github. So your options are to push again to the branch to trigger a rebuild, or to turn off the branch protection rules, merge, then turn them back on.

### Develop and Master builds

For every push to develop or master, a build is run that compiles the code, runs unittests, linters, and deploys the documentation sites. For frontend repositories this will also deploy the storybook sites. For backend we will deploy the godoc pkg site. See [Static Sites](./static-sites.md) for more information about the sites and their deployed layout.

For master builds, all docs and storybook sites are deployed to the `latest` folders. For develop builds they are sent to `nightly`. So for

### Release Builds

On every tag push to the repo, a release version will be build. This will compile the code, run the tests, linters, then build production packages.

For `orolia/prisma` this means the debians will be built as well as the PRISMA_Server_Install installer. For the `orolia/prisma-ui` repo the electron Windows installer and the mac package will also be built. The version for the packages is derived directly from the tag name, so be sure the tag always follows `vX.X.X` or `v.X.X.X-rcX`. See [Update Version Numbers](../../developer/update-version-numbers.md) and [Release Process](../../developer/release-process.md) for more information.

## Build Images (Docker)

Our builds in CodeBuild are mostly done inside a docker image provided by Amazon ECR. We create these images and push to ECR locally in our development environments. See [ECR](#ECR) below for more information about pushing.

### Our Images

We currently have 3 docker images used for building.

 * [prisma](https://github.com/orolia/prisma/tree/develop/Dockerfile)
 * [prisma-ui](https://github.com/orolia/prisma-ui/tree/develop/Dockerfile)
 * [prisma-build-electron](https://github.com/orolia/prisma-ui/tree/develop/packages/prisma-electron/Dockerfile)

To build an image go into the folder containing the Docker file and run:

```
docker image build -t <image-name>:<tag> .
```

Image name is the image name above, and tag is a version (1.7.6-rc1) or `latest` to build as the `latest`. In general its usually best practice to give a specific tag to every build, then additionally tag the latest build as well, so both the version and the `latest` tag will point to a single image. You can make as many tags of an image as needed.

Tagging:
```
docker tag <existing-tag> <new-tag>
```

Eg:
```
docker tag prisma:1.7.6 prisma:latest
```

If you want to spin up an instance of an image, you can use the following command to start up the image and get a shell:

```
docker container run -ti <tag> /bin/bash
```

Example:
```
docker container run -ti prisma-ui:latest /bin/bash
```

#### prisma

This is the docker image to build all of the `orolia/prisma` repo. This is used for all pull requests, develop/master, and release builds. This is based on a go image, so if we upgrade the go version we just need to change the `FROM` line to upgrade to a new go.

#### prisma-ui

This image is used for building the develop/master and pull request builds for `orolia/prisma-ui`. This is based off a node image so if we upgrade node we just need to change the `FROM` line.

#### prisma-electron-build

This image only builds the release installers for `orolia/prisma-ui`. This is based off `electron-userland/wine-mono` which sets up a lot of the build system for building the windows packages in an debian based linux container.

### ECR

Authenticate docker with AWS. (must install awscli first using `pip3 install --upgrade awscli` on
Linux or `brew install awscli` on macOS). This authentication is good for 12 hours.
([AWS ECR Docs](https://docs.aws.amazon.com/AmazonECR/latest/userguide/Registries.html)). If you are
installing for the first time the aws cli, you can authenticate using your access key id and secret
`aws configure`. Default region is `us-east-1` and other than access key id, secret, and region,
everything else can be left blank.

```
$ `aws ecr get-login --region us-east-1 --no-include-email`
```

To push to ECR, you just need to tag a docker image with the right tag then you can docker push
that tag which will push it into the ECR right away.

General, the tags should be the following:

```
<ACCOUNTID>.dkr.ecr.<REGION>.amazonaws.com/<ERC REPOSITORY NAME>:<TAG>
```

For example, for the prisma build images, we have the following tags:

```
370461448094.dkr.ecr.us-east-1.amazonaws.com/prisma:latest
370461448094.dkr.ecr.us-east-1.amazonaws.com/prisma:1.7.6
370461448094.dkr.ecr.us-east-1.amazonaws.com/prisma:1.8.0-alpha1
```

The builds generally always use the `latest` tag, so if you push a new image to latest then the next
build or build retry will grab the new docker image. You can specify these tags to use during build
in the Environment section of Code Build.

These commands can also be found for each image at [AWS ECR Console](https://console.aws.amazon.com/ecr/repositories?region=us-east-1) and after selecting an image hitting the `View Push Commands` button on the top right.
