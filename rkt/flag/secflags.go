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

package flag

import "strings"

const (
	insecureNone  = 0
	insecureImage = 1 << (iota - 1)
	insecureTLS
	insecureOnDisk
	insecureHTTP
	insecurePubKey

	insecureAll = (insecureImage | insecureTLS | insecureOnDisk | insecureHTTP | insecurePubKey)
)

var (
	insecureOptions = []string{"none", "image", "tls", "ondisk", "http", "pubkey", "all"}

	insecureOptionsMap = map[string]int{
		insecureOptions[0]: insecureNone,
		insecureOptions[1]: insecureImage,
		insecureOptions[2]: insecureTLS,
		insecureOptions[3]: insecureOnDisk,
		insecureOptions[4]: insecureHTTP,
		insecureOptions[5]: insecurePubKey,
		insecureOptions[6]: insecureAll,
	}
)

type SecFlags struct {
	*bitFlags
}

func NewSecFlags(defOpts string) (*SecFlags, error) {
	bf, err := newBitFlags(insecureOptions, defOpts, insecureOptionsMap)
	if err != nil {
		return nil, err
	}

	sf := &SecFlags{
		bitFlags: bf,
	}
	return sf, nil
}

func (sf *SecFlags) SkipImageCheck() bool {
	return sf.hasFlag(insecureImage)
}

func (sf *SecFlags) SkipTLSCheck() bool {
	return sf.hasFlag(insecureTLS)
}

func (sf *SecFlags) SkipOnDiskCheck() bool {
	return sf.hasFlag(insecureOnDisk)
}

func (sf *SecFlags) AllowHTTP() bool {
	return sf.hasFlag(insecureHTTP)
}

func (sf *SecFlags) ConsiderInsecurePubKeys() bool {
	return sf.hasFlag(insecurePubKey)
}

func (sf *SecFlags) SkipAllSecurityChecks() bool {
	return sf.hasFlag(insecureAll)
}

func (sf *SecFlags) SkipAnySecurityChecks() bool {
	return sf.flags != 0
}

func (sf *SecFlags) String() string {
	opts := []string{}

	for optstr, opt := range insecureOptionsMap {
		if sf.hasFlag(opt) {
			opts = append(opts, optstr)
		}
	}

	return strings.Join(opts, ",")
}
