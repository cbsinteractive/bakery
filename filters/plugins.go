package filters

import "github.com/zencoder/go-dash/mpd"

type execPluginDASH func(manifest *mpd.MPD)

var (
	pluginDASH = map[string]execPluginDASH{
		"dvsRoleOverride": dvsRoleOverride,
	}
)

func dvsRoleOverride(manifest *mpd.MPD) {
	for _, period := range manifest.Periods {
		for _, as := range period.AdaptationSets {
			for i, access := range as.AccessibilityElems {
				if access != nil && *access.SchemeIdUri == "urn:tva:metadata:cs:AudioPurposeCS:2007" {
					as.Roles[i].Value = strptr("description")
				}
			}
		}
	}
}
