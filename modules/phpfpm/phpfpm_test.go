package phpfpm

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/netdata/go-orchestrator/module"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testStatusJSON, _     = ioutil.ReadFile("testdata/status.json")
	testStatusFullJSON, _ = ioutil.ReadFile("testdata/status-full.json")
	testStatusText, _     = ioutil.ReadFile("testdata/status.txt")
	testStatusFullText, _ = ioutil.ReadFile("testdata/status-full.txt")
)

func TestNew(t *testing.T) {
	job := New()

	assert.Implements(t, (*module.Module)(nil), job)
	assert.Equal(t, "http://127.0.0.1/status?full&json", job.UserURL)
	assert.Equal(t, time.Second, job.Timeout.Duration)
}

func TestPhpfpm_Init(t *testing.T) {
	job := New()

	got := job.Init()

	require.True(t, got)
	assert.NotNil(t, job.client)
}

func TestPhpfpm_Check(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(testStatusText)
			}))
	defer ts.Close()

	job := New()
	job.UserURL = ts.URL
	job.Init()
	require.True(t, job.Init())

	got := job.Check()

	assert.True(t, got)
}

func TestPhpfpm_CheckReturnsFalseOnFailure(t *testing.T) {
	job := New()
	job.UserURL = "http://127.0.0.1:38001/us"
	require.True(t, job.Init())

	got := job.Check()

	assert.False(t, got)
}

func TestPhpfpm_Charts(t *testing.T) {
	job := New()

	got := job.Charts()

	assert.NotNil(t, got)
}

func TestPhpfpm_CollectJSON(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(testStatusJSON)
			}))
	defer ts.Close()

	job := New()
	job.UserURL = ts.URL + "/?json"
	require.True(t, job.Init())

	got := job.Collect()

	want := map[string]int64{
		"active":    1,
		"idle":      1,
		"maxActive": 1,
		"reached":   0,
		"requests":  21,
		"slow":      0,
	}
	assert.Equal(t, want, got)
}

func TestPhpfpm_CollectJSONFull(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(testStatusFullJSON)
			}))
	defer ts.Close()

	job := New()
	job.UserURL = ts.URL + "/?json"
	require.True(t, job.Init())

	got := job.Collect()

	want := map[string]int64{
		"active":    1,
		"idle":      1,
		"maxActive": 1,
		"reached":   0,
		"requests":  22,
		"slow":      0,
		"minReqCpu": 0,
		"maxReqCpu": 10,
		"avgReqCpu": 5,
		"minReqDur": 834,
		"maxReqDur": 919,
		"avgReqDur": 876,
		"minReqMem": 2093045,
		"maxReqMem": 2097152,
		"avgReqMem": 2095098,
	}
	assert.Equal(t, want, got)
}

func TestPhpfpm_CollectText(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(testStatusText)
			}))
	defer ts.Close()

	job := New()
	job.UserURL = ts.URL
	require.True(t, job.Init())

	got := job.Collect()

	want := map[string]int64{
		"active":    1,
		"idle":      1,
		"maxActive": 1,
		"reached":   0,
		"requests":  19,
		"slow":      0,
	}
	assert.Equal(t, want, got)
}

func TestPhpfpm_CollectTextFull(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(testStatusFullText)
			}))
	defer ts.Close()

	job := New()
	job.UserURL = ts.URL
	require.True(t, job.Init())

	got := job.Collect()

	want := map[string]int64{
		"active":    1,
		"idle":      1,
		"maxActive": 1,
		"reached":   0,
		"requests":  20,
		"slow":      0,
		"minReqCpu": 0,
		"maxReqCpu": 10,
		"avgReqCpu": 5,
		"minReqDur": 536,
		"maxReqDur": 834,
		"avgReqDur": 685,
		"minReqMem": 2093045,
		"maxReqMem": 2097152,
		"avgReqMem": 2095098,
	}
	assert.Equal(t, want, got)
}

func TestPhpfpm_CollectReturnsNothingWhenInvalidData(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("hello and goodbye\nfrom someone\nfoobar"))
			}))
	defer ts.Close()

	job := New()
	job.UserURL = ts.URL
	require.True(t, job.Init())

	got := job.Collect()

	assert.Len(t, got, 0)
}

func TestPhpfpm_CollectReturnsNothingWhenEmptyData(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte{})
			}))
	defer ts.Close()

	job := New()
	job.UserURL = ts.URL
	require.True(t, job.Init())

	got := job.Collect()

	assert.Len(t, got, 0)
}

func TestPhpfpm_CollectReturnsNothingWhenBadStatusCode(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}))
	defer ts.Close()

	job := New()
	job.UserURL = ts.URL
	require.True(t, job.Init())

	got := job.Collect()

	assert.Len(t, got, 0)
}

func TestPhpfpm_Cleanup(t *testing.T) {
	New().Cleanup()
}
