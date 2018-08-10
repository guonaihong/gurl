package gurlib

import (
	"time"
)

func ParseTime(t string) (rv time.Duration) {

	t0 := 0
	for k, _ := range t {
		v := int(t[k])
		switch {
		case v >= '0' && v <= '9':
			t0 = t0*10 + (v - '0')
		case v == 's':
			rv += time.Duration(t0) * time.Second
			t0 = 0
		case v == 'm':
			rv += time.Duration(t0*60) * time.Second
			t0 = 0
		case v == 'h':
			rv += time.Duration(t0*60*60) * time.Second
			t0 = 0
		case v == 'd':
			rv += time.Duration(t0*60*60*24) * time.Second
			t0 = 0
		case v == 'w':
			rv += time.Duration(t0*60*60*24*7) * time.Second
			t0 = 0
		case v == 'M':
			rv += time.Duration(t0*60*60*24*7*31) * time.Second
			t0 = 0
		case v == 'y':
			rv += time.Duration(t0*60*60*24*7*31*365) * time.Second
			t0 = 0
		}
	}

	return
}
