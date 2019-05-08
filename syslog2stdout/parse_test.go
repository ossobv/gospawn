package syslog2stdout

import (
	"testing"
)

func TestParseLog(t *testing.T) {
	type inout struct {
		in  []byte
		out string
	}
	list := []inout{
		// Regular message
		{[]byte("<22>May  8 10:02:21 postfix/postfix-script[103]: starting"),
			"mail.info: postfix/postfix-script[103]: starting"},
		// Message with trailing LF and invalid level
		{[]byte("<193>May  8 10:02:22 postfix/postfix-script[104]: starting\n"),
			"unknown.193: postfix/postfix-script[104]: starting"},
		// Message with low ascii
		{[]byte("<0>May  8 10:02:23 a.out: Exception:\r\n\t\x00\x00."),
			"kern.emerg: a.out: Exception:\r\n\t\x00\x00."}, // leave as-is?
		// Message with invalid utf-8
		{[]byte("<79>May  8 10:02:24 Välid Inv\xe4lid (cp-1252)"),
			"cron.debug: Välid Inv\xe4lid (cp-1252)"}, // should be C-escaped?
		// Utter crap
		{[]byte("Whatever!  \n"),
			"unknown.0: Whatever!"},
		// Utter crap, but with a leading level
		{[]byte("<190> Whatever 2!  \n"),
			"local7.info: Whatever 2!"},
	}
	for i := 0; i < len(list); i++ {
		input := list[i].in
		expected := list[i].out
		actual := parseLog(input)
		if actual != expected {
			t.Errorf("#%d: in %q, out %q, expected %q",
				i, input, actual, expected)
		}
	}
}
