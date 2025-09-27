import { useMemo } from "react";
import { useGLTF } from "@react-three/drei";
import { useFrame, useThree } from "@react-three/fiber";
import * as THREE from "three";
import ProjectedHexGrid from "./ProjectedHexGrid.tsx";
import { GameDto } from "../../../types/generated/api-types.ts";
import { useMarsRotation } from "../../../contexts/MarsRotationContext.tsx";

interface MarsSphereProps {
  gameState?: GameDto;
  onHexClick?: (hex: string) => void;
}

export default function MarsSphere({ gameState, onHexClick }: MarsSphereProps) {
  const { marsGroupRef } = useMarsRotation();
  const { camera } = useThree();

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
        }
        child.material = material;

        // Enable shadow casting and receiving
        child.castShadow = true;
        child.receiveShadow = true;
      }
    });

    return clonedScene;
  }, [scene, marsColorTint]);

  // Create atmospheric fresnel effect material
  const atmosphereMaterial = useMemo(() => {
    return new THREE.ShaderMaterial({
      vertexShader: `
        varying vec3 vWorldPosition;
        varying vec3 vNormal;

        void main() {
          vNormal = normalize(normalMatrix * normal);
          vec4 worldPosition = modelMatrix * vec4(position, 1.0);
          vWorldPosition = worldPosition.xyz;
          gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
        }
      `,
      fragmentShader: `
        uniform vec3 cameraPosition;
        uniform vec3 atmosphereColor;
        uniform float intensity;

        varying vec3 vWorldPosition;
        varying vec3 vNormal;

        void main() {
          vec3 viewDirection = normalize(cameraPosition - vWorldPosition);
          float fresnel = 1.0 - abs(dot(viewDirection, vNormal));
          fresnel = pow(fresnel, 2.0);

          float alpha = fresnel * intensity;
          gl_FragColor = vec4(atmosphereColor, alpha);
        }
      `,
      uniforms: {
        cameraPosition: { value: new THREE.Vector3() },
        atmosphereColor: { value: new THREE.Color(0.8, 0.4, 0.2) }, // Mars-like atmospheric color
        intensity: { value: 0.6 },
      },
      transparent: true,
      blending: THREE.AdditiveBlending,
      side: THREE.BackSide, // Render from inside out for better effect
    });
  }, []);

  // Update camera position in shader uniforms for fresnel effect
  useFrame(() => {
    if (atmosphereMaterial.uniforms.cameraPosition) {
      atmosphereMaterial.uniforms.cameraPosition.value.copy(camera.position);
    }
  });

  return (
    <group ref={marsGroupRef}>
      {/* Mars GLB model with terraforming color tint */}
      <primitive object={marsScene} />

      {/* Atmospheric fresnel effect - slightly larger sphere */}
      <mesh scale={[2.08, 2.08, 2.08]}>
        <sphereGeometry args={[1, 64, 64]} />
        <primitive object={atmosphereMaterial} />
      </mesh>

      {/* Projected hexagonal grid "wrapped" around Mars sphere */}
      <ProjectedHexGrid gameState={gameState} onHexClick={onHexClick} />
    </group>
  );
}
