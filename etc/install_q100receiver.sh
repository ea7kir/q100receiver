#!/bin/bash

echo
echo "-------------------------------"
echo "-- Updateing the OS"
echo "-------------------------------"
echo

sudo apt update
sudo apt full-upgrade -y
sudo apt autoremove -y
sudo apt clean

echo
echo "-------------------------------"
echo "-- Updating eeprom firmware"
echo "-------------------------------"
echo

sudo rpi-eeprom-update -a

echo
echo "-------------------------------"
echo "-- Updating eeprom firmware"
echo "-------------------------------"
echo

# NOTE: only if advised to do so
# sudo rpi-update

echo
echo "-------------------------------"
echo "-- running rfkill"
echo "-------------------------------"
echo

rfkill block 0
rfkill block 1

echo
echo "-------------------------------"
echo "-- Setting .profile"
echo "-------------------------------"
echo

echo -e '\n\nexport PATH=$PATH:/usr/local/go/bin\n\n' >> /home/pi/.profile

echo
echo "-------------------------------"
echo "-- Updating eeprom firmware"
echo "-------------------------------"
echo

sudo rpi-update

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
echo "-- Install longmynd"
echo "-------------------------------"
echo

mkdir /home/pi/Q100
cd /home/pi/Q100
git clone https://github.com/ea7kir/longmynd.git
cd longmynd
sudo apt install make gcc libusb-1.0-0-dev libasound2-dev
Make
mkfifo longmynd_main_status
mkfifo longmynd_main_ts

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
echo
echo "     connect an HDMI monitor"
echo "     and reboot"
echo 
echo "-------------------------------"
echo
