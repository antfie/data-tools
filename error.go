package main

import "errors"

var (
	ErrCouldNotResolvePath                 = errors.New("could not resolve path")
	ErrPathAlreadyAdded                    = errors.New("this path has already been added")
	ErrCouldNotResolveHash                 = errors.New("could not resolve hash")
	ErrCouldNotResolveFileType             = errors.New("could not resolve file type")
	ErrNotOverwritingExistingDifferentFile = errors.New("not overwriting existing (different) file")
	ErrFilesystemIsCaseInsensitive         = errors.New("the file system is case insensitive")
)
