package mail

import (
	"testing"
)

func TestSend(t *testing.T) {
	m := New("/* mail */", "/* password */", "/* host */", "/* port */")
	err := m.Send("/* to */", "Test subject", "<p style='color: blue'>Test body</p>")
	if err != nil {
		t.Fatal(err)
	}
}
