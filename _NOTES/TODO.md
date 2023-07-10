# TODO:

## lmreader.go
```
Got modcod 31 greater than 28,
ERROR in mode_margin() when key is DVB-S2 ? ?
```

```
only slices are sorted
```

```
stopping and stating ffplay when received signals disappears

RxTouch stops/starts ffplay based on has_dvb - ie liveData.State == "Locked"
RxTouch resets liveDatat when not longmynd_running
IE. ffplay is controlled by within my longmynd reader function

Use longmynd from https://github.com/philcrump/longmynd

so it can be controlled over a websocket.

I think VLC can be restarted from this version.

```

## calibrater.go / fftreader.go

move to own namespace - package
