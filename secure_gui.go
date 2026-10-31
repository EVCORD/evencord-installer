//go:build !cli

// SPDX-License-Identifier: GPL-3.0
package main

import (
	"encoding/hex"
	"equilotl/branding"
	"github.com/AllenDang/imgui-go"
)

func applyBrandTheme() {
	c := branding.Project.Colors
	rgba := func(value string) imgui.Vec4 {
		bytes, err := hex.DecodeString(value[1:])
		if err != nil || len(bytes) != 3 {
			panic("invalid branding palette")
		}
		return imgui.Vec4{X: float32(bytes[0]) / 255, Y: float32(bytes[1]) / 255, Z: float32(bytes[2]) / 255, W: 1}
	}
	style := imgui.CurrentStyle()
	for id, value := range map[imgui.StyleColorID]string{
		imgui.StyleColorWindowBg: c.Dark, imgui.StyleColorChildBg: c.Dark,
		imgui.StyleColorPopupBg: c.Surface, imgui.StyleColorFrameBg: c.Surface,
		imgui.StyleColorText: c.Text, imgui.StyleColorButton: c.Primary,
		imgui.StyleColorButtonHovered: c.Surface, imgui.StyleColorButtonActive: c.Primary,
		imgui.StyleColorHeader: c.Primary, imgui.StyleColorHeaderHovered: c.Surface,
		imgui.StyleColorCheckMark: c.Text, imgui.StyleColorBorder: c.Surface,
		imgui.StyleColorSeparator: c.Surface, imgui.StyleColorTitleBgActive: c.Surface,
	} {
		style.SetColor(id, rgba(value))
	}
}
