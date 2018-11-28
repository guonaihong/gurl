package gurlib

import ()

type Message struct {
	In      chan string
	Out     chan string
	InDone  chan string
	OutDone chan string
	K       int
}
