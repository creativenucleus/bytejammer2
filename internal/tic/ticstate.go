package tic

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

var (
	ErrInvalidData = errors.New("invalid data")
)

type State struct {
	Code      []byte
	IsRunning bool
	CursorX   int
	CursorY   int
}

func MakeTicStateRunning(code []byte) State {
	return State{
		Code:      code,
		IsRunning: true,
	}
}

func MakeTicStateEditor(code []byte, cursorX int, cursorY int) State {
	return State{
		Code:    code,
		CursorX: cursorX,
		CursorY: cursorY,
	}
}

// Matches the pos command code
//var regexTicData = regexp.MustCompile(`(?s)^-- pos: (\d+),(\d+)\n(.*)$`)

var regexTicData = regexp.MustCompile(`(?s)^-- pos: (\d+),(\d+)\r?\n(.*)$`)

func MakeTicStateFromExportData(data []byte) (*State, error) {
	ts := State{}

	//	data = []byte("-- pos: 10,1\nfunction TIC()\n hello\n\thello\nend")

	matches := regexTicData.FindStringSubmatch(string(data))
	if matches == nil || len(matches) != 4 {
		return nil, ErrInvalidData
	}

	ts.IsRunning = (matches[1] == "0") && (matches[2] == "0")
	if !ts.IsRunning {
		var err error
		ts.CursorX, err = strconv.Atoi(matches[1])
		if err != nil {
			return nil, err
		}
		ts.CursorY, err = strconv.Atoi(matches[2])
		if err != nil {
			return nil, err
		}
	}

	ts.Code = []byte(matches[3])

	return &ts, nil
}

func (ts State) GetCode() []byte {
	return ts.Code
}

func (ts *State) SetCode(code []byte) {
	ts.Code = code
}

func (ts State) GetIsRunning() bool {
	return ts.IsRunning
}

func (ts State) GetCursorX() int {
	return ts.CursorX
}

func (ts State) GetCursorY() int {
	return ts.CursorY
}

// Adds the control string
// This is --pos: 0,0 if running, otherwise --pos: X,Y (the cursor position)
func (ts State) MakeDataToImport() ([]byte, error) {
	controlString := "-- pos: 0,0\n" // Running
	if !ts.IsRunning {
		controlString = fmt.Sprintf("-- pos: %d,%d\n", ts.CursorX, ts.CursorY)
	}

	return append([]byte(controlString), ts.Code...), nil
}

func (ts1 State) IsEqual(ts2 State) bool {
	return bytes.Equal(ts1.Code, ts2.Code) &&
		ts1.IsRunning == ts2.IsRunning && ts1.CursorX == ts2.CursorX && ts1.CursorY == ts2.CursorY
}

func CodeAddAuthorShim(code []byte, author string) []byte {
	//	shim := CodeReplace(embed.LuaAuthorShim, map[string]string{"DISPLAY_NAME": author})
	//	return append(code, shim...)
	return []byte{}
}
