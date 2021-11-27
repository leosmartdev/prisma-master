#! /bin/bash

# Setup colors for pretty printing.
RED="\033[0;31m";
GREEN="\033[0;32m";
BLUE="\033[0;34m";
NC="\033[0m";

# Function that will check the exit status of the last command and print DONE or ERROR
# depending on the status. Takes two parameters as well, a message to print on success
# and a second message to print on error. On success message will be printed as is with
# the "DONE" printed in Green. On Error the entire line will be printed in Red.
#
# Usage:
#  command_with_return_value
#  exit_status "Running command...", "Command failed to run correctly"
#
# If $? for command_with_return_value is 0 output will be:
#   Running command...DONE
# If $? for command_with_return_value is not 0 output will be:
#   ERROR Command failed to run correctly
function exit_status {
    if [ $? -eq 0 ]; then
        echo -e "$1${GREEN}DONE${NC}";
    else
        echo -e "${RED}ERROR $2${NC}";
	exit 1;
    fi
}

# First we need to inform apt of where the new repo will be and then make sure that it
# prioritizes packages in that repo.
echo -e "${BLUE}Installing local repository...${NC}";
cp prisma.list /etc/apt/sources.list.d/
exit_status "Adding apt local repository..." "Failed to add prisma.list";
cp prisma /etc/apt/preferences.d/
exit_status "Adding apt preferences..." "Failed to add prisma to preferences.d";
# Install the script that allows us to add new packages to the repo and rebuild the manifest
cp update-prisma-repository /usr/local/sbin
exit_status "Installing repository updater..." "Failed to install repository updater to /usr/local/sbin";
chmod a+x /usr/local/sbin/update-prisma-repository
exit_status "Setting updater permissions..." "Couldn't add execute permissions to updater";
# copy the repo, debians, and manifest to /usr/local/prisma-repository
cp -r prisma-repository /usr/local/
exit_status "Copying repository to /usr/local/prisma-repository..." "Failed to copy local repository to /usr/local/prisma-repository";
chmod 755 /usr/local/prisma-repository
exit_status "Setting repository permissions..." "Couldn't add correct permissions to repository directory";
chmod -R 744 /usr/local/prisma-repository/*
exit_status "Setting repository permissions..." "Couldn't add correct permissions to repository files";
chown -R root:root /usr/local/prisma-repository
exit_status "Setting repository owner" "Couldn't add correct owner to repository directory";

# Run apt-get update to pick up the new packages in the new repo.
echo -e "${BLUE}Updating apt with new PRISMA repository...${NC}";
apt-get update
exit_status "" "apt-get update failed ";

# Run the installer to install tms, tms-mcc, tms-db.
# Here is where in the future we should ask which you want to install, then only install those.
echo -e "${BLUE}Installing PRISMA Server...${NC}";
apt-get install -y tms tms-mcc tms-db
exit_status "PRISMA Server Installation  " "Failed to 'apt-get install tms tms-mcc tms-db'";

echo -e "${BLUE}Cleaning up repository...${NC}";

# Remove repo support files, repository, and repository updater script
rm /etc/apt/sources.list.d/prisma.list
exit_status "Cleaning up apt local repository..." "Failed to remove prisma.list";
rm /etc/apt/preferences.d/prisma
exit_status "Cleaning up apt preferences..." "Failed to remove prisma";
rm -rf /usr/local/prisma-repository
exit_status "Cleaning up local repository folder..." "Failed to remove /usr/local/prisma-repository";
rm /usr/local/sbin/update-prisma-repository
exit_status "Cleaning up repository updater script..." "Failed to remove update-prisma-repository";
