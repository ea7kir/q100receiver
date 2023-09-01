#!/bin/bash

echo
echo "-------------------------------"
echo "-- Installing GIT"
echo "-------------------------------"
echo

sudo apt install git -y

echo
echo "-------------------------------"
echo "-- Installing Go"
echo "-------------------------------"
echo

GOVERSION=go1.21.0.linux-arm64.tar.gz
cd /usr/local
sudo wget https://go.dev/dl/$GOVERSION
# sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf $GOVERSION
cd

echo
echo "-------------------------------"
echo "-- Install gioui"
echo "-------------------------------"
echo

sudo apt install gcc pkg-config libwayland-dev libx11-dev libx11-xcb-dev libxkbcommon-x11-dev libgles2-mesa-dev libegl1-mesa-dev libffi-dev libxcursor-dev libvulkan-dev
#go mod edit -dropreplace gioui.org
go get gioui.org@latest

echo
echo "-------------------------------"
echo "-- Install No Video"
echo "-------------------------------"
echo

sudo cp /home/pi/Q100/q100receiver/_INTSTALL/NoVideo.jpg /usr/share/rpd-wallpaper

echo
echo "-------------------------------"
echo "-- Install Service"
echo "-------------------------------"
echo

sudo cp /home/pi/Q100/q100receiver/_INTSTALL/q100receiver.service /etc/systemd/system/
sudo chmod 644 /etc/systemd/system/q100receiver.service
sudo systemctl daemon-reload
echo "To auto start / stop, etc.."
echo "sudo systemctl enable q100receiver"
echo "sudo systemctl start q100receiver"
echo "sudo systemctl status q100receiver"
echo "sudo systemctl stop q100receiver"
echo "sudo systemctl disable q100receiver"

echo
echo "-------------------------------"
echo "-- Done"
echo "-------------------------------"
echo

echo "Clone q100receiver from within VSCODE"
echo "using: https://github.com/ea7kir/q100receiver.git"
echo
echo "To run q100receiver, type: ./q100receiver"

sleep 5

sudo reboot

