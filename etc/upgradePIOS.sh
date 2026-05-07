#!/bin/bash

echo "
###################################################
Upgrade Pi OS
###################################################
"

echo Update Pi OS
sudo apt update
sudo apt -y upgrade
sudo apt -y autoremove
sudo apt clean
