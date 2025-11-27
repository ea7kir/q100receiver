#!/bin/bash

# Install Q100 Receiver on rxtouch.local
# Orignal design by Michael, EA7KIR

# CONFIFIGURATION
GOVERSION=1.25.4
GIOUIVERSION=v0.9.0

# nmcli device
# DEVICE         TYPE      STATE                   CONNECTION
# eth0           ethernet  connected               Wired connection 1
# lo             loopback  connected (externally)  lo
# wlan0          wifi      disconnected            --
# p2p-dev-wlan0  wifi-p2p  disconnected            --

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
   read -p "Install q100receiver using Go version $GOVERSION and GIO $GIOUIVERSION (y/n)? " answer
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
# ###################################################
# Disable WiFi (and Bluetooth)
# ###################################################
"

sudo nmcli radio wifi off

# disable bluetooth with dtoverlay=disable-bt
# but there seems to be more to this...
# see: dtoverlay -h disable-bt and elsewhere
# overall, it's probably better to leave it on


echo "
###################################################
Making changes to .profile
###################################################
"

echo -e '\n# EA7KIR Additions' >> /home/pi/.profile

echo -e 'export PATH=$PATH:/usr/local/go/bin' >> /home/pi/.profile

echo -e 'export DISPLAY=:0' >> /home/pi/.profile


echo "
###################################################
Making changes to raspi-config
###################################################
"

# HDMI_0
sudo raspi-config nonint do_audio 72

# disable
sudo raspi-config nonint do_blanking 1

# echo "
# ###################################################
# Installing Xorg/X11
# ###################################################
# "

# sudo apt -y install xorg

# # TODO: confire HDMI-0 as 2nd monitor


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
Installing gioui tools $GIOUIVERSION
###################################################
"

/usr/local/go/bin/go install gioui.org/cmd/gogio@$GIOUIVERSION


echo "
###################################################
Install the No Video caption
###################################################
"

sudo cp /home/pi/Q100/q100receiver/etc/NoVideo.jpg /usr/share/rpd-wallpaper
# TODO: where to install this for Xorg/X11

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
sudo cp minitiouner.rules /etc/udev/rules.d/ # added 28 May 2025
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

# echo "
# ###################################################
# Making changes to wayfire.ini
# ###################################################
# "
#
# echo "
#
# [output:DSI-1]
# mode = 800x480@60000
# position = 0,0
# transform = normal
#
# [output:HDMI-A-1]
# mode = 1920x1080@50000
# position = 800,0
# transform = normal
#
# " >> ~/.config/wayfire.ini-test


# echo "
# ###################################################
# Making changes to .config for libwc
# ###################################################
# "

# # position DSI-1 and HDMI-A-1
# echo "
# profile {
# 		output DSI-1 enable mode 800x480@60.029 position 0,0 transform normal
# 		output HDMI-A-1 enable mode 1920x1080@100.000 position 800,0 transform normal
# }
# " > ~/.config/kanshi/config

# # [panel] DSI-1
# echo "
# [panel]
# position=top
# icon_size=32
# window-list_max_width=200
# monitor=DSI-1
# " > ~/.config/wf-panel-pi/wf-panel-pi.ini

# # wallpaper HDMI-A-1
# echo "
# [*]
# desktop_bg=#D6D3DE
# desktop_shadow=#D6D3DE
# desktop_fg=#E8E8E8
# desktop_font=Nunito Sans Light 12
# wallpaper=/usr/share/rpd-wallpaper/NoVideo.jpg
# wallpaper_mode=crop
# show_documents=0
# show_trash=0
# show_mounts=0
# folder=/home/pi/Desktop
# " > ~/.config/pcmanfm/default/ desktop-items-HDMI-A-1.conf

echo "
###################################################
Prevent this script form being executed again
###################################################
"

chmod -x /home/pi/Q100/q100receiver/etc/install.sh # to prevent it from being run a second time

echo "
###################################################
INSTALL HAS COMPLETED
###################################################

After rebooting, continue with instructions in the README file,

"

while true; do
    read -p "I have read the above, so continue (y/n)? " answer
    case ${answer:0:1} in
        y|Y ) break;;
        n|N ) exit;;
        * ) echo "Please answer yes or no.";;
    esac
done

sudo sync
sudo sync
sudo reboot
