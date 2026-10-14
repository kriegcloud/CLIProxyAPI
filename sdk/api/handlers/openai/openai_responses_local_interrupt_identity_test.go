package openai

import "testing"

func TestLocalHTTPInterruptRequiresActiveResponseIdentity(t *testing.T) {
	local := newResponsesLocalInterrupt()
	local.begin()
	if local.deliver([]byte(`{"response_id":"unknown"}`)) {
		t.Fatal("unidentified turn accepted cancellation")
	}
	local.observe([]byte(`{"type":"response.created","response":{"id":"active"}}`))
	if local.deliver([]byte(`{"response_id":"unknown"}`)) {
		t.Fatal("unrelated response accepted cancellation")
	}
	if !local.deliver([]byte(`{"response_id":"active"}`)) {
		t.Fatal("active response cancellation rejected")
	}
	<-local.framesChan()
	local.observe([]byte(`{"type":"response.completed","response":{"id":"active"}}`))
	if local.deliver([]byte(`{"response_id":"active"}`)) {
		t.Fatal("completed response accepted cancellation")
	}
}
