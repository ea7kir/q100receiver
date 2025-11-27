# A new install procedure

Install Legacy 64-bit with HDMI monitor connected to HDMI-0
and a keyboard and mouse connected.

## Preparation

- login for the first time

```
ssh pi@rxtouch.local
```

- get wayfire working

```
sudo raspi-config
```

- Select 6 A7 W2
- Finish & Reboot

### After rebooting

- login again

```
ssh pi@rxtouch.local
```

- download the app

```
mkdir /home/pi/Q100
cd /home/pi/Q100
git clone https://github.com/ea7kir/q100receiver.git
cd
```

- copy the no video caption

```
sudo cp /home/pi/Q100/q100receiver/etc/NoVideo.jpg /usr/share/rpd-wallpaper
```

### On the HDMI monitor

- Set Bluetooth to Off
- Set audio volume to maximum
- Direct audio to HDMI (right click)

Select Preferences -> Appearance Setting

- Disable Wastebasket
- Disable External Disks
- Change Picture from fisherman.jpg to NoVideo.jpg

Move the taskbar to the the touchscreen

- Taskbar Location: DSI-1

### On the SSH terminal

- run the install script

```
chmod +x /home/pi/Q100/q100receiver/etc/install.sh
/home/pi/Q100/q100receiver/etc/install.sh
```






