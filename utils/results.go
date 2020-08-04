package utils

import (
	"bytes"
	"fmt"
)

// Results specifies results of the copy.
type Results struct {
	FilesCopied     int
	DirsCopied      int
	SymLinksCreated int

	Info bytes.Buffer
}

func (c *Results) String() string {
	return fmt.Sprintf(`Results
    FilesCopied     : %d
    DirsCopied      : %d
    SymLinksCreated : %d

%s
`, c.FilesCopied, c.DirsCopied, c.SymLinksCreated, string(c.Info.Bytes()))
}
