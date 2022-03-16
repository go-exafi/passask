package passask

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"

	"github.com/go-exafi/shq"
)

func TestAsk_delay(t *testing.T) {
	asker := New("sleep 1;printf hi")
	results := make([]string, 10)
	wg := sync.WaitGroup{}
	for i := 0; i < len(results); i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			results[i], err = asker.String()
			if err != nil {
				t.Errorf("Failure reading in asker %d: %v", i, err)
			}
		}()
	}
	asker.Ask()
	wg.Wait()
	for i := 0; i < len(results); i++ {
		if results[i] != "hi" {
			t.Errorf("asker[%d] did not return expected value 'hi': got '%s'", i, results[i])
		} else {
			t.Logf("asker[%d] was successful", i)
		}
	}
}

func TestAsk_random(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Run("randomized", func(t *testing.T) {
			buf := make([]byte, 48)
			_, err := rand.Read(buf)
			if err != nil {
				t.Errorf("Failed while reading random bytes: %v", err)
				return
			}
			arg := shq.Arg(buf)
			asker := New(fmt.Sprintf("printf %%s %s", arg))
			response, err := asker.String()
			if err != nil {
				t.Errorf("error reading from test PassAsker: %v", err)
				return
			}
			expected := arg.Unescaped()
			if response != expected {
				t.Errorf("expected result (%#v) didn't match actual result (%#v)", expected, response)
				return
			}
		})
	}
}
func Example() {
	// get SUDO_ASKPASS
	askpass := os.Getenv("SUDO_ASKPASS")
	// but this is an example... let's do something else, actually
	askpass = "printf hello"

	// build the asker which will run askpass
	asker := New(askpass)

	// actually consult askpass and return a string and/or error
	response, err := asker.String()
	if err != nil {
		fmt.Printf("Failed to read a password\n%s", asker.Stderr())
		return
	}

	fmt.Printf("Password: %s\n", response)
	// Output:
	// Password: hello
}

func Example_failure() {
	// get SUDO_ASKPASS
	askpass := os.Getenv("SUDO_ASKPASS")
	// but this is an example... let's do something else, actually
	askpass = "echo errar 1>&2;exit 1"

	// build the asker which will run askpass
	asker := New(askpass)

	// actually consult askpass and return a string and/or error
	response, err := asker.String()
	if err != nil {
		fmt.Printf("Failed to read a password\n%s", asker.Stderr())
		return
	}

	fmt.Printf("Password: %s\n", response)
	// Output:
	// Failed to read a password
	// errar
}
