// Copyright 2014 The rkt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//+build linux

package stage0

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/hashicorp/errwrap"
)

func Nsenter(cdir string, podPID int, stage1Path string, cmdline []string) error {
	if err := os.Chdir(cdir); err != nil {
		return errwrap.Wrap(errors.New("error changing to dir"), err)
	}

	argv := []string{"/run/current-system/sw/bin/nsenter", "-m", "-p", "-t", fmt.Sprintf("%d", podPID)}
	argv = append(argv, cmdline...)
	if err := syscall.Exec(argv[0], argv, os.Environ()); err != nil {
		return errwrap.Wrap(errors.New("error execing nsenter"), err)
	}

	// never reached
	return nil
}
