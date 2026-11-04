# Herdr palette attribution

The original palettes in `herdr_palettes.json` were mechanically extracted from `src/app/state.rs` in
[herdrdev/herdr](https://github.com/herdrdev/herdr) at commit
`ee5bfb5041bb3e4a49d591c72308aae3c81ecdde`. The configuration resolution and
color parser in `herdr.go` are adapted from that revision's
`src/config/theme.rs` and `src/app/mod.rs`.

Upstream Herdr is licensed under Apache-2.0; the full license is retained in
`herdr-LICENSE`. This adaptation converts Ratatui colors into native UI colors,
keeps separate resolved light/dark slots, and maps the original surface and
status tokens to this application's semantic palette. Upstream terminal reset
colors resolve to native defaults because this UI does not inherit a host
terminal palette; named ANSI colors use the conventional 16-color RGB values.

To reproduce the palette extraction, read each of the 18 constructors named by
`THEME_NAMES` in `src/config/theme.rs`, extract their `Color::Rgb(r, g, b)` and
named `Color` field values, convert RGB tuples to lowercase hex strings, and
write the field names unchanged. Those upstream palette RGB values are not
hand-maintained.

## Local palettes

`monokai-pro` and `monokai-pro-light` are local additions, not part of the upstream Herdr extraction.
`monokai_pro.go` copies the original variable names and values directly from
[`lua/monokai-pro/palette/pro.lua`](https://github.com/loctvl842/monokai-pro.nvim/blob/a68e38b8e55d69a215d0f02598900a79c356da9d/lua/monokai-pro/palette/pro.lua)
and [`lua/monokai-pro/palette/light.lua`](https://github.com/loctvl842/monokai-pro.nvim/blob/a68e38b8e55d69a215d0f02598900a79c356da9d/lua/monokai-pro/palette/light.lua)
in `loctvl842/monokai-pro.nvim` at commit
`a68e38b8e55d69a215d0f02598900a79c356da9d`.
The MIT license is retained in `monokai-pro-LICENSE`.

The surface mapping follows `lua/monokai-pro/theme/scheme.lua`: the shell and
content use `background`, the sidebar uses `dark1`, raised surfaces use
`dimmed5`, and hovered controls use `dimmed4`. The action accent follows the
active tab's `accent3`. Editor selection blends `dimmed1` at 15% over
`background`; the selected sidebar row blends `text` at 11% over `dark1`.
RGB channels round as in upstream `colors/convert.lua`. The terminal's 16 ANSI
colors follow `set_terminal_colors` in `lua/monokai-pro/theme/init.lua`,
including orange `accent2` for ANSI blue and `dimmed3` for bright black.
