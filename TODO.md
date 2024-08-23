## TODO:

- revise what data to monitor - eg resolution & frame rate
- implement auto calibrate on beacon
- improve marker widths
- more to do in spClient to deal with doubling / fast changes /etc

## Find ways to make install easier

- Appearance Steetings / Taskbar
    - Set Taskbar to DSI-1
- Appearance Steetings / Desktop:
    - Set HDMI wallpaper to NoVideo.jpg
    - Disable Documents, Wastebasket and External Disks for HDMI and DSI-1
- Appearance Setings / Desktop:
    - Set HDMI wallpaper to NoVideo.jpg
    - Disable Documents, Wastebasket and External Disks
- Adjust Volume level to maximum
- Right click Volume and direct audio to HDMI and disable audio jack
- TurnOff Bluetooth

## Auto Start
- Currently using systemctl and NOT wayfire.ini for run at boot
    - because ~/wayland.ini ```[autostart]``` isn't behaving - video appears on touchscreen!

## Maybe one day

- move from Bookworm Desktop to Bookworm Lite or FreeBSD

## Possible ways to disbale bt and wifi

- /boot/firmware/config.txt
	- dtoverlay=disable-wifi
	- dtoverlay=disable-bt
- or maybe
	- sudo systemctl disable bluetooth.service
	- sudo systemctl stop bluetooth.service
- or maybe
	- sudo rfkill block wifi
	- sudo rfkill block bluetooth
- or maybe
	- sudo systemctl disable wpa_supplicant
	- sudo systemctl disable bluetooth
	- sudo systemctl disable hciuart
