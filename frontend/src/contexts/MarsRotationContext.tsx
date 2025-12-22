import { createContext, useContext, useRef, ReactNode } from "react";
import * as THREE from "three";

interface MarsRotationContextType {
  marsGroupRef: React.RefObject<THREE.Group | null>;
}

const MarsRotationContext = createContext<MarsRotationContextType | null>(null);

interface MarsRotationProviderProps {
  children: ReactNode;
}

export function MarsRotationProvider({ children }: MarsRotationProviderProps) {
  const marsGroupRef = useRef<THREE.Group>(null);

  return (
    <MarsRotationContext.Provider value={{ marsGroupRef }}>
      {children}
    </MarsRotationContext.Provider>
  );
}

export function useMarsRotation() {
  const context = useContext(MarsRotationContext);
  if (!context) {
    throw new Error(
      "useMarsRotation must be used within a MarsRotationProvider",
    );
  }
  return context;
}
