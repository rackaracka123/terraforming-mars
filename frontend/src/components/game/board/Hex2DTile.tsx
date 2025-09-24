import { useRef, useState, useMemo } from "react";
import { useFrame } from "@react-three/fiber";
import { Text } from "@react-three/drei";
import * as THREE from "three";
import { HexTile2D } from "../../../utils/hex-grid-2d";

interface Hex2DTileProps {
  tileData: HexTile2D;
  tileType: "empty" | "ocean" | "greenery" | "city" | "special";
  ownerId?: string | null;
  onClick: () => void;
}

export default function Hex2DTile({
  tileData,
  tileType,
  ownerId,
  onClick,
}: Hex2DTileProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  const [hovered, setHovered] = useState(false);

  // Create 2D hexagon geometry (pointy-top)
  const hexGeometry = useMemo(() => {
    const geometry = new THREE.CircleGeometry(0.25, 6);
    // Rotate to pointy-top orientation (pointy up/down, flat left/right)
    geometry.rotateZ(Math.PI / 2);
    return geometry;
  }, []);

  // Create hex border
  const borderGeometry = useMemo(() => {
    const geometry = new THREE.RingGeometry(0.24, 0.25, 6);
    geometry.rotateZ(Math.PI / 2);
    return geometry;
  }, []);

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
          ? new THREE.Color("#0d47a1").multiplyScalar(0.7)
          : new THREE.Color("#6d4c41").multiplyScalar(0.6);
    }
  }, [tileType, tileData.isOceanSpace, hovered]);

  // Border color
  const borderColor = useMemo(() => {
    if (hovered) return new THREE.Color("#ffffff");
    return tileColor.clone().multiplyScalar(0.7);
  }, [tileColor, hovered]);

  // Bonus text
  const bonusText = useMemo(() => {
    const entries = Object.entries(tileData.bonuses);
    if (entries.length === 0) return "";

    return entries
      .map(([key, value]) => {
        const symbols: { [key: string]: string } = {
          steel: "⚙",
          titanium: "🛡",
          plants: "🌱",
          cards: "📋",
        };
        return `${symbols[key] || "💰"}${value}`;
      })
      .join(" ");
  }, [tileData.bonuses]);

  // Hover animation
  useFrame((state) => {
    if (meshRef.current && hovered) {
      const scale = 1 + Math.sin(state.clock.elapsedTime * 4) * 0.05;
      meshRef.current.scale.setScalar(scale);
    } else if (meshRef.current) {
      meshRef.current.scale.setScalar(1);
    }
  });

  return (
    <group position={[tileData.position.x, tileData.position.y, 0]}>
      {/* Main hex tile */}
      <mesh
        ref={meshRef}
        geometry={hexGeometry}
        onPointerEnter={() => setHovered(true)}
        onPointerLeave={() => setHovered(false)}
        onClick={onClick}
      >
        <meshStandardMaterial
          color={tileColor}
          transparent={tileType === "empty"}
          opacity={tileType === "empty" ? 0.6 : 1}
          roughness={0.7}
          metalness={0.1}
        />
      </mesh>

      {/* Hex border */}
      <mesh geometry={borderGeometry} position={[0, 0, 0.001]}>
        <meshBasicMaterial color={borderColor} transparent opacity={0.8} />
      </mesh>

      {/* Ocean space indicator */}
      {tileType === "empty" && tileData.isOceanSpace && (
        <mesh position={[0, 0, 0.001]}>
          <circleGeometry args={[0.1, 16]} />
          <meshBasicMaterial color="#1565c0" transparent opacity={0.6} />
        </mesh>
      )}

      {/* Tile type icon */}
      {tileType !== "empty" && (
        <Text
          position={[0, 0, 0.01]}
          fontSize={0.12}
          color="white"
          anchorX="center"
          anchorY="middle"
        >
          {tileType === "ocean"
            ? "🌊"
            : tileType === "greenery"
              ? "🌲"
              : tileType === "city"
                ? "🏙️"
                : tileType === "special"
                  ? "⭐"
                  : ""}
        </Text>
      )}

      {/* Bonus resources */}
      {bonusText && (
        <Text
          position={[0, -0.15, 0.01]}
          fontSize={0.08}
          color="#ffd54f"
          anchorX="center"
          anchorY="middle"
        >
          {bonusText}
        </Text>
      )}

      {/* Owner indicator */}
      {ownerId && (
        <mesh position={[0.15, 0.15, 0.01]}>
          <circleGeometry args={[0.03, 16]} />
          <meshBasicMaterial
            color={`hsl(${(ownerId.charCodeAt(0) * 137.5) % 360}, 70%, 50%)`}
          />
        </mesh>
      )}
    </group>
  );
}
