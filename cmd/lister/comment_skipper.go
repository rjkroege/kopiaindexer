package main

import (
	"bufio"
)

const (
	sklStartOrNl = 0
	sklOther     = 1
)

func SkipCommentLines(rd *bufio.Reader) error {
	state := sklStartOrNl
	for {
		r, _, err := rd.ReadRune()
		if err != nil {
			// TODO(rjk): Consider better error annotations?
			return err
		}

		switch state {
		case sklStartOrNl:
			if r == '{' {
				rd.UnreadRune()
				return nil
			} else {
				state = sklOther
			}
		case sklOther:
			if r == '\n' {
				state = sklStartOrNl
			}
		}
	}
}
