// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import (
	"fmt"
	"sort"
	"strings"
)

// CILConversionPair holds a single conversion (unit or measure) in a form
// ready for code generation. The Calculation field contains a pre-synthesized
// Go expression string that references the SourceVar parameter.
type CILConversionPair struct {
	SourceName  string
	TargetName  string
	TargetVar   string
	SourceVar   string
	Calculation string
	SourceGroup string // unit group of the source unit (e.g., "DistanceUnits")
	TargetGroup string // unit group of the target unit
}

// SynthesizeConversions emits Go code for a conversion lookup table and a
// helper function. The generated code defines:
//
//	var <mapName> = map[string]map[string]func(float64) float64{ ... }
//	func <funcName>(sourceName, targetName string, value float64) float64
//
// The map is keyed by sourceName, then by targetName. The helper function
// looks up the pair and executes the conversion function, panicking when the
// pair is not found.
func SynthesizeConversions(pairs []CILConversionPair, mapName, funcName, entityName string) string {
	if len(pairs) == 0 {
		return ""
	}

	// Group by source name for nested-map emission.
	groups := make(map[string][]CILConversionPair)
	for _, p := range pairs {
		groups[p.SourceName] = append(groups[p.SourceName], p)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("var %s = map[string]map[string]func(float64) float64{\n", mapName))

	sourceNames := make([]string, 0, len(groups))
	for sn := range groups {
		sourceNames = append(sourceNames, sn)
	}
	sort.Strings(sourceNames)

	for _, sn := range sourceNames {
		sb.WriteString(fmt.Sprintf("\t%q: {\n", sn))
		targets := groups[sn]
		sort.Slice(targets, func(i, j int) bool {
			return targets[i].TargetName < targets[j].TargetName
		})
		for _, p := range targets {
			sb.WriteString(fmt.Sprintf("\t\t%q: func(%s float64) float64 { return %s },\n",
				p.TargetName, p.SourceVar, p.Calculation))
		}
		sb.WriteString("\t},\n")
	}
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("func %s(sourceName, targetName string, value float64) float64 {\n", funcName))
	sb.WriteString(fmt.Sprintf("\tif inner, ok := %s[sourceName]; ok {\n", mapName))
	sb.WriteString("\t\tif fn, ok := inner[targetName]; ok {\n")
	sb.WriteString("\t\t\treturn fn(value)\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString("\t}\n")
	sb.WriteString(fmt.Sprintf("\tpanic(\"unsupported %s conversion: \" + sourceName + \" -> \" + targetName)\n", entityName))
	sb.WriteString("}\n")

	return sb.String()
}
