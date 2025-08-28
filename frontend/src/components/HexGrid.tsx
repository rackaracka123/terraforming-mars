import React, { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';

// Mars board hex data (matches backend layout)
const MARS_HEXES = [
  // Row 0 (top)
  { q: 4, r: -4, s: 0, bonuses: { steel: 2 }, isOceanSpace: false, isRestricted: false },
  
  // Row 1
  { q: 3, r: -3, s: 0, bonuses: {}, isOceanSpace: true, isRestricted: false },
  { q: 4, r: -3, s: -1, bonuses: { steel: 2 }, isOceanSpace: false, isRestricted: false },
  { q: 5, r: -3, s: -2, bonuses: {}, isOceanSpace: true, isRestricted: false },
  
  // Row 2
  { q: 2, r: -2, s: 0, bonuses: { cards: 1 }, isOceanSpace: false, isRestricted: false },
  { q: 3, r: -2, s: -1, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { q: 4, r: -2, s: -2, bonuses: { steel: 1 }, isOceanSpace: false, isRestricted: false },
  { q: 5, r: -2, s: -3, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { q: 6, r: -2, s: -4, bonuses: { titanium: 1 }, isOceanSpace: false, isRestricted: false },
  
  // Row 3
  { q: 1, r: -1, s: 0, bonuses: { plants: 1 }, isOceanSpace: false, isRestricted: false },
  { q: 2, r: -1, s: -1, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { q: 3, r: -1, s: -2, bonuses: { steel: 1 }, isOceanSpace: false, isRestricted: false },
  { q: 4, r: -1, s: -3, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { q: 5, r: -1, s: -4, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { q: 6, r: -1, s: -5, bonuses: { titanium: 1 }, isOceanSpace: false, isRestricted: false },
  { q: 7, r: -1, s: -6, bonuses: {}, isOceanSpace: true, isRestricted: false },
  
  // Row 4 (middle)
  { q: 0, r: 0, s: 0, bonuses: { plants: 1 }, isOceanSpace: false, isRestricted: false },
  { q: 1, r: 0, s: -1, bonuses: { plants: 1 }, isOceanSpace: false, isRestricted: false },
  { q: 2, r: 0, s: -2, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { q: 3, r: 0, s: -3, bonuses: {}, isOceanSpace: true, isRestricted: false },
  { q: 4, r: 0, s: -4, bonuses: {}, isOceanSpace: false, isRestricted: true }, // Tharsis Tholus
  { q: 5, r: 0, s: -5, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { q: 6, r: 0, s: -6, bonuses: {}, isOceanSpace: true, isRestricted: false },
  { q: 7, r: 0, s: -7, bonuses: { titanium: 2 }, isOceanSpace: false, isRestricted: false },
  { q: 8, r: 0, s: -8, bonuses: {}, isOceanSpace: true, isRestricted: false },
  
  // Row 5
  { q: 0, r: 1, s: -1, bonuses: { plants: 2 }, isOceanSpace: false, isRestricted: false },
  { q: 1, r: 1, s: -2, bonuses: { plants: 1 }, isOceanSpace: false, isRestricted: false },
  { q: 2, r: 1, s: -3, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { q: 3, r: 1, s: -4, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { q: 4, r: 1, s: -5, bonuses: { steel: 1 }, isOceanSpace: false, isRestricted: false },
  { q: 5, r: 1, s: -6, bonuses: {}, isOceanSpace: true, isRestricted: false },
  { q: 6, r: 1, s: -7, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { q: 7, r: 1, s: -8, bonuses: { titanium: 1 }, isOceanSpace: false, isRestricted: false },
  
  // Adding more hexes for a fuller board representation
  // Row 6
  { q: 1, r: 2, s: -3, bonuses: { plants: 1 }, isOceanSpace: false, isRestricted: false },
  { q: 2, r: 2, s: -4, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { q: 3, r: 2, s: -5, bonuses: {}, isOceanSpace: true, isRestricted: false },
  { q: 4, r: 2, s: -6, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { q: 5, r: 2, s: -7, bonuses: { steel: 1 }, isOceanSpace: false, isRestricted: false },
  { q: 6, r: 2, s: -8, bonuses: {}, isOceanSpace: false, isRestricted: false },
  
  // Row 7
  { q: 2, r: 3, s: -5, bonuses: { cards: 1 }, isOceanSpace: false, isRestricted: false },
  { q: 3, r: 3, s: -6, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { q: 4, r: 3, s: -7, bonuses: {}, isOceanSpace: true, isRestricted: false },
  { q: 5, r: 3, s: -8, bonuses: { steel: 2 }, isOceanSpace: false, isRestricted: false },
  
  // Row 8 (bottom)
  { q: 4, r: 4, s: -8, bonuses: { steel: 2 }, isOceanSpace: false, isRestricted: false },
];

// Utility functions
function hexToPixel(q: number, r: number, size: number): { x: number, z: number } {
  const x = size * (Math.sqrt(3) * q + Math.sqrt(3) / 2 * r);
  const z = size * (3 / 2 * r);
  return { x, z };
}

function getHexColor(hex: any, gameState: any): string {
  // Color based on tile type and bonuses
  if (hex.isOceanSpace) {
    // Check if ocean is placed here
    return '#4a90e2'; // Ocean blue
  }
  
  if (hex.isRestricted) {
    return '#8b4513'; // Brown for restricted areas
  }
  
  // Color based on bonuses
  if (hex.bonuses.steel) {
    return '#c0c0c0'; // Silver for steel
  }
  if (hex.bonuses.titanium) {
    return '#ffd700'; // Gold for titanium
  }
  if (hex.bonuses.plants) {
    return '#32cd32'; // Green for plants
  }
  if (hex.bonuses.cards) {
    return '#9370db'; // Purple for cards
  }
  
  return '#cd853f'; // Default Mars tan
}

interface HexTileProps {
  hex: any;
  gameState?: any;
  onClick?: (hex: any) => void;
}

function HexTile({ hex, gameState, onClick }: HexTileProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  const position = hexToPixel(hex.q, hex.r, 0.35);
  
  const hexGeometry = useMemo(() => {
    const geometry = new THREE.CylinderGeometry(0.3, 0.3, 0.05, 6);
    return geometry;
  }, []);
  
  const handleClick = () => {
    if (onClick) {
      onClick(hex);
    }
  };
  
  return (
    <mesh
      ref={meshRef}
      position={[position.x, 0.03, position.z]}
      geometry={hexGeometry}
      onClick={handleClick}
    >
      <meshStandardMaterial 
        color={getHexColor(hex, gameState)}
        roughness={0.7}
        metalness={0.1}
        transparent
        opacity={hex.isOceanSpace ? 0.8 : 0.9}
      />
      
      {/* Hex border */}
      <lineSegments position={[0, 0.026, 0]}>
        <edgesGeometry args={[hexGeometry]} />
        <lineBasicMaterial color="#333333" linewidth={1} />
      </lineSegments>
    </mesh>
  );
}

interface HexGridProps {
  gameState?: any;
  onHexClick?: (hex: any) => void;
}

export default function HexGrid({ gameState, onHexClick }: HexGridProps) {
  const groupRef = useRef<THREE.Group>(null);
  
  // Subtle animation for the hex grid
  useFrame((state) => {
    if (groupRef.current) {
      groupRef.current.rotation.y = Math.sin(state.clock.elapsedTime * 0.1) * 0.02;
    }
  });
  
  return (
    <group ref={groupRef}>
      {MARS_HEXES.map((hex, index) => (
        <HexTile
          key={`${hex.q}-${hex.r}`}
          hex={hex}
          gameState={gameState}
          onClick={onHexClick}
        />
      ))}
      
      {/* Grid labels for development (can be removed later) */}
      {MARS_HEXES.slice(0, 5).map((hex, index) => {
        const position = hexToPixel(hex.q, hex.r, 0.35);
        return (
          <mesh key={`label-${hex.q}-${hex.r}`} position={[position.x, 0.08, position.z]}>
            <planeGeometry args={[0.2, 0.1]} />
            <meshBasicMaterial color="#ffffff" transparent opacity={0.8} />
          </mesh>
        );
      })}
    </group>
  );
}