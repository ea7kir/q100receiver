#!/bin/bash

# Install Q100 Receiver on rxtouch.local
# Orignal design by Michael, EA7KIR

GOVERSION=1.22.5

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

echo "
###################################################
Update Pi OS - TODO:
###################################################
"

# sudo apt update
# sudo apt -y full-upgrade
# sudo apt -y autoremove
# sudo apt clean

echo "
###################################################
Making changes to config.txt - TODO:
###################################################
"

#sudo sh -c "echo '\n# EA7KIR Additions' >> /boot/firmware/config.txt"

# Disable Wifi
#sudo sh -c "echo 'dtoverlay=disable-wifi' >> /boot/firmware/config.txt"

# Disable Bluetooth
#sudo sh -c "echo 'dtoverlay=disable-bt' >> /boot/firmware/config.txt"

# EXPERIMENTAL: raspi-config, select System / Audio, choose 1
#sudo sh -c "echo 'dtparam=audio=off' >> /boot/firmware/config.txt"

echo "
###################################################
Making changes to .profile
###################################################
"

sudo sh -c "echo '\n# EA7KIR Additions' >> /home/pi/.profile"

# Disbale Screen Blanking in .profile
# echo -e 'export DISPLAY=:0; xset s noblank; xset s off; xset -dpms' >> /home/pi/.profile

# Adding go path to .profile
echo -e 'export PATH=$PATH:/usr/local/go/bin' >> /home/pi/.profile

echo "
###################################################
Installing Go $GOVERSION
###################################################
"

GOFILE=go$GOVERSION.linux-arm64.tar.gz
cd /usr/local
sudo wget https://go.dev/dl/$GOFILE
sudo tar -C /usr/local -xzf $GOFILE
cd

echo "
###################################################
Installing gioui dependencies
###################################################
"

sudo apt -y install pkg-config libwayland-dev libx11-dev libx11-xcb-dev libxkbcommon-x11-dev libgles2-mesa-dev libegl1-mesa-dev libffi-dev libxcursor-dev libvulkan-dev

echo "
###################################################
Installing gioui tools
###################################################
"

/usr/local/go/bin/go install gioui.org/cmd/gogio@v0.7.1

echo "
###################################################
Install the No Video caption
###################################################
"

sudo cp /home/pi/Q100/q100receiver/etc/NoVideo.jpg /usr/share/rpd-wallpaper

echo "
###################################################
Install longmynd dependencies
###################################################
"

sudo apt -y install libusb-1.0-0-dev libasound2-dev

echo "
###################################################
Cloning longmynd to /home/pi/Q100
###################################################
"

cd /home/pi/Q100
git clone https://github.com/ea7kir/longmynd.git
cd longmynd
make
mkfifo longmynd_main_status
mkfifo longmynd_main_ts
cd

echo "
###################################################
Copying q100receiver.service
###################################################
"

cd /home/pi/Q100/q100receiver/etc
sudo cp q100receiver.service /etc/systemd/system/
sudo chmod 644 /etc/systemd/system/q100receiver.service
sudo systemctl daemon-reload
cd

echo "
###################################################
Making changes to wayfire.ini
###################################################
"

echo "

[output:DSI-1]
mode = 800x480@60000
position = 0,0
transform = normal

[output:HDMI-A-1]
mode = 1920x1080@50000
position = 800,0
transform = normal

" >> ~/.config/wayfire.ini-test

echo "
###################################################
Prevent this script form being executed again
###################################################
"

chmod -x /home/pi/Q100/q100receiver/etc/install.sh

echo "
###################################################
INSTALL HAS COMPLETED
###################################################

    AFTER REBOOTING... AFTER REBOOTING... AFTER REBOOTING... AFTER REBOOTING...

    Cconfigure some Desktop settings:

    Appearance Steetings / Taskbar
        Set Taskbar to DSI-1

    Appearance Steetings / Desktop:
        Set HDMI wallpaper to NoVideo.jpg
        Disable Documents, Wastebasket and External Disks for HDMI and DSI-1

    Appearance Steetings / Desktop:
        Set HDMI wallpaper to NoVideo.jpg
        Disable Documents, Wastebasket and External Disks

    Adjust Volume level to maximum

    Right click Volume and direct audio to HDMI and disable audio jack

    TurnOff Bluetooth

    If updates are available, install then now

    Then login from your PC, Mac, or Linux computer

    ssh pi@rxtouch.local or open VSCODE to RxTouch  ~/Q100/q100receiver/q100reciever

    Now execute the following commands
    
    cd Q100/q100receiver
    go mod tidy
    go build --tags nox11 .

    And execute it with

    ./q100receiver
    
    If all goes well it can be run at boot, by enabling auto run at boot
        sudo systemctl enable q100receiver
        sudo systemctl start q100receiver

    Note: omit the -shutdown flag to prevent a full shutdown if required

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
