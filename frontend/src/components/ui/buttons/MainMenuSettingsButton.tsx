import React, { useState, useRef, useEffect } from "react";
import SoundToggleButton from "./SoundToggleButton.tsx";
import { GamePopover } from "../GamePopover";

const MainMenuSettingsButton: React.FC = () => {
  const [menuOpen, setMenuOpen] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(!!document.fullscreenElement);
  const gearButtonRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement);
    };
    document.addEventListener("fullscreenchange", handleFullscreenChange);
    return () => document.removeEventListener("fullscreenchange", handleFullscreenChange);
  }, []);

  const handleEnterFullscreen = () => {
    void document.documentElement.requestFullscreen();
  };

  return (
    <>
      {!isFullscreen && (
        <div className="fixed top-[30px] left-1/2 -translate-x-1/2 z-50">
          <button
            onClick={handleEnterFullscreen}
            className="bg-space-black-darker/80 border border-white/20 rounded-lg px-4 py-2.5 text-white text-sm cursor-pointer hover:brightness-125 transition-all flex items-center gap-2"
          >
            <span>For optimal view, enter fullscreen</span>
            <svg
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <polyline points="15 3 21 3 21 9" />
              <polyline points="9 21 3 21 3 15" />
              <line x1="21" y1="3" x2="14" y2="10" />
              <line x1="3" y1="21" x2="10" y2="14" />
            </svg>
          </button>
        </div>
      )}
      <div className="fixed top-[30px] right-[30px] z-50">
        <button
          ref={gearButtonRef}
          onClick={() => setMenuOpen(!menuOpen)}
          className="bg-space-black-darker/80 border border-white/20 text-white p-2.5 rounded-lg cursor-pointer hover:bg-white/20 transition-colors"
          aria-label="Settings"
        >
          <svg
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <circle cx="12" cy="12" r="3" />
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
          </svg>
        </button>

        <GamePopover
          isVisible={menuOpen}
          onClose={() => setMenuOpen(false)}
          position={{ type: "anchor", anchorRef: gearButtonRef, placement: "below" }}
          theme="menu"
          width={200}
          maxHeight="auto"
          animation="slideDown"
          excludeRef={gearButtonRef}
        >
          <SoundToggleButton />
        </GamePopover>
      </div>
    </>
  );
};

export default MainMenuSettingsButton;
