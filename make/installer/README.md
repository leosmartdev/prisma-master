# PRISMA Installer

The PRISMA Installer is a self extracting tarball that allows for completely offline installation.
This installer will extract a debian repository locally on the system then configure apt to read
that repository for our debian packages as well as required dependencies. Then the installer will
run `apt install tms tms-mcc tms-db` to install the PRISMA system onto the server.

## Using the Installer

## Building the Installer Package

The installer can be built using the [AWS CodeBuild project `package-prisma-installer`](https://console.aws.amazon.com/codesuite/codebuild/projects/package-prisma-installer/history?region=us-east-1).

This will pull the debian packages from the S3 bucket folder set in the Source settings (Source 2 `built_debians`)
and all dependency debian packages from `s3://prisma-c2/dependencies/all-dependencies`.

For more information and detailed instructions on running the CodeBuild installer project, see the section on the installer packaging in [release PRISMA C2 tutorials](http://documentation.mcmurdo.io/developer/release-process#run-the-package-installer-build).

To run manually, the installer packager (`package-prisma-repository`) must be run on a system that
already has a prisma repo created in `/usr/local/prisma-repository`.

### Creating the local repository

If you need to build manually, you will need to create the local repository first. Use the following commands to create the folder and copy the dependency debians and built TMS debians into the repository and run the `dpkg` command to initialize the repo.

```
mkdir -p /usr/local/prisma-repository
cp path/to/debian/dependencies/*.deb /usr/local/prisma-repository
cp path/to/built/tms/debians/*.deb /usr/local/prisma-repository
./update-prisma-repository
```


### Installing on a system

```
sudo sh PRISMA_Server_Install-X.X.X
```

This will run the installer and install `tms` `tms-db` and `tms-mcc` onto the server.

## Structure

The folder here has the following scripts and files:

  - extract.sh
  - install.sh
  - package-prisma-respository
  - update-prisma-repository
  - prisma
  - prisma.list
  - This README.md

### The shell scripts: `extract.sh` and `install.sh`

The two shell scripts, `extract.sh` and `install.sh` are bundled into the tarball package itself and
are run during extraction of the tarball and installation.

`extract.sh` is the script that turns the tarball into the self extracting tarball. It's prepended
to the tarball during creation and it's whats run intitially when you call
`sh PRISMA_Install_<version>`. It will extract the tarball into a local directory, step into that
directory then call `./install.sh` to do the actual installation.

`install.sh` will then install the repository, set permissions, and run the `apt install` to install
the tms server packages.

`extract.sh` will then finish by removing the directory that was created when the tarball was
extracted to clean up.

### Helper Scripts

There are two helper scripts, `package-prisma-repository` and `update-prisma-repository` that are
available.

#### `update-prisma-repository`

`update-prisma-repository` is a script that is currently installed by the installer, but can be run
from anywhere. It will update the debian repository at `/usr/local/prisma-repository` by running a
dpkg-scan on that directory and updating the Packages.tar.gz in that directory. This basically
re-generates the manifest for apt to read that describes all the packages in that directory.

So to add a new package, say a new build of tms debian, just copy the debians into
`/usr/local/prisma-repository` and run `update-prisma-repository` to update.

#### `package-prisma-repository`

This script is not installed anywhere, but when run in this directory will create the executable
tarball. It will create a `./dist` directory where it will copy all the files from this directory
and the repo from `/usr/local/prisma-repository` then create the tarball and prepend the
`extract.sh` to make the tarball executable.

Currently, this script does require the repository to be created and updated in
`/usr/local/prisma-repository`.

### Apt Configuration Files

The two other files in this directory are `prisma` and `prisma.list` that help configure the
repository with `apt` when installed.

`prisma.list` is a file that's installed in `/etc/apt/sources.list.d/`. This file informs apt of the
location of the repository and add its to the list of repos.

`prisma` is installed in `/etc/apt/preferences.d/` and informs apt of any packages that should be
prioritized in the prisma repository. This means that instead of trying to installs packages listed
in this file from external internet sources, apt will prioritize looking in the prisma repository
first and install locally. Without this file, offline installs would fail by finding the package in
multiverse but not being able to install it without the connection to the internet.

!!! todo "TODO: A couple of things would be nice additons to this installer."

    1. We should ask which packages you would like to install so you can pick `tms` `tms-mcc` and
      `tms-db`. This way, on multi-server installs you do not install things you don't need.
      The prompts should probably be something like the following:

    ```
    Do you want to install the PRISMA TMS Application on this server? [Y/n]:
    $ Y
    Do you want to install the PRISMA TMS Database on this server? [Y/n]:
    $ Y
    Do you want this server to receive MCC messages? [Y/n]:
    $ Y
    ```

    2. After installation is complete, do we want to remove the repository from the system to clear
      space?

    3. The repo installation itself might be better suited as a debian itself. This way permissions,
      configs, etc... are handled by the debian installer and we can also make it easier to clear
      the repo if we decide to in step 2 (or make it an option like step 1).
      ```
      Do you want to remove the debian repository after installation? (saves around 200mb of space) [Y/n]
      $ Y
      ```

    4. Remove `update-prisma-repository` from being installed on the system. This is a helper for creating the repository and is not needed on a production system.
