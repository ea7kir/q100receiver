## TODO:

- test install.sh
- revise what data to monitor - eg resolution & frame r ate
- revise what parameters to use
- implement auto calibrate on beacon
- improve marker widths
- during install, raspi-config, select System / Audio, choose 1, then reboot
    - or append 'dtparam=audio=off' to boot/config.txt
    - however, I think I've already dealt with this
- more to do in spClient to deal with doubling / fast changes /etc

## Find ways to make install easier

- Appearance Steetings / Taskbar
    - Set Taskbar to DSI-1
- Appearance Steetings / Desktop:
    - Set HDMI wallpaper to NoVideo.jpg
    - Disable Documents, Wastebasket and External Disks for HDMI and DSI-1
- Appearance Steetings / Desktop:
    - Set HDMI wallpaper to NoVideo.jpg
    - Disable Documents, Wastebasket and External Disks
- Adjust Volume level to maximum
- Right click Volume and direct audio to HDMI and disable audio jack
- TurnOff Bluetooth
- If updates are available, install updates

## Auto Start
- Currently using wayfire.ini NOT system servoces

## Maybe one day - Bookworm Light

- eg: [Kiosk #1](https://raspberrypi.stackexchange.com/questions/120345/starting-rpi-gui-application-at-boot-without-desktop-gui-and-other-functionaliti)
- eg: [Kiosk #2](https://medium.com/@daddycat/setting-up-raspberry-pi-to-launch-python-gui-app-without-raspbian-desktop-5022a90e5b63)
- eg: [Ultra Minimal Kiosk Setup](https://gist.github.com/seffs/2395ca640d6d8d8228a19a9995418211)
- eg: [cage](https://github.com/cage-kiosk/cage)
- but Bookworm may have other ways - sooner or later
