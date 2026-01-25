import { PopoverTheme, PopoverThemeName, POPOVER_THEMES } from "./types";

export function hexToRgb(hex: string): string {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  if (!result) return "0, 0, 0";
  return `${parseInt(result[1], 16)}, ${parseInt(result[2], 16)}, ${parseInt(result[3], 16)}`;
}

export function getTheme(theme: PopoverThemeName | PopoverTheme): PopoverTheme {
  if (typeof theme === "string") {
    return POPOVER_THEMES[theme];
  }
  return theme;
}

export function getThemeStyles(theme: PopoverThemeName | PopoverTheme): React.CSSProperties {
  const resolvedTheme = getTheme(theme);
  return {
    "--popover-accent": resolvedTheme.accent,
    "--popover-accent-rgb": hexToRgb(resolvedTheme.accent),
  } as React.CSSProperties;
}
