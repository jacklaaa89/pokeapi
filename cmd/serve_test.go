package cmd

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const expectedPort = 5555

// checkServe used to allow us to call the TestRoot_Serve test again in a separate thread.
var checkServe = flag.Bool("check_serve", false, "if true, TestRoot_Serve will fail if SIGINT is not received.")

// TestRoot_Serve this test serves to check that
// we can:
// - start a HTTP server using the serve command on the CLI
// - perform HTTP requests against the server
// - send a os.Interrupt to the server and it closes down with no error.
//
// this test is not designed to test the specific API endpoints as there are
// other, more comprehensive tests for that, this is literally testing the points
// above.
//
// as the serve function is blocking, and ran in isolation, coverage cannot be picked up
// from this test.
func TestRoot_Serve(t *testing.T) {
	// if checkServe is true, that means we have
	// specifically ran this test in isolation with the `check_serve` flag set to true.
	// in that case we want to start the blocking process so that the parent can
	// send it a signal to shutdown.
	if *checkServe {
		out := new(bytes.Buffer)
		cmd := Root()
		cmd.SetArgs([]string{"serve", fmt.Sprintf("--port=%d", expectedPort)})
		cmd.SetOut(out)
		cmd.SetErr(out)
		assert.NoError(t, cmd.Execute())

		// Execute is blocking on the serve function
		// we can only reach this point if (and only if)
		// we have received either os.Interrupt (SIGINT), SIGINT or SIGTERM
		// from the parent process
		fmt.Print("received SIGINT")
		return
	}

	var subTimeout time.Duration
	if deadline, ok := t.Deadline(); ok {
		subTimeout = time.Until(deadline)
		subTimeout -= subTimeout / 10 // Leave 10% headroom for cleaning up subprocess.
	}

	args := []string{
		"-test.v",
		"-test.run=TestRoot_Serve$",
		"-check_serve",
	}
	if subTimeout != 0 {
		args = append(args, fmt.Sprintf("-test.timeout=%v", subTimeout))
	}

	// run this test again in isolation, setting the `check_serve` flag to true.
	cmd := exec.Command(os.Args[0], args...)
	b := &bytes.Buffer{}
	cmd.Stdout, cmd.Stderr = b, b
	require.NoError(t, cmd.Start())

	// firstly check that the server is running in the other process.
	require.NoError(t, <-waitForStatus(t, "/status", http.StatusOK))

	// send an interrupt to the spawned process
	require.NoError(t, cmd.Process.Signal(os.Interrupt))

	// wait for the process to exit
	ps, err := cmd.Process.Wait()
	assert.NoError(t, err)

	// assert that the serve function returned a zero (no error) exit code.
	assert.Equal(t, 0, ps.ExitCode())

	// assert we retrieved the print from the child process
	// that we received the SIGINT.
	assert.Contains(t, b.String(), "received SIGINT")
}

// newRequest generates a new HTTP request to point to the locally running server.
func newRequest(t *testing.T, path string) *http.Request {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u := fmt.Sprintf("http://localhost:%d%s", expectedPort, path)
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet, u,
		nil,
	)
	require.NoError(t, err)
	return req
}

// waitForStatus polls the local endpoint at the supplied path until
// the supplied status code is retrieved or the context is timed out.
// this will purposefully ignore any other error from the request itself
// except for the context timing out
//
// the polling is performed on a separate go-routine and a channel with a signal
// is returned in which an error is sent to whether we could contact the endpoint
// before the context expired.
func waitForStatus(t *testing.T, path string, statusCode int) <-chan error {
	// give our process 2 seconds to respond with the correct status
	// this should be ample.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

	sig := make(chan error, 1)
	go func() {
		var res *http.Response
		defer cancel()                // ensure we cancel the context to stop memory leaks
		defer func() { close(sig) }() // ensure the channel is closed too.

		// while we have no response or the status code doesnt match
		for res == nil || res.StatusCode != statusCode {
			// continue to perform HTTP requests against our local server
			// until we either receive a HTTP OK or a context timeout.
			select {
			case <-ctx.Done():
				sig <- ctx.Err()
				return
			default:
				req := newRequest(t, path)
				res, _ = http.DefaultClient.Do(req.WithContext(ctx))
			}
		}

		sig <- nil
	}()

	return sig
}
