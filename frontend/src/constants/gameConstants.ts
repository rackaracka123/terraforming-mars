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
} as const;
