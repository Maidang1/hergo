// Palette and selection formulas adapted from loctvl842/monokai-pro.nvim.
// Copyright (c) 2023 Loc. SPDX-License-Identifier: MIT
// See monokai-pro-LICENSE and herdr-ATTRIBUTION.md.
package design

import (
	"fmt"
	"math"

	"github.com/egoist/mygo/ui"
)

// Keep the names and values from lua/monokai-pro/palette/pro.lua together.
var monokaiProPalette = map[string]string{
	"dark2":      "#19181a",
	"dark1":      "#221f22",
	"background": "#2d2a2e",
	"text":       "#fcfcfa",
	"accent1":    "#ff6188",
	"accent2":    "#fc9867",
	"accent3":    "#ffd866",
	"accent4":    "#a9dc76",
	"accent5":    "#78dce8",
	"accent6":    "#ab9df2",
	"dimmed1":    "#c1c0c0",
	"dimmed2":    "#939293",
	"dimmed3":    "#727072",
	"dimmed4":    "#5b595c",
	"dimmed5":    "#403e41",
}

// Keep the names and values from lua/monokai-pro/palette/light.lua together.
var monokaiProLightPalette = map[string]string{
	"dark2":      "#d3cdcc",
	"dark1":      "#ede7e5",
	"background": "#faf4f2",
	"text":       "#29242a",
	"accent1":    "#e14775",
	"accent2":    "#e16032",
	"accent3":    "#cc7a0a",
	"accent4":    "#269d69",
	"accent5":    "#1c8ca8",
	"accent6":    "#7058be",
	"dimmed1":    "#706b6e",
	"dimmed2":    "#918c8e",
	"dimmed3":    "#a59fa0",
	"dimmed4":    "#bfb9ba",
	"dimmed5":    "#d3cdcc",
}

func monokaiProTokens(p map[string]string) map[string]string {
	// Match theme/scheme.lua and colors/convert.lua, including RGB rounding.
	blend := func(foreground string, alpha float64, background string) string {
		fg, bg := ui.Hex(p[foreground]), ui.Hex(p[background])
		mix := func(f, b uint8) uint8 {
			return uint8(math.Floor((1-alpha)*float64(b) + alpha*float64(f) + .5))
		}
		return fmt.Sprintf("#%02x%02x%02x", mix(fg.R, bg.R), mix(fg.G, bg.G), mix(fg.B, bg.B))
	}
	return map[string]string{
		"accent": p["accent3"], "panel_bg": p["background"], "sidebar_bg": p["dark1"],
		"active_row_bg": blend("text", .11, "dark1"),
		"selection_bg":  blend("dimmed1", .15, "background"),
		"surface0":      p["dimmed5"], "surface1": p["dimmed4"], "surface_dim": p["background"],
		"overlay0": p["dimmed2"], "overlay1": p["dimmed3"],
		"text": p["text"], "subtext0": p["dimmed1"],
		"mauve": p["accent6"], "green": p["accent4"], "yellow": p["accent3"],
		"red": p["accent1"], "blue": p["accent2"], "teal": p["accent5"], "peach": p["accent2"],
	}
}

// Match set_terminal_colors in lua/monokai-pro/theme/init.lua.
func monokaiProANSI(p map[string]string) *[16]ui.Color {
	keys := [16]string{
		"dark1", "accent1", "accent4", "accent3", "accent2", "accent6", "accent5", "text",
		"dimmed3", "accent1", "accent4", "accent3", "accent2", "accent6", "accent5", "text",
	}
	var colors [16]ui.Color
	for i, key := range keys {
		colors[i] = ui.Hex(p[key])
	}
	return &colors
}
