#!/bin/bash

sudo apt update
sudo apt -y full-upgrade
sudo apt -y autoremove
sudo apt clean

###################################################

GOVERSION=1.22.5

echo Installing Go $GOVERSION
GOFILE=go$GOVERSION.linux-arm64.tar.gz
cd /usr/local
sudo rm -rf go
sudo wget https://go.dev/dl/$GOFILE
sudo tar -C /usr/local -xzf $GOFILE
cd /home/pi/Q100/q100receiver
go mod tidy
go version
