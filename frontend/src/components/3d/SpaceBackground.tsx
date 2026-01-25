import { Suspense, useEffect, useState, useRef } from "react";
import { Canvas, useFrame, useThree } from "@react-three/fiber";
import SkyboxLoader from "../game/view/SkyboxLoader.tsx";
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
  const [skyboxLoadingState, setSkyboxLoadingState] = useState<SkyboxLoadingState>({
    isLoading: false,
    isLoaded: false,
    error: null,
    texture: null,
  });

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
          <SkyboxLoader />
          <AnimatedCamera speed={animationSpeed} />
          <ambientLight intensity={0.02} color="#0a0a1a" />
          <directionalLight position={[10, 10, 10]} intensity={0.3} color="#1a1a3e" />
        </Suspense>
      </Canvas>

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

          {skyboxLoadingState.isLoading && (
            <div
              style={{
                position: "fixed",
                bottom: "200px",
                left: "50%",
                transform: "translateX(-50%)",
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
                gap: "12px",
                zIndex: 10,
              }}
            >
              <div
                style={{
                  width: "40px",
                  height: "40px",
                  border: "4px solid rgba(255, 255, 255, 0.2)",
                  borderTop: "4px solid rgba(255, 255, 255, 0.9)",
                  borderRadius: "50%",
                  animation: "spin 1s linear infinite",
                }}
              />
              <style>
                {`
                  @keyframes spin {
                    0% { transform: rotate(0deg); }
                    100% { transform: rotate(360deg); }
                  }
                `}
              </style>
              <div
                style={{
                  color: "rgba(255, 255, 255, 0.8)",
                  fontSize: "14px",
                  fontFamily: "Orbitron, sans-serif",
                  letterSpacing: "0.05em",
                }}
              >
                Loading...
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
