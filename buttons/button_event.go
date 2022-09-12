package buttons

type ButtonEventType int
const (
    EVT_NONE ButtonEventType = iota
    EVT_SINGLE
    EVT_MULTI
    EVT_LONG
    EVT_LONG_LONG
)

type ButtonEvent struct {
    ButtonName  string
    Type        ButtonEventType
    ClickCount  uint8
    RepeatCount uint8
}
