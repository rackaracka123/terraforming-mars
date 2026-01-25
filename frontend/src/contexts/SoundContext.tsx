import React, { createContext, useCallback, useContext, useEffect, useState } from "react";
import { audioService } from "../services/audioService.ts";
import { getSoundSettings, saveSoundSettings, SoundSettings } from "../utils/soundStorage.ts";

interface SoundContextType {
  enabled: boolean;
  volume: number;
  toggleMute: () => void;
  setVolume: (volume: number) => void;
}

const SoundContext = createContext<SoundContextType | undefined>(undefined);

/**
 * SoundProvider - Manages global sound settings and syncs with audioService
 */
export function SoundProvider({ children }: { children: React.ReactNode }) {
  const [settings, setSettings] = useState<SoundSettings>(() => getSoundSettings());

  useEffect(() => {
    audioService.setEnabled(settings.enabled);
    audioService.setVolume(settings.volume);
    saveSoundSettings(settings);
  }, [settings]);

  const toggleMute = useCallback(() => {
    setSettings((prev) => ({
      ...prev,
      enabled: !prev.enabled,
    }));
  }, []);

  const setVolume = useCallback((volume: number) => {
    setSettings((prev) => ({
      ...prev,
      volume: Math.max(0, Math.min(1, volume)),
    }));
  }, []);

  const contextValue: SoundContextType = {
    enabled: settings.enabled,
    volume: settings.volume,
    toggleMute,
    setVolume,
  };

  return <SoundContext.Provider value={contextValue}>{children}</SoundContext.Provider>;
}

/**
 * Hook to access sound settings and controls
 */
export function useSound() {
  const context = useContext(SoundContext);
  if (context === undefined) {
    throw new Error("useSound must be used within a SoundProvider");
  }
  return context;
}
