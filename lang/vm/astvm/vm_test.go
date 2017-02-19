package astvm

import (
	"testing"

	"github.com/aybabtme/streamql/lang/vm/vmtest"
)

func TestInterpreter(t *testing.T) {
	vmtest.Verify(t, Interpreter)
}
