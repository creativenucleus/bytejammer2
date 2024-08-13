package keyboard

import (
	term "github.com/nsf/termbox-go"
)

type Keyboard struct {
	ChUserExitRequest chan bool
	ChKeyPress        chan term.Key
}

func NewKeyboard() *Keyboard {
	return &Keyboard{
		ChUserExitRequest: make(chan bool),
		ChKeyPress:        make(chan term.Key),
	}
}

func (k Keyboard) Start() error {
	err := term.Init()
	if err != nil {
		return err
	}
	defer term.Close()

	for {
		switch event := term.PollEvent(); event.Type {
		case term.EventKey:
			term.Sync()
			if event.Type == term.EventKey {
				if event.Key == term.KeyEsc || event.Key == 3 { // ESC or CTRL+C
					k.ChUserExitRequest <- true
				} else {
					k.ChKeyPress <- event.Key
				}
			}

		case term.EventError:
			return event.Err
		}
	}
}
