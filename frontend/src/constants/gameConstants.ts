/**
 * Game VP values - centralized constants for victory point calculations.
 */
export const VP_VALUES = {
  MILESTONE: 5,
  AWARD_FIRST: 5,
  AWARD_SECOND: 2,
} as const;

/**
 * Animation timing constants for end-game UI sequences.
 * All values in milliseconds.
 */
export const ANIMATION_TIMINGS = {
  /** Delay between revealing individual items (milestones, awards, cards) */
  ITEM_REVEAL: 400,
  /** Delay between major section transitions */
  SECTION_TRANSITION: 500,
  /** Duration for bar chart fill animation */
  BAR_CHART_FILL: 2000,
  /** Delay between podium placement reveals */
  PODIUM_REVEAL: 800,
  /** Delay for tile VP reveals */
  TILE_REVEAL: 400,
  /** Phase transition timings */
  PHASE_INTRO: 1500,
  PHASE_TR: 2000,
  PHASE_MILESTONES: 2000,
  PHASE_AWARDS: 2000,
  PHASE_SUMMARY: 3000,
  PHASE_RANKINGS: 4000,
  /** Card VP display duration */
  CARD_VP_DISPLAY: 3000,
  /** Brief pause between animations */
  BRIEF_PAUSE: 300,
  /** Cleanup delay between phases */
  PHASE_CLEANUP: 500,
  /** Delay after tiles phase before cards */
  POST_TILES_DELAY: 1000,
} as const;

/**
 * VP Sequence phase order for determining visibility.
 */
export const VP_PHASE_ORDER = [
  "intro",
  "tr",
  "milestones",
  "awards",
  "tiles",
  "cards",
  "summary",
  "rankings",
  "complete",
] as const;

export type VPSequencePhase = (typeof VP_PHASE_ORDER)[number];

/**
 * Check if the current phase is at or after a target phase.
 * Useful for progressive VP reveal logic.
 */
export function isPhaseAtOrAfter(
  current: VPSequencePhase,
  target: VPSequencePhase,
): boolean {
  return VP_PHASE_ORDER.indexOf(current) >= VP_PHASE_ORDER.indexOf(target);
}

/**
 * Tile highlight types for 3D board visualization.
 */
export type TileHighlightType = "greenery" | "city" | "adjacent" | null;
