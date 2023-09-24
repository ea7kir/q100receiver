#!/bin/bash

# Install Q100 Receiver on Raspberry Pi 4
# Orignal design by Michael, EA7KIR

GOVERSION=1.21.0

echo WARNING: THIS INSTALL SCRIPT HAS NOT BEEN TESTED

whoami | grep -q pi
if [ $? != 0 ]; then
  echo Install must be performed as user pi
  exit
fi

hostname | grep -q rxtouch
if [ $? != 0 ]; then
  echo Install must be performed on host rxtouch
  exit
fi

while true; do
    read -p "Install q100receiver using Go version $GOVERSION (y/n)? " answer
    case ${answer:0:1} in
        y|Y ) break;;
        n|N ) exit;;
        * ) echo "Please answer yes or no.";;
    esac
done

mkdir /home/pi/Q100

echo Updateing Pi OS
sudo apt update
sudo apt -y full-upgrade
sudo apt -y autoremove
sudo apt clean

echo Running rfkill # not sure if this dupicates config.txt
rfkill block 0
rfkill block 1

echo Making changes to config.txt

echo Disbaling Wifi
echo -e "\ndtoverlay=disable-wifi" >> /boot/config.txt

echo Disbaling Bluetooth
echo -e "\ndtoverlay=disable-bt" >> /boot/config.txt

echo Installing GIT
sudo apt -y install git

echo Adding go path to .profile
echo -e '\n\nexport PATH=$PATH:/usr/local/go/bin\n\n' >> /home/pi/.profile

echo Installing Go $GOVERSION
GOFILE=go$GOVERSION.linux-arm64.tar.gz
cd /usr/local
sudo wget https://go.dev/dl/$GOFILE
# sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf $GOFILE
cd

echo Installing gioui dependencies
sudo apt install gcc pkg-config libwayland-dev libx11-dev libx11-xcb-dev libxkbcommon-x11-dev libgles2-mesa-dev libegl1-mesa-dev libffi-dev libxcursor-dev libvulkan-dev

echo Cloning q100receiver to /home/pi/Q100
cd /home/pi/Q100
git clone https://github.com/ea7kir/q100receiverr.git
cd

echo Install the No Video caption
sudo cp /home/pi/Q100/q100receiver/etc/NoVideo.jpg /usr/share/rpd-wallpaper

echo Cloning longmynd to /home/pi/Q100
cd /home/pi/Q100
git clone https://github.com/ea7kir/longmynd.git
cd longmynd
make
mkfifo longmynd_main_status
mkfifo longmynd_main_ts
cd

echo Copying q100receiver.service
cd /home/pi/Q100/etc
sudo cp q100receiver.service /etc/systemd/system/
sudo chmod 644 /etc/systemd/system/q100receiver.service
sudo systemctl daemon-reload
cd

echo "\n
INSTALL HAS COMPLETED
   after rebooting, build and auto exec...

   cd Q100/q100receiver
   go mod tidy
   go build .
   sudo systemctl enable q100receiver
   sudo systemctl start q100receiver

   now type sudo reboot
"
