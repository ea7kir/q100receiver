# Note
The Makefile has been modified from the original.

Line #22
```
COPT_RPI34 = -mfpu=neon-fp-armv8
```
Commented out as follows
```
# <- REMOVED TO COMPILE ON BULLSEYE 64-BIT -> COPT_RPI34 = -mfpu=neon-fp-armv8
```
This may not be the besy way, buit it works for me.
