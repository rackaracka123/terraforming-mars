import { Suspense, useEffect, useState, useRef, useCallback } from "react";
import { Canvas } from "@react-three/fiber";
import { PanControls } from "../controls/PanControls.tsx";
import MarsSphere from "../board/MarsSphere.tsx";
import SkyboxLoader from "./SkyboxLoader.tsx";
import LoadingSpinner from "./LoadingSpinner.tsx";
import { GameDto } from "@/types/generated/api-types.ts";
import { MarsRotationProvider } from "../../../contexts/MarsRotationContext.tsx";
import {
  skyboxCache,
  SkyboxLoadingState,
} from "../../../services/SkyboxCache.ts";

interface Game3DViewProps {
  gameState: GameDto;
}

export default function Game3DView({ gameState }: Game3DViewProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [cameraConfig, setCameraConfig] = useState({
    position: [0, 0, 8] as [number, number, number],
    fov: 50,
  });
  const [skyboxLoadingState, setSkyboxLoadingState] =
    useState<SkyboxLoadingState>({
      isLoading: false,
      isLoaded: false,
      error: null,
      texture: null,
    });

  const updateCameraConfig = useCallback(() => {
    const width = window.innerWidth;
    let fov = 50;
    let position: [number, number, number] = [0, 0, 8];

    if (width <= 768) {
      fov = 60;
      position = [0, 0, 10];
    } else if (width <= 1200) {
      fov = 55;
      position = [0, 0, 9];
    }

    setCameraConfig({ position, fov });
  }, []);

  useEffect(() => {
    updateCameraConfig();
    window.addEventListener("resize", updateCameraConfig);

    return () => window.removeEventListener("resize", updateCameraConfig);
  }, [updateCameraConfig]);

  // Subscribe to skybox loading state
  useEffect(() => {
    const unsubscribe = skyboxCache.subscribe((state) => {
      setSkyboxLoadingState(state);
    });

    return unsubscribe;
  }, []);

  const handleHexClick = (hexCoordinate: string) => {
    console.warn("Hex clicked:", hexCoordinate);

    // TODO: Implement proper tile placement validation
    // For now, just show the coordinate that was clicked

    // Basic validation would check:
    // 1. Is it the current player's turn?
    // 2. Does the player have resources to place a tile?
    // 3. Is the tile placement valid based on game rules?
    // 4. Are there adjacency requirements?

    // Example future implementation:
    // if (canPlaceTile(hexCoordinate, gameState)) {
    //   // Send tile placement to server
    //   socket.emit('place-tile', {
    //     coordinate: hexCoordinate,
    //     tileType: selectedTileType,
    //     playerId: currentPlayer.id
    //   });
    // }
  };

  return (
    <div
      ref={containerRef}
      style={{
        flex: 1,
        height: "100%",
        width: "100%",
        minHeight: 0,
        position: "relative",
      }}
    >
      <Canvas
        camera={{
          position: cameraConfig.position,
          fov: cameraConfig.fov,
        }}
        style={{
          background: "#000000", // Fallback background
          width: "100%",
          height: "100%",
          position: "relative",
          zIndex: 0,
        }}
        resize={{ scroll: false, debounce: { scroll: 50, resize: 0 } }}
        dpr={typeof window !== "undefined" ? window.devicePixelRatio : 1}
        shadows
      >
        <MarsRotationProvider>
          <Suspense fallback={null}>
            {/* EXR Skybox - now uses cached texture */}
            <SkyboxLoader />

            {/* Realistic Lighting Setup */}
            {/* Very low ambient light for deep shadows */}
            <ambientLight intensity={0.03} color="#1a1a2e" />

            <directionalLight
              position={[8, 6, 15]}
              intensity={2.6}
              color="#fff8e1"
              castShadow
              shadow-mapSize-width={2048}
              shadow-mapSize-height={2048}
              shadow-camera-far={50}
              shadow-camera-left={-20}
              shadow-camera-right={20}
              shadow-camera-top={20}
              shadow-camera-bottom={-20}
            />

            {/* Cool blue rim light for moody atmosphere */}
            <directionalLight
              position={[-8, -3, -10]}
              intensity={0.35}
              color="#4488ff"
            />

            {/* Atmospheric fog for depth and mood */}
            <fog attach="fog" args={["#0a0a1a", 8, 25]} />

            {/* Mars with hexagonal tiles */}
            <MarsSphere gameState={gameState} onHexClick={handleHexClick} />

            {/* Orbital camera controls */}
            <PanControls />
          </Suspense>
        </MarsRotationProvider>
      </Canvas>

      {/* Show loading spinner when skybox is loading */}
      {skyboxLoadingState.isLoading && (
        <LoadingSpinner message="Loading 3D environment..." />
      )}
    </div>
  );
}
