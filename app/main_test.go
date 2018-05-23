package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	flags "github.com/jessevdk/go-flags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplication(t *testing.T) {
	app, ctx := prepApp(t, 18080, 500*time.Millisecond)
	go app.Run(ctx)
	time.Sleep(100 * time.Millisecond) // let server start

	resp, err := http.Get("http://localhost:18080/api/v1/ping")
	require.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, "pong", string(body))
	app.Wait()
}

func TestApplicationShutdown(t *testing.T) {
	app, ctx := prepApp(t, 18090, 500*time.Millisecond)
	st := time.Now()
	app.Run(ctx)
	assert.True(t, time.Since(st).Seconds() < 1, "should take about 500msec")
	app.Wait()
}

func prepApp(t *testing.T, port int, duration time.Duration) (*Application, context.Context) {
	// prepare options
	opts := Opts{}
	p := flags.NewParser(&opts, flags.Default)
	p.ParseArgs([]string{"--secret=123456"})
	opts.AvatarStore, opts.BackupLocation = "/tmp", "/tmp"
	opts.BoltPath = fmt.Sprintf("/tmp/%d", port)
	opts.GithubCSEC, opts.GithubCID = "csec", "cid"
	opts.Port = port

	// create app
	app, err := New(opts)
	require.Nil(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(duration)
		log.Print("[TEST] terminate app")
		cancel()
	}()
	return app, ctx
}
