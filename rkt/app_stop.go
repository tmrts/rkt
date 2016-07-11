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
	cmdAppStop = &cobra.Command{
		Use: "stop [--app=APPNAME] UUID",
		Run: ensureSuperuser(runWrapper(runStop)),
	}
)

func init() {
	cmdApp.AddCommand(cmdAppStop)
	cmdAppStop.Flags().StringVar(&flagAppName, "app", "", "name of the app")
}

func runStop(cmd *cobra.Command, args []string) (exit int) {
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

	if err = stage0.Nsenter(p.path(), podPID, stage1RootFS, []string{"/bin/systemctl", "stop", "reaper-" + (*appName).String()}); err != nil {
		stderr.PrintE("nsenter entrypoint failed", err)
		return 1
	}
	if err = stage0.Nsenter(p.path(), podPID, stage1RootFS, []string{"/bin/systemctl", "stop", (*appName).String()}); err != nil {
		stderr.PrintE("nsenter entrypoint failed", err)
		return 1
	}

	return 0
}
