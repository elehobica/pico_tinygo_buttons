# Raspberry Pi Pico TinyGo Buttons
## Overview
This project is a library for handling multiple buttons on Raspberry Pi Pico by TinyGo.
* confirmed with TinyGo 0.22.0

This project features to detect:
* Single Push event
* Repeated Single Push event
* Multiple Push event (exclusive with Repeated Single)
* Long (Long) Push event (exclusive with Repeated Single)

## Supported Board and Device
* Raspberry Pi Pico
* 5 Way Switch + 2 Buttons

## Pin Assignment & Connection
### 5 Way Switch + 2 Buttons
| Pico Pin # | Pin Name | Function | Connection |
----|----|----|----
| 23 | GND | GND | COM |
| 24 | GP18 | GPIO Input | RESET |
| 25 | GP19 | GPIO Input | SET |
| 26 | GP20 | GPIO Input | CENTER |
| 27 | GP21 | GPIO Input | RIGHT |
| 29 | GP22 | GPIO Input | LEFT |
| 31 | GP26 | GPIO Input | DOWN |
| 32 | GP27 | GPIO Input | UP |

## How to build
* Build is confirmed only in TinyGo Docker environment with Windows WSL2 integration
* Before starting docker, clone repository to your local enviroment (by GitBash etc.)
```
> cd /d/somewhere/share
> git clone -b main https://github.com/elehobica/pico_tinygo_buttons.git
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
(# cp -r /share/pico_tinygo_buttons ~/ && cd ~ )

# cd pico_tinygo_buttons
```

* Go Module Configuration
```
# go mod init github.com/elehobica/pico_tinygo_buttons
# go mod tidy
```

* TinyGo Build
```
# tinygo build -target=pico -o pico_tinygo_buttons.uf2

(copy UF2 back to Windows local if working on docker native directory)
(# cp pico_tinygo_buttons.uf2 /share/pico_tinygo_buttons/ )
```

* Put UF2 

Then, go back to Windows environment and put "pico_tinygo_buttons.uf2" on RPI-RP2 drive

## Usage Guide
### Button Function Assignment
* Multiple / Long / LongLong detection for Center button
* Single / Repeated Single detection for Left/Right/Up/Down buttons
* Single detection for Set/Reset buttons

### Note
* If Multiple detection enabled, time lag defined by 'actFinishCnt' is needed to determine action
* Repeat count information of Repeated Single detection is served for UI items to accelerate something by continuous button push
* Use NewButtonConfig() for customizing button detection parameters other than default configurations
* Trriple clicks of Center button shows processing time of button scan function (in this example project)

### Log Example
```
=========================
== pico_tinygo_buttons ==
=========================
center: 1
left: 1
right: 1
up: 1
up: 1
right: 1
down: 1
down: 1 (Repeated 1)
down: 1 (Repeated 2)
center: 2
set: 1
reset: 1
center: Long
center: LongLong
center: 3
time 41us (scan: 650)
```