import { useMemo } from "react";
import { useGLTF } from "@react-three/drei";
import * as THREE from "three";
import ProjectedHexGrid from "./ProjectedHexGrid.tsx";
import { TileHighlightMode } from "./ProjectedHexTile.tsx";
import { TileVPIndicator } from "../../ui/overlay/EndGameOverlay.tsx";
import { GameDto } from "../../../types/generated/api-types.ts";
import { useMarsRotation } from "../../../contexts/MarsRotationContext.tsx";

interface MarsSphereProps {
  gameState?: GameDto;
  onHexClick?: (hex: string) => void;
  tileHighlightMode?: TileHighlightMode;
  vpIndicators?: TileVPIndicator[];
}

export default function MarsSphere({
  gameState,
  onHexClick,
  tileHighlightMode,
  vpIndicators = [],
}: MarsSphereProps) {
  const { marsGroupRef } = useMarsRotation();

  // Load the Mars GLTF model
  const { scene } = useGLTF("/assets/models/mars.glb");

  // Disabled rotation for better tile visibility
  // useFrame((state) => {
  //   if (groupRef.current) {
  //     groupRef.current.rotation.y = state.clock.elapsedTime * 0.05;
  //   }
  // });

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

  // Clone the scene to avoid modifying the original
  const marsScene = useMemo(() => {
    const clonedScene = scene.clone();

    // Calculate bounding box to normalize size to radius 2.05 (slightly larger to align with hex tiles)
    const box = new THREE.Box3().setFromObject(clonedScene);
    const size = box.getSize(new THREE.Vector3());
    const maxDimension = Math.max(size.x, size.y, size.z);
    const targetRadius = 2.017;
    const scale = (targetRadius * 2) / maxDimension;

    clonedScene.scale.setScalar(scale);

    // Rotate Mars 65 degrees to show a brighter area
    clonedScene.rotation.y = (65 * Math.PI) / 180; // Convert degrees to radians

    // Apply terraforming color tint and configure shadows for all materials
    clonedScene.traverse((child) => {
      if (child instanceof THREE.Mesh && child.material) {
        const material = child.material.clone();
        if (material instanceof THREE.MeshStandardMaterial) {
          // Mix original color with terraforming progress tint
          const originalColor = material.color.clone();
          material.color = originalColor.lerp(marsColorTint, 0.3);

          // Enhance material properties for better lighting response
          material.roughness = 0.8; // Mars surface is rough
          material.metalness = 0.1; // Very low metalness for rock/soil

          // Fix texture encoding for sRGB textures
          if (material.map) {
            material.map.colorSpace = THREE.SRGBColorSpace;
          }
        }
        child.material = material;

        // Enable shadow casting and receiving
        child.castShadow = true;
        child.receiveShadow = true;
      }
    });

    return clonedScene;
  }, [scene, marsColorTint]);

  return (
    <group ref={marsGroupRef}>
      {/* Mars GLB model with terraforming color tint */}
      <primitive object={marsScene} />

      {/* Projected hexagonal grid "wrapped" around Mars sphere */}
      <ProjectedHexGrid
        gameState={gameState}
        onHexClick={onHexClick}
        tileHighlightMode={tileHighlightMode}
        vpIndicators={vpIndicators}
      />
    </group>
  );
}
