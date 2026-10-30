// SPDX-License-Identifier: GPL-3.0
package branding

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"equilotl/internal/release"
)

//go:embed runtime.cjs
var Runtime []byte

//go:embed loader.cjs
var Loader []byte

//go:embed trust.json
var TrustData []byte

//go:embed legacy-trust.json
var LegacyTrustData []byte

func ReleaseTrust() release.Trust {
	var t release.Trust
	if e := json.Unmarshal(TrustData, &t); e != nil {
		panic(e)
	}
	return t
}

func init() {
	Runtime = bytes.ReplaceAll(Runtime, []byte("\r\n"), []byte("\n"))
	Loader = bytes.ReplaceAll(Loader, []byte("\r\n"), []byte("\n"))
	TrustData = bytes.ReplaceAll(TrustData, []byte("\r\n"), []byte("\n"))
	LegacyTrustData = bytes.ReplaceAll(LegacyTrustData, []byte("\r\n"), []byte("\n"))
}
