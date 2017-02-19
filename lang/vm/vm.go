package vm

import "github.com/aybabtme/streamql/lang/msg"

// A VM runs a query on a source of message and puts the result on a sink.
// A builder of message is used when the VM needs to construct new messages
// as part of the query.
type VM interface {
	Run(build msg.Builder, src msg.Source, sink msg.Sink) error
}

// Options that can be used by a VM
type Options struct {
	// Strict execution means that errors encountered during the
	// query's execution will stop the execution on further messages.
	// By default, a VM will emit an error when a query turns out
	// to be invalid when running against a message. In normal mode,
	// the VM will simply skip that message and process the next one.
	// In strict mode, the VM will stop and return an error.
	Strict bool
}
