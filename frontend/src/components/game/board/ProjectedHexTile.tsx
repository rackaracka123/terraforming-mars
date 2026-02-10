import { useRef, useState, useMemo, useEffect } from "react";
import { useFrame, useLoader, useThree } from "@react-three/fiber";
import { Text, useGLTF, useTexture } from "@react-three/drei";
import * as THREE from "three";
import { HexTile2D } from "../../../utils/hex-grid-2d";
import { SkeletonUtils } from "three-stdlib";
import { panState } from "../controls/PanControls";
import CityEmergenceEffect from "./effects/CityEmergenceEffect";

// Preload 3D models and textures for better performance
useGLTF.preload("/assets/models/city.glb");
useGLTF.preload("/assets/models/forrest.glb");

// Ocean water vertex shader - passes world position and computes tangent basis
const OCEAN_VERTEX_SHADER = `
  varying vec4 worldPosition;
  varying vec3 vNormal;
  varying vec3 vTangent;
  varying vec3 vBitangent;
  varying vec2 vUv;

  void main() {
    vUv = uv;
    worldPosition = modelMatrix * vec4(position, 1.0);

    // Transform normal to world space
    vNormal = normalize(mat3(modelMatrix) * normal);

    // Build tangent basis for spherical surface
    vec3 up = abs(vNormal.y) < 0.999 ? vec3(0.0, 1.0, 0.0) : vec3(1.0, 0.0, 0.0);
    vTangent = normalize(cross(up, vNormal));
    vBitangent = cross(vNormal, vTangent);

    gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
  }
`;

// Ocean water fragment shader - exact Three.js Water algorithm
// Adapted for spherical hex tiles with tangent-space normal perturbation
const OCEAN_FRAGMENT_SHADER = `
  uniform float time;
  uniform float size;
  uniform float distortionScale;
  uniform float alpha;
  uniform sampler2D normalSampler;
  uniform vec3 sunColor;
  uniform vec3 sunDirection;
  uniform vec3 eye;
  uniform vec3 waterColor;

  varying vec4 worldPosition;
  varying vec3 vNormal;
  varying vec3 vTangent;
  varying vec3 vBitangent;
  varying vec2 vUv;

  // Exact getNoise from Three.js Water.js
  vec4 getNoise(vec2 uv) {
    vec2 uv0 = (uv / 103.0) + vec2(time / 17.0, time / 29.0);
    vec2 uv1 = uv / 107.0 - vec2(time / -19.0, time / 31.0);
    vec2 uv2 = uv / vec2(8907.0, 9803.0) + vec2(time / 101.0, time / 97.0);
    vec2 uv3 = uv / vec2(1091.0, 1027.0) - vec2(time / 109.0, time / -113.0);
    vec4 noise = texture2D(normalSampler, uv0) +
      texture2D(normalSampler, uv1) +
      texture2D(normalSampler, uv2) +
      texture2D(normalSampler, uv3);
    return noise * 0.5 - 1.0;
  }

  // Exact sunLight from Three.js Water.js
  void sunLight(const vec3 surfaceNormal, const vec3 eyeDirection, float shiny, float spec, float diffuse, inout vec3 diffuseColor, inout vec3 specularColor) {
    vec3 reflection = normalize(reflect(-sunDirection, surfaceNormal));
    float direction = max(0.0, dot(eyeDirection, reflection));
    specularColor += pow(direction, shiny) * sunColor * spec;
    diffuseColor += max(dot(sunDirection, surfaceNormal), 0.0) * sunColor * diffuse;
  }

  void main() {
    // Use world position for UV sampling like Three.js Water
    // Project onto tangent plane for consistent wave direction on sphere
    vec2 projectedPos = vec2(
      dot(worldPosition.xyz, vTangent),
      dot(worldPosition.xyz, vBitangent)
    );
    vec4 noise = getNoise(projectedPos * size);

    // Build surface normal in tangent space then transform to world space
    // noise.xzy with scaling is exactly from Three.js Water
    vec3 tangentNormal = normalize(noise.xzy * vec3(1.5, 1.0, 1.5));

    // Transform tangent-space normal to world space
    vec3 surfaceNormal = normalize(
      vTangent * tangentNormal.x +
      vNormal * tangentNormal.y +
      vBitangent * tangentNormal.z
    );

    vec3 diffuseLight = vec3(0.0);
    vec3 specularLight = vec3(0.0);

    vec3 worldToEye = eye - worldPosition.xyz;
    vec3 eyeDirection = normalize(worldToEye);

    // Sun lighting - exact Three.js Water parameters
    sunLight(surfaceNormal, eyeDirection, 100.0, 2.0, 0.5, diffuseLight, specularLight);

    float distance = length(worldToEye);

    // Fresnel reflectance - exact Three.js Water formula
    float theta = max(dot(eyeDirection, surfaceNormal), 0.0);
    float rf0 = 0.3;
    float reflectance = rf0 + (1.0 - rf0) * pow((1.0 - theta), 5.0);

    // Scatter color - exact Three.js Water formula
    vec3 scatter = max(0.0, dot(surfaceNormal, eyeDirection)) * waterColor;

    // Simulated sky reflection (since we don't have mirror texture)
    vec3 skyColor = vec3(0.4, 0.5, 0.7);
    vec3 horizonColor = vec3(0.7, 0.75, 0.85);
    float skyGradient = max(0.0, surfaceNormal.y);
    vec3 reflectionSample = mix(horizonColor, skyColor, skyGradient);

    // Final color blend - based on Three.js Water albedo calculation
    vec3 albedo = mix(
      (sunColor * diffuseLight * 0.3 + scatter),
      (vec3(0.1) + reflectionSample * 0.9 + reflectionSample * specularLight),
      reflectance
    );

    // Soft hex edge fade
    vec2 center = vUv - 0.5;
    float distFromCenter = length(center);
    float edgeAlpha = 1.0 - smoothstep(0.35, 0.5, distFromCenter);

    gl_FragColor = vec4(albedo, alpha * edgeAlpha);
  }
`;

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
  /** Whether this tile was just placed (triggers emergence animation for cities) */
  isNewlyPlaced?: boolean;
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
  isNewlyPlaced = false,
}: ProjectedHexTileProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  const vpTextRef = useRef<THREE.Group>(null);
  const [hovered, setHovered] = useState(false);
  const animationStartTimeRef = useRef<number | null>(null);

  // Access camera for eye position in water shader
  const { camera } = useThree();

  // Load water normals texture for three.js Water algorithm
  const waterNormals = useTexture("/assets/textures/waternormals.jpg");
  waterNormals.wrapS = waterNormals.wrapT = THREE.RepeatWrapping;

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

  // Create animated ocean water material - exact Three.js Water uniforms
  const oceanWaterMaterial = useMemo(() => {
    return new THREE.ShaderMaterial({
      vertexShader: OCEAN_VERTEX_SHADER,
      fragmentShader: OCEAN_FRAGMENT_SHADER,
      uniforms: {
        time: { value: 0.0 },
        size: { value: 250.0 },
        distortionScale: { value: 3.7 },
        alpha: { value: 1.0 },
        normalSampler: { value: waterNormals },
        sunColor: { value: new THREE.Vector3(1.0, 1.0, 1.0) },
        sunDirection: { value: new THREE.Vector3(0.70707, 0.70707, 0.0).normalize() },
        eye: { value: new THREE.Vector3() },
        waterColor: { value: new THREE.Vector3(0.0, 0.118, 0.059) },
      },
      transparent: true,
      side: THREE.DoubleSide,
      depthWrite: false,
    });
  }, [waterNormals]);

  // Create circular geometry for the animated ocean water - sized to fill hex
  const oceanWaterGeometry = useMemo(() => {
    // Match hex size (0.166) with higher segment count for smooth edges
    const geometry = new THREE.CircleGeometry(0.166, 32);
    return geometry;
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

    // Animate ocean water shader - use elapsed time for consistent animation
    if (oceanWaterMaterial.uniforms) {
      oceanWaterMaterial.uniforms.time.value = state.clock.elapsedTime * 0.8;
      oceanWaterMaterial.uniforms.eye.value.copy(camera.position);
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
          opacity={tileType === "ocean" ? 0 : tileType === "empty" ? 0.3 : 0.7}
          roughness={0.7}
          metalness={0.1}
          side={THREE.DoubleSide}
        />
      </mesh>

      {/* Hex border - hidden for ocean tiles */}
      {tileType !== "ocean" && (
        <mesh geometry={borderGeometry} position={[0, 0, 0.001]}>
          <meshBasicMaterial color={borderColor} transparent opacity={0.9} />
        </mesh>
      )}

      {/* Ocean space indicator - blue gradient fading to center */}
      {tileType === "empty" && tileData.isOceanSpace && (
        <mesh
          position={[0, 0, 0.002]}
          geometry={oceanGradientGeometry}
          material={oceanBorderMaterial}
        />
      )}

      {/* Hover glow effect with pulsation - hidden for ocean tiles */}
      {tileType !== "ocean" && (
        <mesh
          position={[0, 0, 0.0015]}
          geometry={oceanGradientGeometry}
          material={hoverGlowMaterial}
        />
      )}

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

      {/* Animated ocean water effect using three.js Water algorithm */}
      {tileType === "ocean" && (
        <mesh
          position={[0, 0, 0.004]}
          geometry={oceanWaterGeometry}
          material={oceanWaterMaterial}
        />
      )}

      {/* Tile type 3D model (city, greenery) */}
      {(tileType === "city" || tileType === "greenery") && (
        <TileModel
          tileType={tileType}
          position={[0, 0, 0.03]}
          isNewlyPlaced={isNewlyPlaced}
          surfaceNormal={tileData.normal}
          worldPosition={adjustedPosition}
        />
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

// Component for rendering 3D tile models (city, greenery)
interface TileModelProps {
  tileType: "city" | "greenery";
  position: [number, number, number];
  isNewlyPlaced?: boolean;
  surfaceNormal?: THREE.Vector3;
  worldPosition?: THREE.Vector3;
  onEmergenceComplete?: () => void;
}

function TileModel({
  tileType,
  position,
  isNewlyPlaced = false,
  surfaceNormal,
  worldPosition,
  onEmergenceComplete,
}: TileModelProps) {
  const groupRef = useRef<THREE.Group>(null);
  const emergenceStartRef = useRef<number | null>(null);
  const [isEmerging, setIsEmerging] = useState(isNewlyPlaced && tileType === "city");
  const [showParticles, setShowParticles] = useState(isNewlyPlaced && tileType === "city");

  useEffect(() => {
    if (isNewlyPlaced && tileType === "city") {
      setIsEmerging(true);
      setShowParticles(true);
      emergenceStartRef.current = null;
    }
  }, [isNewlyPlaced, tileType]);

  useFrame((state) => {
    if (!isEmerging || !groupRef.current) return;

    if (emergenceStartRef.current === null) {
      emergenceStartRef.current = state.clock.elapsedTime * 1000;
    }

    const elapsed = state.clock.elapsedTime * 1000 - emergenceStartRef.current;
    const riseDuration = 800;
    const riseDepth = 0.08;

    const riseProgress = Math.min(elapsed / riseDuration, 1);
    const easedProgress = 1 - Math.pow(1 - riseProgress, 3);

    const zOffset = -riseDepth * (1 - easedProgress);

    const shakeIntensity = 0.01 * (1 - easedProgress);
    const shakeX = (Math.random() - 0.5) * 2 * shakeIntensity;
    const shakeY = (Math.random() - 0.5) * 2 * shakeIntensity;

    groupRef.current.position.set(
      position[0] + shakeX,
      position[1] + shakeY,
      position[2] + zOffset,
    );

    if (riseProgress >= 1) {
      setIsEmerging(false);
      groupRef.current.position.set(position[0], position[1], position[2]);
      onEmergenceComplete?.();
    }
  });

  const handleParticleComplete = () => {
    setShowParticles(false);
  };
  // Load appropriate model based on tile type
  const modelPath = tileType === "city" ? "/assets/models/city.glb" : "/assets/models/forrest.glb";

  const { scene } = useGLTF(modelPath);

  // Clone and configure the model with proper transformations
  const configuredModel = useMemo(() => {
    // Clone the scene properly
    const clonedScene = SkeletonUtils.clone(scene);

    // Calculate bounding box and scale - larger size for greenery to fill hex
    const targetSize = tileType === "greenery" ? 0.3 : 0.28;
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

    // Enable shadows
    clonedScene.traverse((child) => {
      if (child instanceof THREE.Mesh) {
        child.castShadow = true;
        child.receiveShadow = true;
      }
    });

    return clonedScene;
  }, [scene, tileType]);

  // Rotation for specific tile types - both city and greenery need +90° rotation
  const rotation: [number, number, number] = useMemo(() => {
    return [Math.PI / 2, 0, 0];
  }, []);

  return (
    <>
      <group ref={groupRef} position={position} rotation={rotation}>
        <primitive object={configuredModel} />
      </group>
      {showParticles && surfaceNormal && worldPosition && (
        <CityEmergenceEffect
          position={worldPosition}
          normal={surfaceNormal}
          duration={2600}
          onComplete={handleParticleComplete}
        />
      )}
    </>
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
