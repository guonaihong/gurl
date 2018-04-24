package gurlib

type For struct {
	Range string `json:"range,omitempty"`
	K     string `json:"k,omitempty"`
	V     string `json:"v,omitempty"`
	Base  `json:"-"`
}
