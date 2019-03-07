// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package benchgit

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Do runs benchmarks on the supplied git refs and returns its
// data as an io.Reader.
func Do(benchArgs string, ref string) (io.Reader, error) {
	origRef, err := FindRef()
	if err != nil {
		return nil, err
	}

	defer func() {
		if ref == origRef {
			return
		}

		_, err := runCommand("git", "checkout", origRef)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not revert back to original ref %s: %s", origRef, err)
		}
	}()

	cmd := append([]string{"test"}, strings.Split(benchArgs, " ")...)

	if ref != origRef {
		_, err := runCommand("git", "checkout", ref)
		if err != nil {
			return nil, err
		}
	}

	fmt.Fprintf(os.Stderr, "running in %s: go %s\n", ref, strings.Join(cmd, " "))
	result, err := runCommand("go", cmd...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FindRef finds the current ref.
func FindRef() (string, error) {
	result, err := runCommandString("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}

	if result != "HEAD" {
		return strings.TrimSpace(result), nil
	}

	tags, err := runCommandString("git", "tags", "--points-at", "HEAD")
	if err != nil {
		return "", err
	}

	if len(tags) == 0 {
		result, err := runCommandString("git", "rev-parse", "--abbrev-ref", "HEAD")
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(result), nil
	}

	return strings.TrimSpace(strings.Split(tags, "\n")[0]), nil
}

func runCommand(name string, arg ...string) (*bytes.Buffer, error) {
	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	cmd := exec.Command(name, arg...)
	cmd.Stdout = outBuf
	cmd.Stderr = errBuf

	err := cmd.Run()
	if err != nil {
		os.Stderr.Write(errBuf.Bytes())
	}

	return outBuf, err
}

func runCommandString(name string, arg ...string) (string, error) {
	out, err := runCommand(name, arg...)
	return out.String(), err
}
