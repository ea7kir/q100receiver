# Installing Pi OS

NOTE: CURRENTLY REQUIRES PI OS BULLSEYE 64-BIT (FULL DESKTOP VERSION)

## Using Raspberry Pi Imager:

```
CHOOSE OS: Raspberry Pi OS (other) -> Raspberry Pi OS (64-bit)

CONFIGURE:
	Set hostname:			q100receiver
	Enable SSH
		Use password authentication
	Set username and password
		Username:			pi
		Password: 			<password>
	Set locale settings
		Time zone:			<Europe/Madrid>
		Keyboard layout:	<us>
	Eject media when finished
SAVE and WRITE
```

Insert the card into the Raspberry Pi and switch on

WARNING: the Pi may reboot during the install, so please allow it to complete

## Remote login from a Mac, PC or Linux host
```
ssh pi@q100receiver.local
```
and execute the following commands:
```
mkdir /home/pi/Q100
cd /home/pi/Q100
git clone https://github.com/ea7kir/q100receiver.git
cd /home/pi/Q100/q100receiver/_INSTALL
./install_q100receiver.sh
```
When DONE, reboot the Pi.
