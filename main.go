/*

PassAsk will ask for a pass through an askpass interface.

Will ask only once per PassAsker.  String() or Bytes()
will consult the askpass program upon their first invocation.
Subsequent calls will return the same value.  If no result
is available yet, String() and Bytes() will wait for a result.

Output on stderr is ignored.
*/

package passask

import (
	"io"
	"io/ioutil"
	"os/exec"
	"sync"
)

type PassAsker struct {
	askpass         string
	once            sync.Once
	ok              bool
	stdout          string
	stderr          string
	stdoutReadError error
	stderrReadError error
	cmdStartError   error
	cmdError        error
}

// Create a new PassAsker which will call askpass in a shell
// on demand.
func New(askpass string) *PassAsker {
	return &PassAsker{askpass: askpass}
}

func handlePipe(opener func() (io.ReadCloser, error), rwg *sync.WaitGroup, output *string, reterr *error) {
	pipe, err := opener()
	if err != nil {
		*reterr = err
		return
	}

	rwg.Add(1)
	go func() {
		defer rwg.Done()
		stdout, err := ioutil.ReadAll(pipe)
		if err != nil {
			*reterr = err
		}
		*output = string(stdout)
	}()
}

// Cause PassAsker to call askpass immediately.
//
// Waits for askpass to complete before returning.
func (p *PassAsker) Ask() error {
	p.once.Do(func() {
		p.ok = false
		cmd := exec.Command("/bin/sh", "-c", p.askpass)
		rwg := &sync.WaitGroup{}
		handlePipe(cmd.StdoutPipe, rwg, &p.stdout, &p.stdoutReadError)
		handlePipe(cmd.StderrPipe, rwg, &p.stderr, &p.stderrReadError)

		err := cmd.Start()
		if err != nil {
			p.cmdStartError = err
		}
		rwg.Wait()
		p.cmdError = cmd.Wait()
		if p.Error() == nil {
			p.ok = true
		}
	})
	return p.Error()
}

// Return all collected errors from running askpass
func (p *PassAsker) Errors() []error {
	var errors []error
	addError := func(e error) {
		if e != nil {
			errors = append(errors, e)
		}
	}
	addError(p.cmdStartError)
	addError(p.cmdError)
	addError(p.stdoutReadError)
	addError(p.stderrReadError)
	return errors
}

// Return the most significant error from running askpass
func (p *PassAsker) Error() error {
	errors := p.Errors()
	if len(errors) == 0 {
		return nil
	}
	return errors[0]
}

// Return stderr output from askpass
func (p *PassAsker) Stderr() string {
	p.Ask()
	return p.stderr
}

// Return stdout output from askpass as a byte slice
func (p *PassAsker) Bytes() ([]byte, error) {
	stdout, err := p.String()
	return []byte(stdout), err
}

// Return stdout output from askpass as a string
func (p *PassAsker) String() (string, error) {
	err := p.Ask()
	str := p.stdout
	return str, err
}
