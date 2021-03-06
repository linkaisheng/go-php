// Copyright 2016 Alexander Palaistras. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.

package engine

import (
	"bytes"
	"testing"
)

func TestReceiverStart(t *testing.T) {
	e, _ = New()
	t.SkipNow()
}

type testReceiver struct {
	Var    string
	hidden int64
}

func (t *testReceiver) Ignore() {
}

func (t *testReceiver) Hello(p string) string {
	return "Hello " + p
}

func (t *testReceiver) Goodbye(p string) (string, string) {
	return "Goodbye", p
}

func (t *testReceiver) invalid() string {
	return "I'm afraid I can't let you do that, Dave"
}

func newTestReceiver(args []interface{}) interface{} {
	value := "Foo"

	if len(args) > 0 {
		switch v := args[0].(type) {
		case bool:
			return nil
		case string:
			value = v
		}
	}

	return &testReceiver{Var: value, hidden: 42}
}

var newReceiverTests = []struct {
	script   string
	expected string
}{
	{"$t = new TestReceiver; echo is_object($t);", "1"},
	{`try {
		$t = new TestReceiver(false);
	} catch (Exception $e) {
		echo $e->getMessage();
	}`, "Failed to instantiate method receiver"},

	{"$t = new TestReceiver; echo $t->Var;", "Foo"},
	{"$t = new TestReceiver; echo $t->hidden;", ""},
	{"$t = new TestReceiver('wow'); echo $t->Var;", "wow"},

	{"$t = new TestReceiver; $t->Var = 'Bar'; echo $t->Var;", "Bar"},
	{"$t = new TestReceiver; $t->hello = 'wow'; echo $t->hello;", ""},

	{"$t = new TestReceiver; echo $t->Ignore();", ""},
	{"$t = new TestReceiver; echo $t->Hello('World');", "Hello World"},
	{"$t = new TestReceiver; echo json_encode($t->Goodbye('Doge'));", `["Goodbye","Doge"]`},
	{"$t = new TestReceiver; echo $t->invalid();", ""},

	{"$t = new TestReceiver; echo ($t->Var) ? 1 : 0;", "1"},
	{"$t = new TestReceiver; echo isset($t->Var) ? 1 : 0;", "1"},
	{"$t = new TestReceiver; echo empty($t->Var) ? 1 : 0;", "0"},

	{"$t = new TestReceiver; echo isset($t->hidden) ? 1 : 0;", "0"},
}

func TestNewReceiver(t *testing.T) {
	var w bytes.Buffer

	c, _ := NewContext()
	c.Output = &w

	r, err := NewReceiver("TestReceiver", newTestReceiver)
	if err != nil {
		t.Fatalf("NewReceiver(): Failed to define method receiver: %s", err)
	}

	for _, tt := range newReceiverTests {
		_, err := c.Eval(tt.script)
		if err != nil {
			t.Errorf("Context.Eval('%s'): %s", tt.script, err)
			continue
		}

		actual := w.String()
		w.Reset()

		if actual != tt.expected {
			t.Errorf("Context.Eval('%s'): Expected output '%s', actual '%s'", tt.script, tt.expected, actual)
		}
	}

	r.Destroy()
	c.Destroy()
}

func TestReceiverDestroy(t *testing.T) {
	c, _ := NewContext()
	defer c.Destroy()

	r, err := NewReceiver("TestReceiver", newTestReceiver)
	if err != nil {
		t.Fatalf("NewReceiver(): Failed to define method receiver: %s", err)
	}

	r.Destroy()
	if r.create != nil || r.objects != nil {
		t.Errorf("Receiver.Destroy(): Did not set internal fields to `nil`")
	}

	// Attempting to destroy a receiver twice should be a no-op.
	r.Destroy()
}

func TestReceiverEnd(t *testing.T) {
	e.Destroy()
	t.SkipNow()
}
