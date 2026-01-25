import { ModalTheme, ModalThemeName, MODAL_THEMES } from "./types";

export function hexToRgb(hex: string): string {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  if (!result) return "0, 0, 0";
  return `${parseInt(result[1], 16)}, ${parseInt(result[2], 16)}, ${parseInt(result[3], 16)}`;
}

export function getTheme(theme: ModalThemeName | ModalTheme): ModalTheme {
  if (typeof theme === "string") {
    return MODAL_THEMES[theme];
  }
  return theme;
}

export function getThemeStyles(theme: ModalThemeName | ModalTheme): React.CSSProperties {
  const resolvedTheme = getTheme(theme);
  return {
    "--modal-accent": resolvedTheme.accent,
    "--modal-accent-rgb": hexToRgb(resolvedTheme.accent),
  } as React.CSSProperties;
}
