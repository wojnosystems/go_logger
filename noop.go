package go_logger

// NoOp is a logger that does nothing
// This is useful if you want to globally disable logging
type NoOp struct {
}

// Does nothing
func (n *NoOp) Log(tag string, msg Msg, skip int) {
	// no-op
}

// Does nothing
// @return nil
func (n *NoOp) ReOpen() error {
	// no-op
	return nil
}

// Does nothing
// @return nil
func (n *NoOp) Close() error {
	// no-op
	return nil
}
