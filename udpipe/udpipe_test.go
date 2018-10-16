package udpipe

import "testing"

func TestUdpipe(t *testing.T) {
	text := "Bob brings pizza to Alice."
	result, e := Parse(text)
	if e!=nil{
		t.Fatal(e)
	}
	t.Logf("Result: [%#v]\n", result)
}
