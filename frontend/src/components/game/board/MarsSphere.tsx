import { useMemo } from "react";
import * as THREE from "three";
import TileGrid from "./TileGrid.tsx";
import { TileHighlightMode } from "./Tile.tsx";
import { TileVPIndicator } from "../../ui/overlay/EndGameOverlay.tsx";
import { GameDto } from "../../../types/generated/api-types.ts";
import { useMarsRotation } from "../../../contexts/MarsRotationContext.tsx";
import { useTextures } from "../../../hooks/useTextures.ts";

// Sphere radius - must match TileGrid.tsx SPHERE_RADIUS
const MARS_RADIUS = 2.02;

interface MarsSphereProps {
  gameState?: GameDto;
  onHexClick?: (hex: string) => void;
  tileHighlightMode?: TileHighlightMode;
  vpIndicators?: TileVPIndicator[];
  animateHexEntrance?: boolean;
}

export default function MarsSphere({
  gameState,
  onHexClick,
  tileHighlightMode,
  vpIndicators = [],
  animateHexEntrance = false,
}: MarsSphereProps) {
  const { marsGroupRef } = useMarsRotation();

  const { mars: diffuseMap } = useTextures();

  // Get Mars color based on terraforming progress for tinting
  const marsColorTint = useMemo(() => {
    const temp = gameState?.globalParameters?.temperature || -30;
    const oxygen = gameState?.globalParameters?.oxygen || 0;

    const tempProgress = Math.max(0, (temp + 30) / 38);
    const oxygenProgress = oxygen / 14;

    const red = 1 - tempProgress * 0.3;
    const green = tempProgress * 0.2 + oxygenProgress * 0.3;
    const blue = oxygenProgress * 0.2;

    return new THREE.Color(red, green, blue);
  }, [gameState?.globalParameters]);

  // Create smooth sphere geometry with high segment count
  const sphereGeometry = useMemo(() => {
    // 128 width segments, 64 height segments for smooth appearance
    return new THREE.SphereGeometry(MARS_RADIUS, 128, 64);
  }, []);

  // Create material with terraforming tint
  const marsMaterial = useMemo(() => {
    const baseMarsColor = new THREE.Color(1, 1, 1);
    const tintedColor = baseMarsColor.lerp(marsColorTint, 0.3);

    return new THREE.MeshStandardMaterial({
      map: diffuseMap,
      color: tintedColor,
      roughness: 0.85,
      metalness: 0.05,
    });
  }, [diffuseMap, marsColorTint]);

  return (
    <group ref={marsGroupRef}>
      {/* Smooth Mars sphere with textures */}
      <mesh
        geometry={sphereGeometry}
        material={marsMaterial}
        rotation={[0, (65 * Math.PI) / 180, 0]}
        castShadow
        receiveShadow
      />

      {/* Projected hexagonal grid "wrapped" around Mars sphere */}
      <TileGrid
        gameState={gameState}
        onHexClick={onHexClick}
        tileHighlightMode={tileHighlightMode}
        vpIndicators={vpIndicators}
        animateHexEntrance={animateHexEntrance}
      />
    </group>
  );
}
