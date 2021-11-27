#!/bin/bash

RED="\033[0;31m";
BLUE="\033[0;34m";
NC="\033[0m";

if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}This script must be run as root${NC}"
   exit 1
fi

echo -e "${BLUE}Extracting installer bundle into `pwd`${NC}"
# searches for the line number where finish the script and start the tar.gz
SKIP=`awk '/^__TARFILE_FOLLOWS__/ { print NR + 1; exit 0; }' $0`
#remember our file name
THIS=`pwd`/$0
# take the tarfile and pipe it into tar
tail -n +$SKIP $THIS | tar -xz
# Any script here will happen after the tar file extract.
echo -e "${BLUE}Running PRISMA Server Installation Script${NC}"
cd PRISMA_Server_Install
./install.sh
cd ../
rm -rf PRISMA_Server_Install
exit 0
# NOTE: Don't place any newline characters after the last line below.
__TARFILE_FOLLOWS__
