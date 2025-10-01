import { Suspense, useEffect, useState, useCallback, useRef } from "react";
import { Canvas, useFrame, useThree } from "@react-three/fiber";
import SkyboxLoader from "../game/view/SkyboxLoader.tsx";
import LoadingSpinner from "../game/view/LoadingSpinner.tsx";
import { skyboxCache, SkyboxLoadingState } from "../../services/SkyboxCache.ts";
import * as THREE from "three";

interface SpaceBackgroundProps {
  animationSpeed?: number;
  overlayOpacity?: number;
  showOverlay?: boolean;
  children?: React.ReactNode;
}

/**
 * Animated camera component that creates slow panning motion
 */
function AnimatedCamera({ speed }: { speed: number }) {
  const { camera } = useThree();
  const spherical = useRef(
    new THREE.Spherical(12, Math.PI / 2, 0), // radius, phi (vertical), theta (horizontal)
  );

  useFrame((state) => {
    const time = state.clock.getElapsedTime();

    // Slow horizontal rotation (theta)
    spherical.current.theta = time * speed * 0.05;

    // Subtle vertical oscillation (phi) - stays near equator
    spherical.current.phi = Math.PI / 2 + Math.sin(time * speed * 0.03) * 0.1;

    // Update camera position from spherical coordinates
    camera.position.setFromSpherical(spherical.current);
    camera.lookAt(0, 0, 0);
  });

  return null;
}

/**
 * SpaceBackground component - Reusable 3D space environment with animated camera
 * Used across landing page, create/join pages
 */
export default function SpaceBackground({
  animationSpeed = 1,
  overlayOpacity = 0.2,
  showOverlay = true,
  children,
}: SpaceBackgroundProps) {
  const [cameraConfig] = useState({
    position: [0, 0, 12] as [number, number, number],
    fov: 60,
  });
  const [skyboxLoadingState, setSkyboxLoadingState] =
    useState<SkyboxLoadingState>({
      isLoading: false,
      isLoaded: false,
      error: null,
      texture: null,
    });

  // Subscribe to skybox loading state
  useEffect(() => {
    const unsubscribe = skyboxCache.subscribe((state) => {
      setSkyboxLoadingState(state);
    });

    return unsubscribe;
  }, []);

  return (
    <div
      style={{
        position: "fixed",
        top: 0,
        left: 0,
        width: "100vw",
        height: "100vh",
        zIndex: 0,
      }}
    >
      <Canvas
        camera={{
          position: cameraConfig.position,
          fov: cameraConfig.fov,
        }}
        style={{
          background: "#000000",
          width: "100%",
          height: "100%",
        }}
        dpr={typeof window !== "undefined" ? window.devicePixelRatio : 1}
      >
        <Suspense fallback={null}>
          {/* EXR Skybox */}
          <SkyboxLoader />

          {/* Animated camera for panning effect */}
          <AnimatedCamera speed={animationSpeed} />

          {/* Minimal ambient lighting for depth */}
          <ambientLight intensity={0.02} color="#0a0a1a" />

          {/* Directional light for subtle depth */}
          <directionalLight
            position={[10, 10, 10]}
            intensity={0.3}
            color="#1a1a3e"
          />
        </Suspense>
      </Canvas>

      {/* Dark overlay to reduce detail */}
      {showOverlay && (
        <div
          style={{
            position: "absolute",
            top: 0,
            left: 0,
            width: "100%",
            height: "100%",
            backgroundColor: `rgba(0, 0, 0, ${overlayOpacity})`,
            pointerEvents: "none",
          }}
        />
      )}

      {/* Show loading spinner when skybox is loading */}
      {skyboxLoadingState.isLoading && (
        <LoadingSpinner message="Loading space environment..." />
      )}

      {/* Content layer - always on top */}
      <div
        style={{
          position: "absolute",
          top: 0,
          left: 0,
          width: "100%",
          height: "100%",
          pointerEvents: "none",
          zIndex: 1,
        }}
      >
        <div
          style={{
            pointerEvents: "auto",
            width: "100%",
            height: "100%",
          }}
        >
          {children}
        </div>
      </div>
    </div>
  );
}
