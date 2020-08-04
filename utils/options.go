package utils

import "os"

// FileHandlerFunc type to define a function for carrying out a file processing for copying or processing as a template
type FileHandlerFunc func(src, dest string, info os.FileInfo, opt Options, results *Results) (err error)

// Options specifies optional actions on copying.
type Options struct {
	// OnSymlink can specify what to do on symlink
	OnSymlink func(src string) SymlinkAction
	// Skip can specify which files should be skipped
	Skip func(src string) (bool, error)
	// AddPermission to every entities,
	// NO MORE THAN 0777
	AddPermission os.FileMode
	// Sync file after copy.
	// Useful in case when file must be on the disk
	// (in case crash happens, for example),
	// at the expense of some performance penalty
	Sync bool

	// FileHandler - returns a handler for file - if nill will use the default copy handler
	FileHandler func(src, dest string, info os.FileInfo) FileHandlerFunc

	// ShouldCopy - return bool if dir or file should be copied
	ShouldCopy func(opt os.FileInfo) bool
}

// SymlinkAction represents what to do on symlink.
type SymlinkAction int

const (
	// Deep creates hard-copy of contents.
	Deep SymlinkAction = iota
	// Shallow creates new symlink to the dest of symlink.
	Shallow
	// Skip does nothing with symlink.
	Skip
)

// DefaultCopyOptions provides default options,
// which would be modified by usage-side.
func DefaultCopyOptions() Options {
	return Options{
		OnSymlink: func(string) SymlinkAction {
			return Shallow // Do shallow copy
		},
		Skip: func(string) (bool, error) {
			return false, nil // Don't skip
		},
		AddPermission: 0,     // Add nothing
		Sync:          false, // Do not sync
	}
}
