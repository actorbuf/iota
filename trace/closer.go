package trace

type NullCloser struct{}

func (*NullCloser) Close() error { return nil }
