//go:build !unix

package mobile

func stoppedBySignal(err error) bool { return false }
