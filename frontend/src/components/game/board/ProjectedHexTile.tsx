import { useRef, useState, useMemo, useEffect } from "react";
import { useFrame, useLoader } from "@react-three/fiber";
import { Text, useGLTF } from "@react-three/drei";
import * as THREE from "three";
import { HexTile2D } from "../../../utils/hex-grid-2d";
import { SkeletonUtils } from "three-stdlib";
import { panState } from "../controls/PanControls";

// Preload 3D models for better performance
useGLTF.preload("/assets/models/city.glb");
useGLTF.preload("/assets/models/forrest.glb");
useGLTF.preload("/assets/models/water.glb");

interface ProjectedHexTileData extends HexTile2D {
  spherePosition: THREE.Vector3;
  normal: THREE.Vector3;
}

export type TileHighlightMode = "greenery" | "city" | "adjacent" | null;

interface ProjectedHexTileProps {
  tileData: ProjectedHexTileData;
  tileType: "empty" | "ocean" | "greenery" | "city" | "special";
  ownerId?: string | null;
  reservedById?: string | null;
  displayName?: string;
  onClick: () => void;
  isAvailableForPlacement?: boolean;
  /** Highlight mode for end game VP counting animation */
  highlightMode?: TileHighlightMode;
  /** VP amount to display as floating text */
  vpAmount?: number;
  /** Whether the VP indicator should animate (float up) */
  vpAnimating?: boolean;
  /** Whether to animate entrance (scale from 0 to 1) */
  animateEntrance?: boolean;
  /** Delay in ms before starting entrance animation */
  entranceDelay?: number;
}

export default function ProjectedHexTile({
  tileData,
  tileType,
  ownerId,
  reservedById,
  displayName,
  onClick,
  isAvailableForPlacement = false,
  highlightMode = null,
  vpAmount,
  vpAnimating = false,
  animateEntrance = false,
  entranceDelay = 0,
}: ProjectedHexTileProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  const vpTextRef = useRef<THREE.Group>(null);
  const [hovered, setHovered] = useState(false);
  const animationStartTimeRef = useRef<number | null>(null);

  // Entrance animation state
  const [entranceScale, setEntranceScale] = useState(animateEntrance ? 0 : 1);
  const entranceStartRef = useRef<number | null>(null);
  const entranceDoneRef = useRef(!animateEntrance);

  useEffect(() => {
    if (animateEntrance && entranceDoneRef.current) {
      // Reset for a new entrance animation
      setEntranceScale(0);
      entranceStartRef.current = null;
      entranceDoneRef.current = false;
    }
  }, [animateEntrance]);

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

  // Create end game highlight material for VP counting animation
  const endGameHighlightMaterial = useMemo(() => {
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
        uniform vec3 highlightColor;
        uniform float opacity;
        varying vec2 vUv;

        void main() {
          // Convert UV to centered coordinates (-0.5 to 0.5)
          vec2 center = vUv - 0.5;

          // Calculate distance from center (0.0 at center, ~0.5 at edges)
          float distFromCenter = length(center);

          // Create gradient that's strong at edges and fades toward center
          float gradient = smoothstep(0.1, 0.45, distFromCenter);

          // Pulsing animation - slower and more pronounced
          float pulse = 0.6 + 0.4 * sin(time * 3.0);

          float alpha = gradient * pulse * 0.7 * opacity;

          gl_FragColor = vec4(highlightColor, alpha);
        }
      `,
      uniforms: {
        time: { value: 0.0 },
        highlightColor: { value: new THREE.Vector3(0.13, 0.77, 0.27) }, // Default green
        opacity: { value: 0.0 },
      },
      transparent: true,
      side: THREE.DoubleSide,
      blending: THREE.AdditiveBlending,
    });
  }, []);

  // Update shader time uniforms and handle hover/entrance animations
  useFrame((state) => {
    // Entrance animation: staggered scale from 0 to 1
    if (animateEntrance && !entranceDoneRef.current) {
      if (entranceStartRef.current === null) {
        entranceStartRef.current = state.clock.elapsedTime;
      }
      const elapsed = (state.clock.elapsedTime - entranceStartRef.current) * 1000;
      if (elapsed >= entranceDelay) {
        const animDuration = 400;
        const t = Math.min((elapsed - entranceDelay) / animDuration, 1);
        // Ease-out cubic
        const eased = 1 - Math.pow(1 - t, 3);
        setEntranceScale(eased);
        if (t >= 1) {
          entranceDoneRef.current = true;
        }
      }
    }
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

    if (endGameHighlightMaterial.uniforms) {
      endGameHighlightMaterial.uniforms.time.value = state.clock.elapsedTime;

      // Animate opacity based on highlight mode (fade in/out smoothly)
      const targetOpacity = highlightMode ? 1.0 : 0.0;
      endGameHighlightMaterial.uniforms.opacity.value = THREE.MathUtils.lerp(
        endGameHighlightMaterial.uniforms.opacity.value,
        targetOpacity,
        0.1, // Smooth fade speed
      );

      // Update highlight color based on mode
      if (highlightMode) {
        let color: THREE.Vector3;
        switch (highlightMode) {
          case "greenery":
            color = new THREE.Vector3(0.13, 0.77, 0.27); // Green
            break;
          case "city":
            color = new THREE.Vector3(0.58, 0.64, 0.7); // Gray-blue
            break;
          case "adjacent":
            color = new THREE.Vector3(1.0, 0.84, 0.0); // Gold
            break;
          default:
            color = new THREE.Vector3(0.13, 0.77, 0.27);
        }
        endGameHighlightMaterial.uniforms.highlightColor.value = color;
      }
    }

    // VP text floating animation
    if (vpTextRef.current && vpAmount !== undefined && vpAnimating) {
      // Start animation timer
      if (animationStartTimeRef.current === null) {
        animationStartTimeRef.current = state.clock.elapsedTime;
      }

      const elapsed = state.clock.elapsedTime - animationStartTimeRef.current;
      const duration = 2.0; // 2 second animation

      if (elapsed < 0.3) {
        // Enter phase: scale up, rise
        const progress = elapsed / 0.3;
        vpTextRef.current.scale.setScalar(progress);
        vpTextRef.current.position.z = 0.02 + progress * 0.05;
      } else if (elapsed < 1.8) {
        // Idle phase: gentle float
        vpTextRef.current.scale.setScalar(1);
        vpTextRef.current.position.z = 0.07 + Math.sin(elapsed * 2) * 0.01;
      } else if (elapsed < duration) {
        // Exit phase: fade out, shrink
        const progress = 1 - (elapsed - 1.8) / 0.2;
        vpTextRef.current.scale.setScalar(Math.max(0, progress));
      } else {
        // Animation complete - hide
        vpTextRef.current.scale.setScalar(0);
      }
    } else if (vpTextRef.current && !vpAnimating) {
      // Reset animation when not animating
      animationStartTimeRef.current = null;
      vpTextRef.current.scale.setScalar(vpAmount !== undefined ? 1 : 0);
      vpTextRef.current.position.z = 0.07;
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
    return tileData.spherePosition.clone().add(tileData.normal.clone().multiplyScalar(0.01));
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

  // Bonus icon group structure for proper spacing
  interface BonusIconGroup {
    type: string;
    iconPath: string;
    count: number;
    isCredits: boolean;
  }

  // Bonus icons data - grouped by resource type
  const bonusIconGroups = useMemo((): BonusIconGroup[] => {
    const entries = Object.entries(tileData.bonuses);
    if (entries.length === 0) return [];

    const iconPaths: { [key: string]: string } = {
      steel: "/assets/resources/steel.png",
      titanium: "/assets/resources/titanium.png",
      plants: "/assets/resources/plant.png",
      plant: "/assets/resources/plant.png",
      cards: "/assets/resources/card.png",
      "card-draw": "/assets/resources/card.png",
      credit: "/assets/resources/megacredit.png",
    };

    return entries.map(([key, value]) => ({
      type: key,
      iconPath: iconPaths[key] || "/assets/resources/megacredit.png",
      count: value,
      isCredits: key === "credit",
    }));
  }, [tileData.bonuses]);

  // Calculate icon positions with proper spacing between same-type and different-type icons
  const calculateIconPositions = (groups: BonusIconGroup[]) => {
    const ICON_GAP = 0.005; // Gap between same-type icons
    const GROUP_GAP = 0.01; // Gap between different resource groups
    const ICON_SIZE = 0.05;

    const positions: { x: number; group: BonusIconGroup; indexInGroup: number }[] = [];

    // Calculate total width (credits show 1 icon with count as text)
    let totalWidth = 0;
    groups.forEach((group, groupIndex) => {
      if (groupIndex > 0) totalWidth += GROUP_GAP;
      const iconCount = group.isCredits ? 1 : group.count;
      totalWidth += iconCount * ICON_SIZE + Math.max(0, iconCount - 1) * ICON_GAP;
    });

    // Calculate centered positions
    let currentX = -totalWidth / 2;
    groups.forEach((group, groupIndex) => {
      if (groupIndex > 0) currentX += GROUP_GAP;

      const iconCount = group.isCredits ? 1 : group.count;
      for (let i = 0; i < iconCount; i++) {
        if (i > 0) currentX += ICON_GAP;
        positions.push({ x: currentX + ICON_SIZE / 2, group, indexInGroup: i });
        currentX += ICON_SIZE;
      }
    });

    return positions;
  };

  return (
    <group
      position={adjustedPosition}
      quaternion={surfaceQuaternion}
      scale={[entranceScale, entranceScale, entranceScale]}
    >
      {/* Main hex tile */}
      <mesh
        ref={meshRef}
        geometry={hexGeometry}
        onPointerEnter={() => {
          if (!panState.isPanning) setHovered(true);
        }}
        onPointerLeave={() => {
          setHovered(false);
        }}
        onClick={(event) => {
          if (panState.isPanning) return;
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

      {/* End game VP counting highlight effect - always rendered, opacity controlled via shader */}
      <mesh
        position={[0, 0, 0.003]}
        geometry={oceanGradientGeometry}
        material={endGameHighlightMaterial}
      />

      {/* Tile type 3D model (city, greenery, ocean) */}
      {(tileType === "city" || tileType === "greenery" || tileType === "ocean") && (
        <TileModel tileType={tileType} position={[0, 0, tileType === "ocean" ? 0.00001 : 0.03]} />
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
      {tileType !== "greenery" && (
        <>
          {displayName && bonusIconGroups.length > 0 ? (
            // Two-row layout when both displayName and bonuses exist
            <>
              {/* Display name in upper row */}
              <Text
                position={[0, 0.03, 0.01]}
                fontSize={0.035}
                font="/assets/Prototype.ttf"
                color="white"
                outlineWidth={0.003}
                outlineColor="black"
                anchorX="center"
                anchorY="middle"
                textAlign="center"
                maxWidth={0.25}
              >
                {displayName}
              </Text>

              {/* Bonus icons in lower row */}
              {(() => {
                const positions = calculateIconPositions(bonusIconGroups);
                return positions.map((pos) => (
                  <BonusIcon
                    key={`${pos.group.type}-${pos.indexInGroup}`}
                    iconPath={pos.group.iconPath}
                    position={[pos.x, -0.03, 0.01]}
                    isCredits={pos.group.isCredits}
                    creditAmount={pos.group.isCredits ? pos.group.count : undefined}
                    lightIntensity={entranceScale}
                  />
                ));
              })()}
            </>
          ) : displayName ? (
            // Display name centered when no bonuses
            <Text
              position={[0, 0, 0.01]}
              fontSize={0.04}
              font="/assets/Prototype.ttf"
              color="white"
              outlineWidth={0.003}
              outlineColor="black"
              anchorX="center"
              anchorY="middle"
              textAlign="center"
              maxWidth={0.25}
            >
              {displayName}
            </Text>
          ) : (
            // Only bonus icons when no display name
            (() => {
              const positions = calculateIconPositions(bonusIconGroups);
              return positions.map((pos) => (
                <BonusIcon
                  key={`${pos.group.type}-${pos.indexInGroup}`}
                  iconPath={pos.group.iconPath}
                  position={[pos.x, 0, 0.01]}
                  isCredits={pos.group.isCredits}
                  creditAmount={pos.group.isCredits ? pos.group.count : undefined}
                  lightIntensity={entranceScale}
                />
              ));
            })()
          )}
        </>
      )}

      {/* Owner indicator */}
      {ownerId && (
        <mesh position={[0.1, 0.1, 0.01]}>
          <circleGeometry args={[0.02, 16]} />
          <meshBasicMaterial color={`hsl(${(ownerId.charCodeAt(0) * 137.5) % 360}, 70%, 50%)`} />
        </mesh>
      )}

      {/* Reserved tile marker (land claim) - small flag icon at corner */}
      {reservedById && !ownerId && (
        <group position={[0, 0, 0.01]}>
          {/* Flag pole */}
          <mesh position={[0.08, 0.05, 0]}>
            <boxGeometry args={[0.004, 0.06, 0.004]} />
            <meshBasicMaterial color="#333333" />
          </mesh>
          {/* Flag (triangular marker) */}
          <mesh position={[0.1, 0.07, 0]}>
            <circleGeometry args={[0.025, 3]} />
            <meshBasicMaterial
              color={`hsl(${(reservedById.charCodeAt(0) * 137.5) % 360}, 70%, 50%)`}
            />
          </mesh>
        </group>
      )}

      {/* Floating VP indicator text */}
      {vpAmount !== undefined && (
        <group ref={vpTextRef} position={[0, 0, 0.07]}>
          <Text
            fontSize={0.08}
            color="#FFD700"
            anchorX="center"
            anchorY="middle"
            fontWeight="bold"
            outlineWidth={0.005}
            outlineColor="#000000"
          >
            +{vpAmount}
          </Text>
        </group>
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

    // Calculate bounding box and scale - larger size for greenery and ocean to fill hex
    const targetSize = tileType === "greenery" ? 0.3 : tileType === "ocean" ? 0.33 : 0.28;
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
      return [Math.PI / 2, 0, 0]; // +90° rotation on x-axis to flip city buildings right-side up
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
  isCredits?: boolean;
  creditAmount?: number;
  lightIntensity?: number;
}

function BonusIcon({
  iconPath,
  position,
  isCredits,
  creditAmount,
  lightIntensity = 1,
}: BonusIconProps) {
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

  // Determine light color based on resource type
  const lightColor = useMemo(() => {
    if (iconPath.includes("steel")) return "#8b7355";
    if (iconPath.includes("titanium")) return "#c9a227";
    if (iconPath.includes("plant")) return "#4a7c3f";
    if (iconPath.includes("megacredit")) return "#c9a227";
    if (iconPath.includes("card")) return "#6699cc";
    return "#a89070";
  }, [iconPath]);

  return (
    <group position={position}>
      {/* Ambient light emission matching resource color */}
      <pointLight
        position={[0, 0, 0.015]}
        intensity={0.07 * lightIntensity}
        distance={0.12}
        color={lightColor}
        decay={2}
      />

      {/* Main icon */}
      <mesh>
        <planeGeometry args={dimensions} />
        <meshBasicMaterial
          transparent
          alphaTest={0.1}
          map={texture}
          toneMapped={false}
          color={new THREE.Color(0.7, 0.7, 0.7)}
        />
      </mesh>

      {/* Credit amount overlay */}
      {isCredits && creditAmount !== undefined && (
        <Text
          position={[0, 0, 0.002]}
          fontSize={0.025}
          font="/assets/Prototype.ttf"
          color="black"
          anchorX="center"
          anchorY="middle"
        >
          {creditAmount}
        </Text>
      )}
    </group>
  );
}
