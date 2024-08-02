## TODO:

- test install.sh
- test update.sh
- revise what data to monitor - eg resolution & frame r ate
- revise what parameters to use
- implement auto calibrate on beacon
- improve marker widths
- during install, raspi-config, select System / Audio, choose 1, then reboot
    - or append 'dtparam=audio=off' to boot/config.txt
    - however, I think I've already dealt with this
- more to do in spClient to deal with doubling / fast changes /etc

## Maybe one day

- find a way to run on Bookworm Light
- eg: [Kiosk #1](https://raspberrypi.stackexchange.com/questions/120345/starting-rpi-gui-application-at-boot-without-desktop-gui-and-other-functionaliti)
- eg: [Kiosk #2](https://medium.com/@daddycat/setting-up-raspberry-pi-to-launch-python-gui-app-without-raspbian-desktop-5022a90e5b63)
- eg: [Ultra Minimal Kiosk Setup](https://gist.github.com/seffs/2395ca640d6d8d8228a19a9995418211)
- eg: [cage](https://github.com/cage-kiosk/cage)
- but Bookworm may have other ways - sooner or later
