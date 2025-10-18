import React, { useEffect, useRef, useState } from "react";
import audioSettingsManager from "../../../services/audioSettingsManager.ts";

interface SettingsDropdownProps {
  isVisible: boolean;
  onClose: () => void;
  buttonRef: React.RefObject<HTMLButtonElement | null>;
}

const SettingsDropdown: React.FC<SettingsDropdownProps> = ({
  isVisible,
  onClose,
  buttonRef,
}) => {
  const dropdownRef = useRef<HTMLDivElement>(null);
  const [settings, setSettings] = useState(audioSettingsManager.getSettings());

  // Close dropdown when clicking outside
  useEffect(() => {
    if (!isVisible) return;

    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node) &&
        buttonRef.current &&
        !buttonRef.current.contains(event.target as Node)
      ) {
        onClose();
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isVisible, onClose, buttonRef]);

  if (!isVisible) return null;

  const handleSoundEffectsToggle = () => {
    const newEnabled = !settings.soundEffects.enabled;
    audioSettingsManager.setSoundEffectsEnabled(newEnabled);
    setSettings(audioSettingsManager.getSettings());
  };

  const handleBackgroundMusicToggle = () => {
    const newEnabled = !settings.backgroundMusic.enabled;
    audioSettingsManager.setBackgroundMusicEnabled(newEnabled);
    setSettings(audioSettingsManager.getSettings());
  };

  return (
    <div
      ref={dropdownRef}
      className="fixed bg-space-black-darker/95 border-2 border-space-blue-400 rounded-lg p-4 shadow-[0_8px_32px_rgba(0,0,0,0.6),0_0_20px_rgba(30,60,150,0.3)] backdrop-blur-space"
      style={{
        top: buttonRef.current
          ? `${buttonRef.current.getBoundingClientRect().bottom + 8}px`
          : "68px",
        right: "20px",
        zIndex: 100000,
        minWidth: "280px",
      }}
    >
      <h3 className="text-white font-orbitron text-base font-bold mb-4 pb-3 border-b border-space-blue-400/30">
        ⚙️ Settings
      </h3>

      {/* Sound Effects Toggle */}
      <div className="mb-4">
        <div className="flex justify-between items-center mb-2">
          <label className="text-white text-sm font-medium">
            Sound Effects
          </label>
          <button
            onClick={handleSoundEffectsToggle}
            className={`relative w-12 h-6 rounded-full transition-all duration-300 ${
              settings.soundEffects.enabled
                ? "bg-space-blue-600"
                : "bg-gray-600"
            }`}
          >
            <div
              className={`absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full transition-transform duration-300 ${
                settings.soundEffects.enabled
                  ? "translate-x-6"
                  : "translate-x-0"
              }`}
            />
          </button>
        </div>
        <p className="text-white/60 text-xs">
          Button clicks and game sound effects
        </p>
      </div>

      {/* Background Music Toggle */}
      <div className="mb-4">
        <div className="flex justify-between items-center mb-2">
          <label className="text-white text-sm font-medium">
            Background Music
          </label>
          <button
            onClick={handleBackgroundMusicToggle}
            className={`relative w-12 h-6 rounded-full transition-all duration-300 ${
              settings.backgroundMusic.enabled
                ? "bg-space-blue-600"
                : "bg-gray-600"
            }`}
          >
            <div
              className={`absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full transition-transform duration-300 ${
                settings.backgroundMusic.enabled
                  ? "translate-x-6"
                  : "translate-x-0"
              }`}
            />
          </button>
        </div>
        <p className="text-white/60 text-xs">Menu and lobby background music</p>
      </div>

      {/* Footer */}
      <div className="pt-3 border-t border-space-blue-400/30">
        <p className="text-white/40 text-[10px] text-center">
          Settings saved automatically
        </p>
      </div>
    </div>
  );
};

export default SettingsDropdown;
