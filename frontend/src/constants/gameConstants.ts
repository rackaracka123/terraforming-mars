/**
 * Game VP values - centralized constants for victory point calculations.
 */
export const VP_VALUES = {
  MILESTONE: 5,
  AWARD_FIRST: 5,
  AWARD_SECOND: 2,
} as const;

/**
 * Speed multipliers for fine-grained animation control.
 * 1.0 = normal speed, 1.5 = 50% slower, 2.0 = twice as slow
 */
export const ANIMATION_SPEED = {
  /** Multiplier for tile reveals (greenery/city counting) */
  TILE: 1.0,
  /** Multiplier for phase durations (milestones, awards, summary, etc.) */
  PHASE: 1.5,
  /** Multiplier for pauses between items */
  PAUSE: 1.5,
  /** Multiplier for transition effects */
  TRANSITION: 1.5,
} as const;

/**
 * Animation timing constants for end-game UI sequences.
 * All values in milliseconds. Adjusted by ANIMATION_SPEED multipliers.
 */
export const ANIMATION_TIMINGS = {
  /** Delay between revealing individual items (milestones, awards, cards) */
  ITEM_REVEAL: 400 * ANIMATION_SPEED.TRANSITION,
  /** Delay between major section transitions */
  SECTION_TRANSITION: 500 * ANIMATION_SPEED.TRANSITION,
  /** Duration for bar chart fill animation */
  BAR_CHART_FILL: 2000 * ANIMATION_SPEED.TRANSITION,
  /** Delay between podium placement reveals */
  PODIUM_REVEAL: 800 * ANIMATION_SPEED.TRANSITION,
  /** Delay for tile VP reveals */
  TILE_REVEAL: 400 * ANIMATION_SPEED.TILE,
  /** Phase transition timings */
  PHASE_INTRO: 1500 * ANIMATION_SPEED.PHASE,
  PHASE_TR: 2000 * ANIMATION_SPEED.PHASE,
  PHASE_MILESTONES: 2000 * ANIMATION_SPEED.PHASE,
  PHASE_AWARDS: 2000 * ANIMATION_SPEED.PHASE,
  PHASE_SUMMARY: 3000 * ANIMATION_SPEED.PHASE,
  PHASE_RANKINGS: 4000 * ANIMATION_SPEED.PHASE,
  /** Card VP display duration */
  CARD_VP_DISPLAY: 3000 * ANIMATION_SPEED.PHASE,
  /** Brief pause between tile animations (uses TILE speed) */
  BRIEF_PAUSE: 300 * ANIMATION_SPEED.TILE,
  /** Cleanup delay between phases */
  PHASE_CLEANUP: 500 * ANIMATION_SPEED.PAUSE,
  /** Delay after tiles phase before cards */
  POST_TILES_DELAY: 1000 * ANIMATION_SPEED.PAUSE,
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
export function isPhaseAtOrAfter(current: VPSequencePhase, target: VPSequencePhase): boolean {
  return VP_PHASE_ORDER.indexOf(current) >= VP_PHASE_ORDER.indexOf(target);
}

/**
 * Tile highlight types for 3D board visualization.
 */
export type TileHighlightType = "greenery" | "city" | "adjacent" | null;
