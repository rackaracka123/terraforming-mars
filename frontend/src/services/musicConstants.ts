/**
 * Music system constants
 * Centralized configuration for background music behavior
 */

// Routes where background music should play
export const LANDING_PAGE_ROUTES = ["/", "/create", "/join"];

// Audio configuration
export const MUSIC_CONFIG = {
  PATH: "/assets/audio/menu-music.mp3",
  DEFAULT_VOLUME: 0.3,
  FADE_DURATION_MS: 2000,
  FADE_STEPS: 50,
} as const;

// Helper function to check if a path is a landing page
export const isLandingPage = (path: string): boolean =>
  LANDING_PAGE_ROUTES.includes(path);
