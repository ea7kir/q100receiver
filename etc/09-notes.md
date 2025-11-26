# 09-notes fo Xorg/X11
Using pi OS Trixie Lite with Xorg - no desktop environment !

Without any special effort we have be default the following:

output from xrandr -q
```
Screen 0: minimum 320 x 200, current 800 x 480, maximum 7680 x 7680
HDMI-1 connected 720x480+0+0 (normal left inverted right x axis y axis) 527mm x 296mm
   1920x1080    100.00    75.00    60.00    60.00    60.00    50.00    59.94  
   1920x1080i    60.00    59.94  
   1680x1050     59.88  
   1600x900      60.00  
   1280x1024     75.02    60.02  
   1440x900      59.90  
   1366x768      60.00  
   1152x864      59.97  
   1280x720      60.00    50.00    59.94  
   1024x768      75.03    70.07    60.00  
   800x600       72.19    75.00    60.32  
   720x480       60.00*   59.94  
   640x480       75.00    72.81    60.00    59.94  
HDMI-2 disconnected (normal left inverted right x axis y axis)
DSI-1 connected primary 800x480+0+0 (normal left inverted right x axis y axis) 0mm x 0mm
   800x480       60.03*+
```
## Objective
- Make DSI-1 the primary display to host the app GUI @ native 800x480 pixels.
- Make HDMI-1 output ffplay or a caption @ native 1920x1080 pixels 50Hz.

## Guide from
- https://medium.com/@muffwaindan/using-multiple-monitors-with-different-resolutions-on-xorg-linux-4f410dc31be2

## Scaling
```
#!/bin/bash
xrandr --output "DSI-1" --scale 0.9999x0.9999 
xrandr --output "HDMI-1" --scale 0.9999x0.9999 
```