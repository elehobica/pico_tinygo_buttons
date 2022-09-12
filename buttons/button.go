package buttons

type Pin interface {
    Get() bool
}

type Button struct {
    name     string
    pin      Pin
    config   *ButtonConfig
    history  historyType
    filtered historyType
}

func NewButton(name string, pin Pin, config *ButtonConfig) *Button {
    button := Button {
        name: name,
        pin: pin,
        config: config,
        history: newHistory(false),
        filtered: newHistory(false),
    }
    return &button
}
