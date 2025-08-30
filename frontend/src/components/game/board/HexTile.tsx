import React, { useRef, useState, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import { Text } from '@react-three/drei';
import * as THREE from 'three';
import { HexCoordinate } from '../../../utils/geodesic';

interface HexTileProps {
  position: [number, number, number];
  coordinate: HexCoordinate;
  tileType: 'empty' | 'ocean' | 'greenery' | 'city' | 'special';
  isOceanSpace: boolean;
  bonuses: { [key: string]: number };
  ownerId?: string | null;
  specialType?: string | null;
  onClick: () => void;
}

export default function HexTile({
  position,
  coordinate,
  tileType,
  isOceanSpace,
  bonuses,
  ownerId,
  specialType,
  onClick
}: HexTileProps) {
  const meshRef = useRef<THREE.Mesh>(null);
  const groupRef = useRef<THREE.Group>(null);
  const [hovered, setHovered] = useState(false);
  

  // Create proper hexagonal geometry with 90-degree rotation - larger size
  const hexGeometry = useMemo(() => {
    const geometry = new THREE.CircleGeometry(0.18, 6);
    // Rotate 90 degrees (π/2 radians) - 45 more degrees
    geometry.rotateZ(Math.PI / 2);
    return geometry;
  }, []);
  
  // Create hexagon border/outline with matching 90-degree rotation - larger size
  const hexBorderGeometry = useMemo(() => {
    const geometry = new THREE.RingGeometry(0.175, 0.18, 6);
    // Rotate 90 degrees to match main hexagon
    geometry.rotateZ(Math.PI / 2);
    return geometry;
  }, []);
  
  // Get color based on tile type and state
  const tileColor = useMemo(() => {
    if (hovered) {
      return new THREE.Color('#ffff88');
    }
    
    switch (tileType) {
      case 'ocean':
        return new THREE.Color('#2196F3');
      case 'greenery':
        return new THREE.Color('#4CAF50');
      case 'city':
        return new THREE.Color('#FF9800');
      case 'special':
        return new THREE.Color('#9C27B0');
      default:
        return isOceanSpace ? 
          new THREE.Color('#1565C0') : 
          new THREE.Color('#8D6E63');
    }
  }, [tileType, isOceanSpace, hovered]);

  // Get border color based on tile type
  const borderColor = useMemo(() => {
    switch (tileType) {
      case 'ocean':
        return new THREE.Color('#0D47A1');
      case 'greenery':
        return new THREE.Color('#2E7D32');
      case 'city':
        return new THREE.Color('#E65100');
      case 'special':
        return new THREE.Color('#6A1B9A');
      default:
        return isOceanSpace ? 
          new THREE.Color('#0D47A1') : 
          new THREE.Color('#5D4037');
    }
  }, [tileType, isOceanSpace]);
  
  // Calculate normal vector pointing outward from sphere center
  const normal = useMemo(() => {
    const vec = new THREE.Vector3(...position);
    return vec.normalize();
  }, [position]);
  
  // Position tile slightly above sphere surface
  const adjustedPosition = useMemo(() => {
    const vec = new THREE.Vector3(...position);
    return vec.add(normal.clone().multiplyScalar(0.06));
  }, [position, normal]);
  
  // Make tiles face along the surface normal (perpendicular to Mars surface)
  const surfaceQuaternion = useMemo(() => {
    const up = new THREE.Vector3(0, 0, 1); // Default up direction
    const quat = new THREE.Quaternion();
    quat.setFromUnitVectors(up, normal);
    return quat;
  }, [normal]);
  
  // Get bonus display text
  const bonusText = useMemo(() => {
    const bonusEntries = Object.entries(bonuses);
    if (bonusEntries.length === 0) return '';
    
    return bonusEntries.map(([key, value]) => {
      const symbol = key === 'steel' ? '🔧' : 
                    key === 'titanium' ? '⚙️' : 
                    key === 'plants' ? '🌱' : 
                    key === 'cards' ? '📋' : '💰';
      return `${symbol}${value}`;
    }).join(' ');
  }, [bonuses]);
  
  // Subtle hover animation
  useFrame((state) => {
    if (meshRef.current && hovered) {
      const scale = 1 + Math.sin(state.clock.elapsedTime * 3) * 0.05;
      meshRef.current.scale.setScalar(scale);
    } else if (meshRef.current) {
      meshRef.current.scale.setScalar(1);
    }
  });
  
  return (
    <group ref={groupRef} position={adjustedPosition} quaternion={surfaceQuaternion}>
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
          transparent={tileType === 'empty'}
          opacity={tileType === 'empty' ? 0.6 : 1}
          roughness={0.7}
          metalness={0.1}
        />
      </mesh>
      
      {/* Hexagon border */}
      <mesh
        geometry={hexBorderGeometry}
        position={[0, 0, 0.001]}
      >
        <meshBasicMaterial 
          color={borderColor} 
          transparent
          opacity={0.9}
        />
      </mesh>
      
      {/* Tile type indicator */}
      {tileType !== 'empty' && (
        <Text
          position={[0, 0, 0.01]}
          fontSize={0.08}
          color="white"
          anchorX="center"
          anchorY="middle"
        >
          {tileType === 'ocean' ? '🌊' :
           tileType === 'greenery' ? '🌲' :
           tileType === 'city' ? '🏙️' :
           tileType === 'special' ? '⭐' : ''}
        </Text>
      )}
      
      {/* Bonus resources indicator */}
      {bonusText && (
        <Text
          position={[0, -0.08, 0.01]}
          fontSize={0.05}
          color="white"
          anchorX="center"
          anchorY="middle"
        >
          {bonusText}
        </Text>
      )}
      
      {/* Ocean space indicator for empty tiles */}
      {tileType === 'empty' && isOceanSpace && (
        <mesh position={[0, 0, -0.01]}>
          <ringGeometry args={[0.08, 0.11, 6]} />
          <meshBasicMaterial color="#1976D2" transparent opacity={0.7} />
        </mesh>
      )}
      
      {/* Owner indicator */}
      {ownerId && (
        <mesh position={[0.08, 0.08, 0.01]}>
          <sphereGeometry args={[0.02]} />
          <meshBasicMaterial color={`hsl(${ownerId.charCodeAt(0) * 137.5 % 360}, 70%, 50%)`} />
        </mesh>
      )}
    </group>
  );
}