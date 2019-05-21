package color

const (
	Gray = "\x1b[30;1m"
	Blue = "\x1b[34;1m"
	End  = "\x1b[0m"
)

func NewKeyVal(open bool) (keyStart, keyEnd, valStart, valEnd string) {
	if !open {
		return
	}

	return Gray, End, Blue, End
}
