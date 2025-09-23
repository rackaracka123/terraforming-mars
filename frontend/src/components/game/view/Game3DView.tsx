import { Suspense, useEffect, useState } from "react";
import { Canvas } from "@react-three/fiber";
import { PanControls } from "../controls/PanControls.tsx";
import MarsSphere from "../board/MarsSphere.tsx";
import { GameDto } from "@/types/generated/api-types.ts";

interface Game3DViewProps {
  gameState: GameDto;
}

export default function Game3DView({ gameState }: Game3DViewProps) {
  const [cameraConfig, setCameraConfig] = useState({
    position: [0, 0, 8] as [number, number, number],
    fov: 50,
  });

  useEffect(() => {
    const updateCameraConfig = () => {
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
    };

    updateCameraConfig();
    window.addEventListener("resize", updateCameraConfig);

    return () => window.removeEventListener("resize", updateCameraConfig);
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
      style={{
        flex: 1,
        height: "100%",
        width: "100%",
        minHeight: 0,
        position: "relative",
      }}
    >
      {/* Background layer with dark filter */}
      <div
        style={{
          position: "absolute",
          top: 0,
          left: 0,
          width: "100%",
          height: "100%",
          background: "url('/assets/backgrounds/space-bg.webp') center/cover no-repeat",
          filter: "brightness(0.5)",
          zIndex: 0,
        }}
      />
      <Canvas
        camera={{
          position: cameraConfig.position,
          fov: cameraConfig.fov,
        }}
        style={{
          background: "transparent",
          width: "100%",
          height: "100%",
          position: "relative",
          zIndex: 1,
        }}
        resize={{ scroll: false, debounce: { scroll: 50, resize: 0 } }}
      >
        <Suspense fallback={null}>
          {/* Lighting */}
          <ambientLight intensity={0.4} />
          <directionalLight position={[10, 10, 5]} intensity={1} castShadow />
          <pointLight position={[-10, -10, -5]} intensity={0.3} />

          {/* Mars with hexagonal tiles */}
          <MarsSphere gameState={gameState} onHexClick={handleHexClick} />

          {/* Pan and zoom controls */}
          <PanControls />
        </Suspense>
      </Canvas>
    </div>
  );
}
