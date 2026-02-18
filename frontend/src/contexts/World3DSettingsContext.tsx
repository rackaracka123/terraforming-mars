import { createContext, useContext, useState, ReactNode } from "react";

export interface World3DSettings {
  sunDirectionX: number;
  sunDirectionY: number;
  sunDirectionZ: number;
  sunIntensity: number;
  sunColor: { r: number; g: number; b: number };
  waterColor: { r: number; g: number; b: number };
  reflectance: number;
  freeCameraEnabled: boolean;
  showCameraFrustum: boolean;
}

export interface StoredCameraState {
  position: { x: number; y: number; z: number };
  spherical: { radius: number; phi: number; theta: number };
}

const defaultSettings: World3DSettings = {
  sunDirectionX: 0.9,
  sunDirectionY: 0.0,
  sunDirectionZ: 0.8,
  sunIntensity: 1.0,
  sunColor: { r: 1.0, g: 0.86, b: 0.72 },
  waterColor: { r: 0.05, g: 0.09, b: 0.1 },
  reflectance: 0.1,
  freeCameraEnabled: false,
  showCameraFrustum: false,
};

interface World3DSettingsContextType {
  settings: World3DSettings;
  updateSettings: (partial: Partial<World3DSettings>) => void;
  resetSettings: () => void;
  storedCameraState: StoredCameraState | null;
  setStoredCameraState: (state: StoredCameraState | null) => void;
}

const World3DSettingsContext = createContext<World3DSettingsContextType | null>(null);

export function World3DSettingsProvider({ children }: { children: ReactNode }) {
  const [settings, setSettings] = useState<World3DSettings>(defaultSettings);
  const [storedCameraState, setStoredCameraState] = useState<StoredCameraState | null>(null);

  const updateSettings = (partial: Partial<World3DSettings>) => {
    setSettings((prev) => ({ ...prev, ...partial }));
  };

  const resetSettings = () => {
    setSettings(defaultSettings);
  };

  return (
    <World3DSettingsContext.Provider
      value={{ settings, updateSettings, resetSettings, storedCameraState, setStoredCameraState }}
    >
      {children}
    </World3DSettingsContext.Provider>
  );
}

export function useWorld3DSettings() {
  const context = useContext(World3DSettingsContext);
  if (!context) {
    throw new Error("useWorld3DSettings must be used within World3DSettingsProvider");
  }
  return context;
}
