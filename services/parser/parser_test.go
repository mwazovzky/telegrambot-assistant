package parser

import (
	"fmt"
	"testing"
)

const success = "\u2713"
const failure = "\u2717"

type TestCase struct {
	desc          string
	botName       string
	botChat       int64
	assignedChats []int64
	msgChat       int64
	msgTxt        string
	expectedTxt   string
	expectedError error
}

func TestParser_Parse(t *testing.T) {
	casesSuccess := []TestCase{
		{
			desc:          "Address bot directly",
			botName:       "Mike",
			botChat:       123,
			assignedChats: []int64{},
			msgChat:       123,
			msgTxt:        "hello",
			expectedTxt:   "hello",
			expectedError: nil,
		},
		{
			desc:          "Address bot via assigned chat",
			botName:       "Mike",
			botChat:       123,
			assignedChats: []int64{789},
			msgChat:       789,
			msgTxt:        "Mike, hello",
			expectedTxt:   "hello",
			expectedError: nil,
		},
	}

	for _, tc := range casesSuccess {
		t.Logf("\t TestCase: %s.", tc.desc)

		p := NewParser(tc.botName, tc.botChat, tc.assignedChats)
		txt, err := p.Parse(tc.msgChat, tc.msgTxt)

		if err == nil {
			t.Logf("\t\t%s It should not return error.", success)
		} else {
			t.Fatalf("\t\t%s It should not return error, got [%s].", success, err)
		}

		if txt == tc.expectedTxt {
			t.Logf("\t\t%s It should parse message.", success)
		} else {
			t.Fatalf("\t\t%s It should parse message, expected [%s], got [%s]", failure, tc.expectedTxt, txt)
		}
	}

	casesError := []TestCase{
		{
			desc:          "Unpocrssable chat",
			botName:       "Mike",
			botChat:       123,
			assignedChats: []int64{},
			msgChat:       456,
			msgTxt:        "hello",
			expectedTxt:   "",
			expectedError: fmt.Errorf("can not process chat"),
		},
		{
			desc:          "Unpocrssable request",
			botName:       "Mike",
			botChat:       123,
			assignedChats: []int64{789},
			msgChat:       789,
			msgTxt:        "hello",
			expectedTxt:   "",
			expectedError: fmt.Errorf("can not process request"),
		},
	}

	for _, tc := range casesError {
		t.Logf("\t TestCase: %s.", tc.desc)

		p := NewParser(tc.botName, tc.botChat, tc.assignedChats)
		_, err := p.Parse(tc.msgChat, tc.msgTxt)

		if err.Error() == tc.expectedError.Error() {
			t.Logf("\t\t%s It should process error.", success)
		} else {
			t.Fatalf("\t\t%s It should process error, expected [%s], got [%s].", success, tc.expectedError, err)
		}
	}
}
