# Q-100 Receiver
Using Go and Gio.\
Running on Pi OS 64bit (the default X11 desktop version).

Wayland causes the app to crash when the screen is touched.

## Links
- [Gio Home](https://gioui.org)
- [Gio Getting Started](https://gioui.org/doc/learn/get-started)
- [Gio Mailing List](https://lists.sr.ht/~eliasnaur/gio)
- [Gio Issue Tracker](https://todo.sr.ht/~eliasnaur/gio)
- [Gio Issue Email](mailto:~eliasnaur/gio@todo.sr.ht)
- [Gio Gophers Slack](https://app.slack.com/client/T029RQSE6/CM87SNCGM/rimeto_profile/U03QX2K2473)
- [Gio Youtub Calls](https://www.youtube.com/channel/UCzuKUnKK5gAFJKNyA1imIHw)
- [Go Documentation](https://go.dev/doc/)
- [Go Packages](https://pkg.go.dev)
- [Go Forum](https://forum.golangbridge.org)

## Prepare
```
bash -c "echo -e '\nexport PATH=\$PATH:/usr/local/go/bin\n' >> .profile"
```

## Config Pi OS
```
sudo apt update; sudo apt -y full-upgrade
sudo rpi-eeprom-update -a
sudo reboot
```

## Install Go
```
cd /usr/local
sudo wget https://go.dev/dl/go1.20.5.linux-arm64.tar.gz
sudo tar -C /usr/local -xzf go1.20.5.linux-arm64.tar.gz
go version
```

## Install Gio
```
sudo apt install gcc pkg-config libwayland-dev libx11-dev libx11-xcb-dev libxkbcommon-x11-dev libgles2-mesa-dev libegl1-mesa-dev libffi-dev libxcursor-dev libvulkan-dev
```

## Gio run options
You can build Gio programs without X11 support with the nox11 build tag:
```
go run --tags nox11 .
```
To build Gio programs without Wayland support use nowayland build tag:
```
go run --tags nowayland .
```

## Future options
- [Kiosk #1](https://raspberrypi.stackexchange.com/questions/120345/starting-rpi-gui-application-at-boot-without-desktop-gui-and-other-functionaliti)
- [Kiosk #2](https://medium.com/@daddycat/setting-up-raspberry-pi-to-launch-python-gui-app-without-raspbian-desktop-5022a90e5b63)

## License

Copyright (c) 2023 Michael Naylor EA7KIR (https://michaelnaylor.es)

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see https://www.gnu.org/licenses/.
