#!/bin/bash
# Install from scratch onto a fresh Pi OS

APPNAME="q100receiver"

# Check current user
whoami | grep -q pi
if [ $? != 0 ]; then
  echo "Install must be performed as user pi"
  exit
fi

echo
echo "EA7KIR Installation Script for $APPNAME"
echo "This script will first prepare the Raspberry Pi OS before"
echo "continuing with the installation of the application."

case "$APPNAME" in
    "q100receiver" )
        GITURL="https://github.com/ea7kir/q100receiver-v3"
        ;;
    "q100transmitter" )
        GITURL="https://github.com/ea7kir/q100transmitter-v1"
        ;;
    "q100server" )
        GITURL="https://github.com/ea7kir/q100server-v1"
        ;;
    * )
        echo
        echo "Error: APPNAME must one of q100receiver, q100transmitter orq100server"
        echo "  Please configure APPNAME correctly"
        exit
        ;;
esac

echo
echo "You are about to install $GETNAME"
echo
echo "During the installation, the Raspberry Pi will restard, so"
echo "please BE PATIENT and wait until you see INSTALLATION COMPLETE"
echo

while true; do
    read -p "Do you wish to continue? (y/n) " yn
    case $yn in
        [Yy]* )
            echo
            echo "-----------------------------------"
            echo "----- Installing $APPNAME -----"
            echo "-----------------------------------"
            echo
        break;;
        [Nn]* )
            echo "Installation of $APPNAME is cancelled"
            exit;;
        * ) echo "Please answer yes or no.";;
    esac
done

#
# https://github.com/BritishAmateurTelevisionClub/ryde-build/blob/master/install_ryde.sh
#
exit

cd
mkdir Q100
cd Q100

# things to considder...

# Update the package manager
echo
echo "------------------------------------"
echo "----- Updating Package Manager -----"
echo "------------------------------------"
echo
sudo dpkg --configure -a
sudo apt-get update --allow-releaseinfo-change

# Uninstall the apt-listchanges package to allow silent install of ca certificates (201704030)
# http://unix.stackexchange.com/questions/124468/how-do-i-resolve-an-apparent-hanging-update-process
sudo apt-get -y remove apt-listchanges

# Upgrade the distribution
echo
echo "-----------------------------------"
echo "----- Performing dist-upgrade -----"
echo "-----------------------------------"
echo
sudo apt-get -y dist-upgrade

sudo apt-get -y install git 

echo
echo "----------------------------------------"
echo "----- Installing Required Packages -----"
echo "----------------------------------------"
echo

#

# Set up the operating system
echo
echo "-------------------------------------------"
echo "----- Setting up the Operating System -----"
echo "--------------------------------------------"
echo

# Set auto login to command line.
sudo raspi-config nonint do_boot_behaviour B2

# Reboot
echo
echo "---------------------------------"
echo "----- INSTALLATION COMPLETE -----"
echo "---------------------------------"
echo
echo "After reboot, log in again."
echo
sleep 1

sudo reboot now
exit
