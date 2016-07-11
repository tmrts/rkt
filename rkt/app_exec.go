// Copyright 2016 The rkt Authors
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

package main

import (
	"fmt"

	"github.com/coreos/rkt/stage0"
	"github.com/coreos/rkt/store"
	"github.com/spf13/cobra"
)

var (
	cmdAppExec = &cobra.Command{
		Use: "exec [--app=APPNAME] UUID [CMD [ARGS ...]]",
		Run: ensureSuperuser(runWrapper(runExec)),
	}
)

func init() {
	cmdApp.AddCommand(cmdAppExec)
	cmdAppExec.Flags().StringVar(&flagAppName, "app", "", "name of the app to enter within the specified pod")

	// Disable interspersed flags to stop parsing after the first non flag
	// argument. This is need to permit to correctly handle
	// multiple "IMAGE -- imageargs ---"  options
	cmdAppExec.Flags().SetInterspersed(false)
}

func runExec(cmd *cobra.Command, args []string) (exit int) {
	if len(args) < 1 {
		cmd.Usage()
		return 1
	}

	p, err := getPodFromUUIDString(args[0])
	if err != nil {
		stderr.PrintE("problem retrieving pod", err)
		return 1
	}
	defer p.Close()

	if !p.isRunning() {
		stderr.Printf("pod %q isn't currently running", p.uuid)
		return 1
	}

	podPID, err := p.getContainerPID1()
	if err != nil {
		stderr.PrintE(fmt.Sprintf("unable to determine the pid for pod %q", p.uuid), err)
		return 1
	}

	appName, err := getAppName(p)
	if err != nil {
		stderr.PrintE("unable to determine app name", err)
		return 1
	}

	argv, err := getExecArgv(p, args)
	if err != nil {
		stderr.PrintE("Exec failed", err)
		return 1
	}

	s, err := store.NewStore(getDataDir())
	if err != nil {
		stderr.PrintE("cannot open store", err)
		return 1
	}

	stage1TreeStoreID, err := p.getStage1TreeStoreID()
	if err != nil {
		stderr.PrintE("error getting stage1 treeStoreID", err)
		return 1
	}

	stage1RootFS := s.GetTreeStoreRootFS(stage1TreeStoreID)

	if err = stage0.Enter(p.path(), podPID, *appName, stage1RootFS, argv); err != nil {
		stderr.PrintE("enter entrypoint failed", err)
		return 1
	}

	return 0
}

func getExecArgv(p *pod, cmdArgs []string) ([]string, error) {
	var argv []string
	if len(cmdArgs) < 2 {
		return nil, fmt.Errorf("no command specified")
	} else {
		argv = cmdArgs[1:]
	}

	return argv, nil
}
