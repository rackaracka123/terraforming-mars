import React from "react";
import { useSound } from "../../../contexts/SoundContext.tsx";

/**
 * Sound control for TopMenuBar dropdown menu
 * Clickable speaker icon for mute/unmute + volume slider
 */
const SoundToggleButton: React.FC = () => {
  const { enabled, volume, toggleMute, setVolume } = useSound();

  const handleVolumeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setVolume(parseFloat(e.target.value));
  };

  // Determine which speaker icon to show based on mute state and volume
  const getSpeakerIcon = () => {
    if (!enabled || volume === 0) {
      // Muted speaker icon
      return (
        <svg
          width="18"
          height="18"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5" />
          <line x1="23" y1="9" x2="17" y2="15" />
          <line x1="17" y1="9" x2="23" y2="15" />
        </svg>
      );
    } else if (volume < 0.5) {
      // Low volume speaker icon
      return (
        <svg
          width="18"
          height="18"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5" />
          <path d="M15.54 8.46a5 5 0 0 1 0 7.07" />
        </svg>
      );
    } else {
      // High volume speaker icon
      return (
        <svg
          width="18"
          height="18"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5" />
          <path d="M15.54 8.46a5 5 0 0 1 0 7.07" />
          <path d="M19.07 4.93a10 10 0 0 1 0 14.14" />
        </svg>
      );
    }
  };

  return (
    <div className="flex items-center gap-3 px-4 py-3 text-white">
      <button
        onClick={toggleMute}
        className="flex-shrink-0 hover:text-space-blue-400 transition-colors cursor-pointer"
        aria-label={enabled ? "Mute sound" : "Unmute sound"}
      >
        {getSpeakerIcon()}
      </button>
      <input
        type="range"
        min="0"
        max="1"
        step="0.05"
        value={enabled ? volume : 0}
        onChange={handleVolumeChange}
        className="w-full h-1.5 bg-white/20 rounded-full appearance-none cursor-pointer [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-3 [&::-webkit-slider-thumb]:h-3 [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-white [&::-webkit-slider-thumb]:cursor-pointer [&::-webkit-slider-thumb]:hover:bg-space-blue-400 [&::-moz-range-thumb]:w-3 [&::-moz-range-thumb]:h-3 [&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:bg-white [&::-moz-range-thumb]:border-0 [&::-moz-range-thumb]:cursor-pointer [&::-moz-range-thumb]:hover:bg-space-blue-400"
        aria-label="Volume"
      />
    </div>
  );
};

export default SoundToggleButton;
