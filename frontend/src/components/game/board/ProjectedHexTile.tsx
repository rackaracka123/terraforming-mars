import { useRef, useState, useMemo } from "react";
import { useFrame, useLoader } from "@react-three/fiber";
import { Text, useGLTF } from "@react-three/drei";
import * as THREE from "three";
import { HexTile2D } from "../../../utils/hex-grid-2d";
import { SkeletonUtils } from "three-stdlib";

// Preload 3D models for better performance
useGLTF.preload("/assets/models/city.glb");
useGLTF.preload("/assets/models/forrest.glb");
useGLTF.preload("/assets/models/water.glb");

interface ProjectedHexTileData extends HexTile2D {
  spherePosition: THREE.Vector3;
  normal: THREE.Vector3;
}

interface ProjectedHexTileProps {
  tileData: ProjectedHexTileData;
  tileType: "empty" | "ocean" | "greenery" | "city" | "special";
  ownerId?: string | null;
  displayName?: string;
  onClick: () => void;
  isAvailableForPlacement?: boolean;
}

export default function ProjectedHexTile({
  tileData,
  tileType,
  ownerId,
  displayName,
  onClick,
  isAvailableForPlacement = false,
}: ProjectedHexTileProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  const [hovered, setHovered] = useState(false);

  // Create hexagon geometry that's oriented along the surface normal
  const hexGeometry = useMemo(() => {
    const geometry = new THREE.CircleGeometry(0.166, 6);
    // Rotate to pointy-top orientation
    geometry.rotateZ(Math.PI / 2);
    return geometry;
  }, []);

  // Create hex border
  const borderGeometry = useMemo(() => {
    const geometry = new THREE.RingGeometry(0.156, 0.166, 6);
    geometry.rotateZ(Math.PI / 2);
    return geometry;
  }, []);

  // Create ocean gradient geometry - full hex circle for gradient effect
  const oceanGradientGeometry = useMemo(() => {
    const geometry = new THREE.CircleGeometry(0.166, 6); // Full circle for gradient
    geometry.rotateZ(Math.PI / 2);
    return geometry;
  }, []);

  // Create gradient ocean border material
  const oceanBorderMaterial = useMemo(() => {
    return new THREE.ShaderMaterial({
      vertexShader: `
        varying vec2 vUv;
        void main() {
          vUv = uv;
          gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
        }
      `,
      fragmentShader: `
        uniform float time;
        varying vec2 vUv;

        void main() {
          // Convert UV to centered coordinates (-0.5 to 0.5)
          vec2 center = vUv - 0.5;

          // Calculate distance from center (0.0 at center, ~0.5 at edges)
          float distFromCenter = length(center);

          // Create gradient that's strong at edges and fades toward center
          float gradient = smoothstep(0.2, 0.45, distFromCenter); // Faster fading

          vec3 darkBlue = vec3(0.05, 0.28, 0.63); // Dark blue #0d47a1

          vec3 finalColor = darkBlue;
          float alpha = gradient * 0.56; // Static alpha at lowest pulse value (0.8 * 0.7 = 0.56)

          gl_FragColor = vec4(finalColor, alpha);
        }
      `,
      uniforms: {
        time: { value: 0.0 },
      },
      transparent: true,
      side: THREE.DoubleSide,
    });
  }, []);

  // Create hover glow material with pulsation
  const hoverGlowMaterial = useMemo(() => {
    return new THREE.ShaderMaterial({
      vertexShader: `
        varying vec2 vUv;
        void main() {
          vUv = uv;
          gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
        }
      `,
      fragmentShader: `
        uniform float time;
        uniform float opacity;
        varying vec2 vUv;

        void main() {
          // Convert UV to centered coordinates (-0.5 to 0.5)
          vec2 center = vUv - 0.5;

          // Calculate distance from center (0.0 at center, ~0.5 at edges)
          float distFromCenter = length(center);

          // Create gradient that's strong at edges and fades toward center
          float gradient = smoothstep(0.15, 0.45, distFromCenter);

          vec3 glowColor = vec3(0.95, 0.95, 1.0); // Slightly blue-white glow

          vec3 finalColor = glowColor;
          float alpha = gradient * opacity;

          gl_FragColor = vec4(finalColor, alpha);
        }
      `,
      uniforms: {
        time: { value: 0.0 },
        opacity: { value: 0.0 },
      },
      transparent: true,
      side: THREE.DoubleSide,
      blending: THREE.AdditiveBlending,
    });
  }, []);

  // Create available placement glow material with pulsing effect
  const availableGlowMaterial = useMemo(() => {
    return new THREE.ShaderMaterial({
      vertexShader: `
        varying vec2 vUv;
        void main() {
          vUv = uv;
          gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
        }
      `,
      fragmentShader: `
        uniform float time;
        varying vec2 vUv;

        void main() {
          // Convert UV to centered coordinates (-0.5 to 0.5)
          vec2 center = vUv - 0.5;

          // Calculate distance from center (0.0 at center, ~0.5 at edges)
          float distFromCenter = length(center);

          // Create gradient that's strong at edges and fades toward center
          float gradient = smoothstep(0.15, 0.45, distFromCenter);

          // Pulsing animation
          float pulse = 0.5 + 0.3 * sin(time * 2.0);

          vec3 glowColor = vec3(0.4, 1.0, 0.4); // Green glow

          vec3 finalColor = glowColor;
          float alpha = gradient * pulse * 0.6; // Pulsing green glow

          gl_FragColor = vec4(finalColor, alpha);
        }
      `,
      uniforms: {
        time: { value: 0.0 },
      },
      transparent: true,
      side: THREE.DoubleSide,
      blending: THREE.AdditiveBlending,
    });
  }, []);

  // Update shader time uniforms and handle hover animations
  useFrame((state) => {
    if (oceanBorderMaterial.uniforms) {
      oceanBorderMaterial.uniforms.time.value = state.clock.elapsedTime;
    }

    if (hoverGlowMaterial.uniforms) {
      hoverGlowMaterial.uniforms.time.value = state.clock.elapsedTime;

      // Animate opacity based on hover state - reduced intensity
      const targetOpacity = hovered ? 0.3 : 0.0;
      hoverGlowMaterial.uniforms.opacity.value = THREE.MathUtils.lerp(
        hoverGlowMaterial.uniforms.opacity.value,
        targetOpacity,
        0.15,
      );
    }

    if (availableGlowMaterial.uniforms) {
      availableGlowMaterial.uniforms.time.value = state.clock.elapsedTime;
    }
  });

  // Calculate orientation quaternion to align with sphere surface
  const surfaceQuaternion = useMemo(() => {
    const up = new THREE.Vector3(0, 0, 1);
    const quaternion = new THREE.Quaternion();
    quaternion.setFromUnitVectors(up, tileData.normal);
    return quaternion;
  }, [tileData.normal]);

  // Position slightly above sphere surface
  const adjustedPosition = useMemo(() => {
    return tileData.spherePosition
      .clone()
      .add(tileData.normal.clone().multiplyScalar(0.01));
  }, [tileData.spherePosition, tileData.normal]);

  // Get tile color
  const tileColor = useMemo(() => {
    if (hovered) return new THREE.Color("#ffff88");

    switch (tileType) {
      case "ocean":
        return new THREE.Color("#1e88e5");
      case "greenery":
        return new THREE.Color("#43a047");
      case "city":
        return new THREE.Color("#ff6f00");
      case "special":
        return new THREE.Color("#8e24aa");
      default:
        return tileData.isOceanSpace
          ? new THREE.Color("#6d4c41").multiplyScalar(0.8) // Lighter brown
          : new THREE.Color("#6d4c41").multiplyScalar(0.8);
    }
  }, [tileType, tileData.isOceanSpace, hovered]);

  // Border color - darker for better visibility
  const borderColor = useMemo(() => {
    if (hovered) return new THREE.Color("#ffffff");
    return tileColor.clone().multiplyScalar(0.25);
  }, [tileColor, hovered]);

  // Bonus icons data - create individual icons for each bonus count
  const bonusIcons = useMemo(() => {
    const entries = Object.entries(tileData.bonuses);
    if (entries.length === 0) return [];

    const icons: string[] = [];
    entries.forEach(([key, value]) => {
      const iconPaths: { [key: string]: string } = {
        steel: "/assets/resources/steel.png",
        titanium: "/assets/resources/titanium.png",
        plants: "/assets/resources/plant.png",
        cards: "/assets/resources/card.png",
      };
      const iconPath = iconPaths[key] || "/assets/resources/megacredit.png";

      // Add the icon multiple times based on value
      for (let i = 0; i < value; i++) {
        icons.push(iconPath);
      }
    });

    return icons;
  }, [tileData.bonuses]);

  return (
    <group position={adjustedPosition} quaternion={surfaceQuaternion}>
      {/* Main hex tile */}
      <mesh
        ref={meshRef}
        geometry={hexGeometry}
        onPointerEnter={() => {
          setHovered(true);
        }}
        onPointerLeave={() => {
          setHovered(false);
        }}
        onClick={(event) => {
          event.stopPropagation();
          onClick();
        }}
      >
        <meshStandardMaterial
          color={tileColor}
          transparent
          opacity={tileType === "empty" ? 0.3 : 0.7}
          roughness={0.7}
          metalness={0.1}
          side={THREE.DoubleSide}
        />
      </mesh>

      {/* Hex border */}
      <mesh geometry={borderGeometry} position={[0, 0, 0.001]}>
        <meshBasicMaterial color={borderColor} transparent opacity={0.9} />
      </mesh>

      {/* Ocean space indicator - blue gradient fading to center */}
      {tileType === "empty" && tileData.isOceanSpace && (
        <mesh
          position={[0, 0, 0.002]}
          geometry={oceanGradientGeometry}
          material={oceanBorderMaterial}
        />
      )}

      {/* Hover glow effect with pulsation */}
      <mesh
        position={[0, 0, 0.0015]}
        geometry={oceanGradientGeometry}
        material={hoverGlowMaterial}
      />

      {/* Available placement glow effect with pulsing animation */}
      {isAvailableForPlacement && (
        <mesh
          position={[0, 0, 0.002]}
          geometry={oceanGradientGeometry}
          material={availableGlowMaterial}
        />
      )}

      {/* Tile type 3D model (city, greenery, ocean) */}
      {(tileType === "city" ||
        tileType === "greenery" ||
        tileType === "ocean") && (
        <TileModel tileType={tileType} position={[0, 0, 0.03]} />
      )}

      {/* Special tile fallback emoji (no 3D model available) */}
      {tileType === "special" && (
        <Text
          position={[0, 0, 0.01]}
          fontSize={0.08}
          color="white"
          anchorX="center"
          anchorY="middle"
        >
          ⭐
        </Text>
      )}

      {/* Display name and bonus icons layout */}
      {displayName && bonusIcons.length > 0 ? (
        // Two-row layout when both displayName and bonuses exist
        <>
          {/* Display name in upper row */}
          <Text
            position={[0, 0.03, 0.01]}
            fontSize={0.035}
            color="white"
            anchorX="center"
            anchorY="middle"
            maxWidth={0.25}
          >
            {displayName}
          </Text>

          {/* Bonus icons in lower row */}
          {bonusIcons.map((iconPath, index) => (
            <BonusIcon
              key={index}
              iconPath={iconPath}
              position={[
                index * 0.05 - (bonusIcons.length - 1) * 0.025,
                -0.03,
                0.01,
              ]}
            />
          ))}
        </>
      ) : displayName ? (
        // Display name centered when no bonuses
        <Text
          position={[0, 0, 0.01]}
          fontSize={0.04}
          color="white"
          anchorX="center"
          anchorY="middle"
          maxWidth={0.25}
        >
          {displayName}
        </Text>
      ) : (
        // Only bonus icons when no display name
        bonusIcons.map((iconPath, index) => (
          <BonusIcon
            key={index}
            iconPath={iconPath}
            position={[index * 0.05 - (bonusIcons.length - 1) * 0.025, 0, 0.01]}
          />
        ))
      )}

      {/* Owner indicator */}
      {ownerId && (
        <mesh position={[0.1, 0.1, 0.01]}>
          <circleGeometry args={[0.02, 16]} />
          <meshBasicMaterial
            color={`hsl(${(ownerId.charCodeAt(0) * 137.5) % 360}, 70%, 50%)`}
          />
        </mesh>
      )}
    </group>
  );
}

// Component for rendering 3D tile models (city, greenery, ocean)
interface TileModelProps {
  tileType: "city" | "greenery" | "ocean";
  position: [number, number, number];
}

function TileModel({ tileType, position }: TileModelProps) {
  // Load appropriate model based on tile type
  const modelPath =
    tileType === "city"
      ? "/assets/models/city.glb"
      : tileType === "greenery"
        ? "/assets/models/forrest.glb"
        : "/assets/models/water.glb";

  const { scene } = useGLTF(modelPath);

  // Clone and configure the model with proper transformations
  const configuredModel = useMemo(() => {
    // Clone the scene properly
    const clonedScene = SkeletonUtils.clone(scene);

    // Calculate bounding box and scale - larger size for greenery
    const targetSize = tileType === "greenery" ? 0.3 : 0.2;
    const box = new THREE.Box3().setFromObject(clonedScene);
    const size = box.getSize(new THREE.Vector3());
    const maxDimension = Math.max(size.x, size.y, size.z);
    const scaleFactor = targetSize / maxDimension;

    // Apply scale to the scene
    clonedScene.scale.setScalar(scaleFactor);

    // Recalculate center after scaling
    const scaledBox = new THREE.Box3().setFromObject(clonedScene);
    const scaledCenter = scaledBox.getCenter(new THREE.Vector3());

    // Position to center the model at origin
    clonedScene.position.set(-scaledCenter.x, -scaledCenter.y, -scaledCenter.z);

    // Enable shadows and apply material modifications
    clonedScene.traverse((child) => {
      if (child instanceof THREE.Mesh) {
        child.castShadow = true;
        child.receiveShadow = true;

        // Apply water-specific material properties for ocean tiles
        if (tileType === "ocean" && child.material) {
          const material = child.material as THREE.MeshStandardMaterial;
          if (material.isMeshStandardMaterial) {
            // Make water more transparent and reflective
            material.transparent = true;
            material.opacity = 0.7;
            material.metalness = 0.3;
            material.roughness = 0.2;
            // Tint with blue color
            material.color = new THREE.Color(0.2, 0.5, 0.8);
          }
        }
      }
    });

    return clonedScene;
  }, [scene, tileType]);

  // Rotation for specific tile types
  const rotation: [number, number, number] = useMemo(() => {
    if (tileType === "city") {
      return [-Math.PI / 2, 0, 0]; // -90° rotation on x-axis for city buildings
    }
    if (tileType === "greenery") {
      return [Math.PI / 2, 0, 0]; // +90° rotation on x-axis to flip trees right-side up
    }
    if (tileType === "ocean") {
      return [-Math.PI / 2, 0, 0]; // -90° rotation to lay water flat on surface
    }
    return [0, 0, 0];
  }, [tileType]);

  return (
    <group position={position} rotation={rotation}>
      <primitive object={configuredModel} />
    </group>
  );
}

// Component for rendering bonus icons with proper aspect ratio
interface BonusIconProps {
  iconPath: string;
  position: [number, number, number];
}

function BonusIcon({ iconPath, position }: BonusIconProps) {
  const texture = useLoader(THREE.TextureLoader, iconPath);

  // Calculate proper dimensions maintaining aspect ratio
  const dimensions = useMemo((): [number, number] => {
    if (!texture.image) return [0.05, 0.05];

    const aspect = texture.image.width / texture.image.height;
    const maxSize = 0.05;

    if (aspect > 1) {
      // Wide image
      return [maxSize, maxSize / aspect];
    } else {
      // Tall image
      return [maxSize * aspect, maxSize];
    }
  }, [texture]);

  // Create shadow gradient geometry (larger for more pronounced effect)
  const shadowDimensions = useMemo(
    (): [number, number] => [dimensions[0] * 1.5, dimensions[1] * 1.5],
    [dimensions],
  );

  // Create shadow gradient material
  const shadowMaterial = useMemo(() => {
    return new THREE.ShaderMaterial({
      vertexShader: `
        varying vec2 vUv;
        void main() {
          vUv = uv;
          gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
        }
      `,
      fragmentShader: `
        varying vec2 vUv;

        void main() {
          // Convert UV to centered coordinates (-0.5 to 0.5)
          vec2 center = vUv - 0.5;

          // Calculate square distance (max of x and y distances for square shape)
          float distFromCenter = max(abs(center.x), abs(center.y));

          // Create border gradient - only show shadow outside the icon area
          // Inner radius (icon size) to outer radius (shadow extent)
          float innerRadius = 0.33; // Approximate icon boundary
          float outerRadius = 0.5;   // Shadow extent

          // Only show gradient in the border area (between inner and outer radius)
          float borderGradient = 1.0 - smoothstep(innerRadius, outerRadius, distFromCenter);

          // Mask out the center (where icon will be)
          float centerMask = step(innerRadius, distFromCenter);

          vec3 shadowColor = vec3(0.0, 0.0, 0.0); // Black shadow
          float alpha = borderGradient * centerMask * 0.3; // Much lighter shadow

          gl_FragColor = vec4(shadowColor, alpha);
        }
      `,
      transparent: true,
      depthWrite: false, // Don't write to depth buffer for shadows
    });
  }, []);

  return (
    <group position={position}>
      {/* Shadow layer behind icon */}
      <mesh position={[0, 0, -0.001]}>
        <planeGeometry args={shadowDimensions} />
        <primitive object={shadowMaterial} attach="material" />
      </mesh>

      {/* Main icon */}
      <mesh>
        <planeGeometry args={dimensions} />
        <meshBasicMaterial
          transparent
          alphaTest={0.1}
          map={texture}
          // Darker filter applied
          toneMapped={false}
          color={new THREE.Color(0.7, 0.7, 0.7)} // Darken the texture
        />
      </mesh>
    </group>
  );
}
