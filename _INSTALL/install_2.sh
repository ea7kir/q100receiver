#!/bin/bash

echo
echo "-------------------------------"
echo "-- Installing GIT"
echo "-------------------------------"
echo

sudo apt install git -y

git config --global user.name "ea7kir"
git config --global user.email "mikenaylorspain@icloud.com"
git config --global init.defaultBranch main

echo
echo "-------------------------------"
echo "-- Installing Go"
echo
echo "-- this will take some time..."
echo "-------------------------------"
echo

sudo wget https://go.dev/dl/go1.20.6.linux-arm64.tar.gz
sudo mv go1.20.6.linux-arm64.tar.gz /usr/local/
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.20.6.linux-arm64.tar.gz

echo
echo "-------------------------------"
echo "-- Install gioui dependencies"
echo "-------------------------------"
echo

sudo apt install gcc pkg-config libwayland-dev libx11-dev libx11-xcb-dev libxkbcommon-x11-dev libgles2-mesa-dev libegl1-mesa-dev libffi-dev libxcursor-dev libvulkan-dev

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

