import React, { useRef, useMemo } from "react";
import { useGLTF } from "@react-three/drei";
import * as THREE from "three";
import HexagonalGrid from "./HexagonalGrid.tsx";
import { Game } from "../../../types/generated/domain";

interface MarsSphereProps {
  gameState?: Game;
  onHexClick?: (hex: string) => void;
}

export default function MarsSphere({ gameState, onHexClick }: MarsSphereProps) {
  const groupRef = useRef<THREE.Group>(null);

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

    // Calculate bounding box to normalize size to radius 2
    const box = new THREE.Box3().setFromObject(clonedScene);
    const size = box.getSize(new THREE.Vector3());
    const maxDimension = Math.max(size.x, size.y, size.z);
    const targetRadius = 2;
    const scale = (targetRadius * 2) / maxDimension;

    clonedScene.scale.setScalar(scale);

    // Apply terraforming color tint to all materials
    clonedScene.traverse((child) => {
      if (child instanceof THREE.Mesh && child.material) {
        const material = child.material.clone();
        if (material instanceof THREE.MeshStandardMaterial) {
          // Mix original color with terraforming progress tint
          const originalColor = material.color.clone();
          material.color = originalColor.lerp(marsColorTint, 0.3);
        }
        child.material = material;
      }
    });

    return clonedScene;
  }, [scene, marsColorTint]);

  return (
    <group ref={groupRef}>
      {/* Mars GLB model with terraforming color tint */}
      <primitive object={marsScene} />

      {/* Hexagonal grid overlay */}
      <HexagonalGrid gameState={gameState} onHexClick={onHexClick} />
    </group>
  );
}
