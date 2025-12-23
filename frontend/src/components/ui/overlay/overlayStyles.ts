/**
 * Shared style constants for card selection overlays
 * Extracted to eliminate duplication across ProductionCardSelection, StartingCardSelection,
 * CardDrawSelection, and PendingCardSelection overlays
 */

export const OVERLAY_CONTAINER_CLASS =
  "relative z-[1] w-[90%] max-w-[1400px] max-h-[90vh] flex flex-col bg-space-black-darker/95 border-2 border-space-blue-400 rounded-[20px] overflow-hidden backdrop-blur-space shadow-[0_20px_60px_rgba(0,0,0,0.6),0_0_60px_rgba(30,60,150,0.5)] max-[768px]:w-full max-[768px]:h-screen max-[768px]:max-h-screen max-[768px]:rounded-none";

export const OVERLAY_BACKGROUND_CLASS =
  "absolute inset-0 bg-black/60 backdrop-blur-sm";

export const OVERLAY_HEADER_CLASS =
  "py-6 px-8 bg-black/40 border-b border-space-blue-600 max-[768px]:p-5";

export const OVERLAY_TITLE_CLASS =
  "m-0 font-orbitron text-[28px] font-bold text-white text-shadow-glow tracking-wider max-[768px]:text-2xl";

export const OVERLAY_DESCRIPTION_CLASS =
  "mt-2 mb-0 text-base text-white/80 max-[768px]:text-sm";

export const OVERLAY_CARDS_CONTAINER_CLASS =
  "flex-1 overflow-x-auto overflow-y-hidden p-8 flex items-center bg-[radial-gradient(ellipse_at_center,rgba(139,69,19,0.1)_0%,transparent_70%)] [&::-webkit-scrollbar]:h-2 [&::-webkit-scrollbar-track]:bg-white/5 [&::-webkit-scrollbar-track]:rounded [&::-webkit-scrollbar-thumb]:bg-white/20 [&::-webkit-scrollbar-thumb]:rounded [&::-webkit-scrollbar-thumb:hover]:bg-white/30 max-[768px]:p-5";

export const OVERLAY_CARDS_INNER_CLASS =
  "flex gap-6 mx-auto py-5 max-[768px]:gap-4";

export const OVERLAY_FOOTER_CLASS =
  "py-6 px-8 bg-black/40 border-t border-space-blue-600 flex justify-between items-center max-[768px]:p-5 max-[768px]:flex-col max-[768px]:gap-5";

export const OVERLAY_FOOTER_LEFT_CLASS =
  "flex gap-8 items-center max-[768px]:w-full max-[768px]:justify-between";

export const OVERLAY_FOOTER_RIGHT_CLASS =
  "flex items-center gap-6 max-[768px]:w-full max-[768px]:flex-col max-[768px]:gap-3";

export const PRIMARY_BUTTON_CLASS =
  "py-4 px-8 bg-space-black-darker/90 border-2 border-space-blue-800 rounded-xl text-xl font-bold text-white cursor-pointer transition-all duration-300 text-shadow-dark shadow-[0_4px_20px_rgba(30,60,150,0.3)] whitespace-nowrap hover:enabled:bg-space-black-darker/95 hover:enabled:border-space-blue-600 hover:enabled:-translate-y-0.5 hover:enabled:shadow-glow active:enabled:translate-y-0 disabled:bg-gray-700/50 disabled:border-gray-500/30 disabled:cursor-not-allowed disabled:transform-none disabled:shadow-none disabled:opacity-60 max-[768px]:w-full max-[768px]:py-3 max-[768px]:px-6 max-[768px]:text-lg";

export const SECONDARY_BUTTON_CLASS =
  "py-3 px-6 bg-space-black-darker/60 border-2 border-space-blue-800/60 rounded-lg text-white font-medium cursor-pointer transition-all duration-200 whitespace-nowrap hover:-translate-y-px hover:bg-space-black-darker/80 hover:border-space-blue-600 active:translate-y-0";

export const RESOURCE_LABEL_CLASS =
  "text-sm text-white/60 uppercase tracking-[0.5px]";

export const RESOURCE_DISPLAY_CLASS = "flex items-center gap-3";

export const INPUT_CLASS =
  "w-full bg-black/60 border border-space-blue-400/50 rounded-lg py-2 px-3 text-white text-sm outline-none focus:border-space-blue-400";

export const INPUT_SMALL_CLASS =
  "w-16 bg-black/60 border border-space-blue-400/50 rounded py-1 px-2 text-white text-sm text-center outline-none focus:border-space-blue-400";
