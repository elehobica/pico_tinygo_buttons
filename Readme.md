# Raspberry Pi Pico TinyGo Timer Alarm
## Overview
This project is an implementation of Timer/Alarm on Raspberry Pi Pico by TinyGo.
* confirmed with TinyGo 0.22.0

This project features:
* T.B.D.

## Supported Board
* Raspberry Pi Pico

## How to build
* Build is confirmed only in TinyGo Docker environment with Windows WSL2 integration
* Before starting docker, clone repository to your local enviroment (by GitBash etc.)
```
> cd /d/somewhere/share
> git clone -b main https://github.com/elehobica/pico_tinygo_timer_alarm.git
```

* Docker
```
> wsl
(in WSL2 shell)
$ docker pull docker pull tinygo/tinygo
$ docker images
$ docker run -it -v /mnt/d/somewhere/share:/share tinygo/tinygo:latest /bin/bash
(in docker container)
# cd /share

(copy repository for docker native directory for best performance of WSL2, otherwise stay /share)
(# cp -r /share/pico_tinygo_timer_alarm ~/ && cd ~ )

# cd pico_tinygo_timer_alarm
```

* Go Module Configuration
```
# go mod init github.com/elehobica/pico_tinygo_timer_alarm
# go mod tidy
```

* TinyGo Build
```
# tinygo build -target=pico -o pico_tinygo_timer_alarm.uf2

(copy UF2 back to Windows local if working on docker native directory)
(# cp pico_tinygo_timer_alarm.uf2 /share/pico_tinygo_timer_alarm/ )
```

* Put UF2 

Then, go back to Windows environment and put "pico_tinygo_timer_alarm.uf2" on RPI-RP2 drive