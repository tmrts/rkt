// Copyright 2015 The rkt Authors
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
	"bytes"
	"fmt"
	"strings"

	"github.com/appc/spec/schema"
	common "github.com/coreos/rkt/common"
	"github.com/dustin/go-humanize"
	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
)

var (
	cmdAppList = &cobra.Command{
		Use: "list",
		Run: runWrapper(runAppList),
	}
)

func init() {
	cmdApp.AddCommand(cmdAppList)
}

func runAppList(cmd *cobra.Command, args []string) int {
	var errors []error
	tabBuffer := new(bytes.Buffer)
	tabOut := getTabOutWithWriter(tabBuffer)

	if !flagNoLegend {
		if flagFullOutput {
			fmt.Fprintf(tabOut, "APP\tIMAGE NAME\tIMAGE ID\tSTATE\tCREATED\tSTARTED\tNETWORKS\n")
		} else {
			fmt.Fprintf(tabOut, "APP\tIMAGE NAME\tSTATE\tCREATED\tSTARTED\tNETWORKS\n")
		}
	}

	if err := walkPods(includeMostDirs, func(p *pod) {
		uuid := p.uuid.String()
		if !strings.HasPrefix(uuid, args[0]) {
			return
		}

		pm := schema.PodManifest{}

		if !p.isPreparing && !p.isAbortedPrepare && !p.isExitedDeleting {
			// TODO(vc): we should really hold a shared lock here to prevent gc of the pod
			pmf, err := p.readFile(common.PodManifestPath(""))
			if err != nil {
				errors = append(errors, newPodListReadError(p, err))
				return
			}

			if err := pm.UnmarshalJSON(pmf); err != nil {
				errors = append(errors, newPodListLoadError(p, err, pmf))
				return
			}

			if len(pm.Apps) == 0 {
				errors = append(errors, newPodListZeroAppsError(p))
				return
			}
		}

		type printedApp struct {
			appName string
			imgName string
			imgID   string
			state   string
			nets    string
			created string
			started string
		}

		var appsToPrint []printedApp
		state := p.getState()
		nets := fmtNets(p.nets)

		created, err := p.getCreationTime()
		if err != nil {
			errors = append(errors, errwrap.Wrap(fmt.Errorf("unable to get creation time for pod %q", uuid), err))
		}
		var createdStr string
		if flagFullOutput {
			createdStr = created.Format(defaultTimeLayout)
		} else {
			createdStr = humanize.Time(created)
		}

		started, err := p.getStartTime()
		if err != nil {
			errors = append(errors, errwrap.Wrap(fmt.Errorf("unable to get start time for pod %q", uuid), err))
		}
		var startedStr string
		if !started.IsZero() {
			if flagFullOutput {
				startedStr = started.Format(defaultTimeLayout)
			} else {
				startedStr = humanize.Time(started)
			}
		}

		for _, app := range pm.Apps {
			imageName, err := getImageName(p, app.Name)
			if err != nil {
				errors = append(errors, newPodListLoadImageManifestError(p, err))
				imageName = "--"
			}

			var imageID string
			if flagFullOutput {
				imageID = app.Image.ID.String()[:19]
			}

			appsToPrint = append(appsToPrint, printedApp{
				appName: app.Name.String(),
				imgName: imageName,
				imgID:   imageID,
				state:   state,
				nets:    nets,
				created: createdStr,
				started: startedStr,
			})
			// clear those variables so they won't be
			// printed for another apps in the pod as they
			// are actually describing a pod, not an app
			state = ""
			nets = ""
			createdStr = ""
			startedStr = ""
		}
		// if we reached that point, then it means that the
		// pod and all its apps are valid, so they can be
		// printed
		for _, app := range appsToPrint {
			if flagFullOutput {
				fmt.Fprintf(tabOut, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", app.appName, app.imgName, app.imgID, app.state, app.created, app.started, app.nets)
			} else {
				fmt.Fprintf(tabOut, "%s\t%s\t%s\t%s\t%s\t%s\n", app.appName, app.imgName, app.state, app.created, app.started, app.nets)
			}
		}

	}); err != nil {
		stderr.PrintE("failed to get pod handles", err)
		return 1
	}

	if len(errors) > 0 {
		sep := "----------------------------------------"
		stderr.Printf("%d error(s) encountered when listing pods:", len(errors))
		stderr.Print(sep)
		for _, err := range errors {
			stderr.Error(err)
			stderr.Print(sep)
		}
		stderr.Print("misc:")
		stderr.Printf("  rkt's appc version: %s", schema.AppContainerVersion)
		stderr.Print(sep)
		// make a visible break between errors and the listing
		stderr.Print("")
	}

	tabOut.Flush()
	stdout.Print(tabBuffer)
	return 0
}
