# Q-100 Receiver
### Control and monitor a DATV receiver with a touch screen.
![tx](doc/rx.jpeg)
### REQUIRES Raspberry PI OS (64-BIT) - the Bookworm Desktop version

## Hardware
- Raspberry Pi 4B with 4GB RAM (minimum)
- Raspberry Pi Official 7" Touch Screen
- BATC MiniTiouner v2.0

** A KEYBOARD & MOUSE IS HELPFUL DURING INSTALLATION **

## Connections
- Wired internet connection (not wifi)
- Mount Pi 4B to the Touch Screen, including GPIO power wires
- Connect 5v to MiniTiouner
- Connect 5v to RPi
- Connect MiniTiouner USB to RPi bottom right USB3 (next to RJ45)
- Connect RPi RJ45 to local network

## Installing
**A keyboard and mouse are not required at any time**

### Using Raspberry Pi Imager v1.8.5:
```
CHOOSE Raspberry Pi Device: Raspberry Pi 4 

CHOOSE Operating Sysytem: Raspberry Pi OS (64-bit)

CONFIGURE:
	Set hostname:			rxtouch
	Enable SSH
		Use password authentication
	Set username and password
		Username:			pi
		Password: 			<choose a password>
	Set locale settings
		Time zone:			<Europe/Madrid> # or wherever you are
		Keyboard layout:	<us>
	Eject media when finished
SAVE and WRITE
```

Insert the card into the Raspberry Pi and switch on

WARNING: the Pi may reboot during the install, so please allow it to complete

### Remote login from a Mac, PC or Linux host
```
ssh pi@rxtouch.local

mkdir /home/pi/Q100
cd /home/pi/Q100
git clone https://github.com/ea7kir/q100receiver.git
chmod +x /home/pi/Q100/q100receiver/etc/install.sh
/home/pi/Q100/q100receiver/etc/install.sh
```
### After rebboting
Use your finger to configure some Desktop settings:

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
```
ssh pi@rxtouch.local or open VSCODE to RxTouch  ~/Q100/q100receiver/q100reciever
```
Now execute the following commands
```
cd Q100/q100receiver
go mod tidy
go build --tags nox11 .
```
And execute it with
```
./q100receiver
```
If all goes well it can be run at boot, by enabling auto run at boot
```
sudo systemctl enable q100receiver
sudo systemctl start q100receiver
```
Note: omit the -shutdown flag in the service file to prevent a full shutdown if required

## License
Copyright (c) 2023 Michael Naylor EA7KIR (https://michaelnaylor.es)

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see https://www.gnu.org/licenses/.


[def]: doc/rx.jpeg