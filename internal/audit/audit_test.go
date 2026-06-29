package audit

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSink_Notify(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	sink := NewFileSink(path)

	require.NoError(t, sink.Notify(Event{Ts: 100, Metrics: []string{"Alloc"}, IPAddress: "1.2.3.4"}))
	require.NoError(t, sink.Notify(Event{Ts: 200, Metrics: []string{"Frees"}, IPAddress: "5.6.7.8"}))

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	var events []Event
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var e Event
		require.NoError(t, json.Unmarshal(sc.Bytes(), &e))
		events = append(events, e)
	}
	require.Len(t, events, 2)
	assert.Equal(t, int64(100), events[0].Ts)
	assert.Equal(t, []string{"Alloc"}, events[0].Metrics)
	assert.Equal(t, "5.6.7.8", events[1].IPAddress)
}

func TestHTTPSink_Notify(t *testing.T) {
	var received Event
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sink := NewHTTPSink(srv.URL)
	require.NoError(t, sink.Notify(Event{Ts: 1, Metrics: []string{"M"}, IPAddress: "9.9.9.9"}))
	assert.Equal(t, "9.9.9.9", received.IPAddress)
	assert.Equal(t, []string{"M"}, received.Metrics)
}

func TestPublisher_Enabled(t *testing.T) {
	var nilPub *Publisher
	assert.False(t, nilPub.Enabled(), "nil-издатель должен быть отключён")

	pub := NewPublisher()
	assert.False(t, pub.Enabled(), "без приёмников аудит отключён")

	pub.Subscribe(NewFileSink("x"))
	assert.True(t, pub.Enabled(), "после подписки аудит включён")
}

func TestPublisher_NotifyAllObservers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	pub := NewPublisher(NewFileSink(path))

	pub.Notify(Event{Ts: 1, Metrics: []string{"Alloc"}, IPAddress: "1.1.1.1"})

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "\"ip_address\":\"1.1.1.1\"")
}
