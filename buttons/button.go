package buttons

import (
    "machine"
)

type Button struct {
    name     string
    pin      *machine.Pin
    config   *ButtonConfig
    history  historyType
    filtered historyType
}

func NewButton(name string, pin *machine.Pin, config *ButtonConfig) *Button {
    button := Button {
        name: name,
        pin: pin,
        config: config,
        history: newHistory(false),
        filtered: newHistory(false),
    }
    button.pin.Configure(machine.PinConfig{Mode: machine.PinInput})
    return &button
}
