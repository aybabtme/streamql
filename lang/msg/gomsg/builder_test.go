package gomsg

import (
	"testing"

	"github.com/aybabtme/streamql/lang/msg/msgtest"
)

func TestBuilder(t *testing.T) {
	msgtest.VerifyBuilder(t, Build)
}
