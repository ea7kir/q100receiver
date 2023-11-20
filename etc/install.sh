#!/bin/bash

# Install Q100 Receiver on Raspberry Pi 4
# Orignal design by Michael, EA7KIR

GOVERSION=1.21.4

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

echo "\n###################################################\n"

echo Updateing Pi OS
sudo apt update
sudo apt -y full-upgrade
sudo apt -y autoremove
sudo apt clean

echo "\n###################################################\n"

echo Running rfkill # not sure if this dupicates config.txt
rfkill block 0
rfkill block 1

echo "\n###################################################\n"

echo Making changes to config.txt

echo Disbaling Wifi
echo -e "\ndtoverlay=disable-wifi" >> /boot/config.txt

echo Disbaling Bluetooth
echo -e "\ndtoverlay=disable-bt" >> /boot/config.txt

echo EXPERIMENTAL: raspi-config, select System / Audio, choose 1
echo -e "\ndtparam=audio=off" >> /boot/config.txt

echo "\n###################################################\n"

echo Adding go path to .profile
echo -e '\n\nexport PATH=$PATH:/usr/local/go/bin\n\n' >> /home/pi/.profile

echo Installing Go $GOVERSION
GOFILE=go$GOVERSION.linux-arm64.tar.gz
cd /usr/local
sudo wget https://go.dev/dl/$GOFILE
sudo tar -C /usr/local -xzf $GOFILE
cd

echo "\n###################################################\n"

echo Installing gioui dependencies
sudo apt install gcc pkg-config libwayland-dev libx11-dev libx11-xcb-dev libxkbcommon-x11-dev libgles2-mesa-dev libegl1-mesa-dev libffi-dev libxcursor-dev libvulkan-dev

echo Installing gioui tools
go install gioui.org/cmd/gogio@latest

echo "\n###################################################\n"

echo Install the No Video caption
sudo cp /home/pi/Q100/q100receiver/etc/NoVideo.jpg /usr/share/rpd-wallpaper

echo "\n###################################################\n"

echo Install longmynd dependencies
sudo apt install make gcc libusb-1.0-0-dev libasound2-dev

echo "\n###################################################\n"

echo Cloning longmynd to /home/pi/Q100
cd /home/pi/Q100
git clone https://github.com/ea7kir/longmynd.git
cd longmynd
make
mkfifo longmynd_main_status
mkfifo longmynd_main_ts
cd

echo "\n###################################################\n"

echo Copying q100receiver.service
cd /home/pi/Q100/q100receiver/etc
sudo cp q100receiver.service /etc/systemd/system/
sudo chmod 644 /etc/systemd/system/q100receiver.service
sudo systemctl daemon-reload
cd

echo "\n###################################################\n"

echo "
INSTALL HAS COMPLETED
   after rebooting, build and auto exec...

   cd Q100/q100receiver
   go mod tidy
   go build .
   sudo systemctl enable q100receiver
   sudo systemctl start q100receiver

"

while true; do
    read -p "I have read the above, so continue (y/n)? " answer
    case ${answer:0:1} in
        y|Y ) break;;
        n|N ) exit;;
        * ) echo "Please answer yes or no.";;
    esac
done

sudo reboot
