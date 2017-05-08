package astvm

import (
	"testing"

	"github.com/aybabtme/streamql/lang/msg/gomsg"
	"github.com/aybabtme/streamql/lang/vm/vmtest"
)

func TestInterpreter(t *testing.T) {
	vmtest.Verify(t, Interpreter)
}

func BenchmarkInterpreter(b *testing.B) {
	vmtest.Bench(b, gomsg.Build, Interpreter)
}
