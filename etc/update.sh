#!/bin/bash

# Update Q100 Receiver on rxtouch.local
# Orignal design by Michael, EA7KIR

GOVERSION=1.22.5

whoami | grep -q pi
if [ $? != 0 ]; then
  echo Update must be performed as user pi
  exit
fi

hostname | grep -q rxtouch
if [ $? != 0 ]; then
  echo Update must be performed on host rxtouch
  exit
fi

while true; do
   read -p "Update q100receiver using Go version $GOVERSION (y/n)? " answer
   case ${answer:0:1} in
       y|Y ) break;;
       n|N ) exit;;
       * ) echo "Please answer yes or no.";;
   esac
done

echo "
###################################################
Update Pi OS
###################################################
"

sudo apt update
sudo apt -y full-upgrade
sudo apt -y autoremove
sudo apt clean

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

sudo apt -y gcc install pkg-config libwayland-dev libx11-dev libx11-xcb-dev libxkbcommon-x11-dev libgles2-mesa-dev libegl1-mesa-dev libffi-dev libxcursor-dev libvulkan-dev

echo "###################################################
Installing gioui tools
###################################################
"

/usr/local/go/bin/go install gioui.org/cmd/gogio@latest

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
UPDATE HAS COMPLETED

    AFTER REBOOTING...

    Login from your PC, Mc, or Linux computer

    ssh pi@rxtouch.local

    and either execute the following commands
    
    cd Q100/q100receiver
    go mod tidy
    go build .
    sudo systemctl enable q100receiver
    sudo systemctl start q100receiver

    or just

    cd Q100/q100receiver
    go mod tidy
    go run .

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
