package types

import "os"

type Filename string

func (f Filename) Expand() string {
	return os.ExpandEnv(string(f))
}
