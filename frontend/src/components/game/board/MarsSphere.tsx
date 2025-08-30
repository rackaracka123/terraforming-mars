import React, { useRef, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';


interface MarsSphereProps {
  gameState?: any;
  onHexClick?: (hex: any) => void;
}



export default function MarsSphere({ gameState, onHexClick }: MarsSphereProps) {
  const sphereRef = useRef<THREE.Mesh>(null);
  const groupRef = useRef<THREE.Group>(null);
  
  
  // Slow rotation animation
  useFrame((state) => {
    if (groupRef.current) {
      groupRef.current.rotation.y = state.clock.elapsedTime * 0.05;
    }
  });
  
  // Get Mars color based on terraforming progress
  const marsColor = useMemo(() => {
    const temp = gameState?.globalParameters?.temperature || -30;
    const oxygen = gameState?.globalParameters?.oxygen || 0;
    
    const tempProgress = Math.max(0, (temp + 30) / 38);
    const oxygenProgress = oxygen / 14;
    
    const red = 1 - tempProgress * 0.3;
    const green = tempProgress * 0.2 + oxygenProgress * 0.3;
    const blue = oxygenProgress * 0.2;
    
    return new THREE.Color(red, green, blue);
  }, [gameState?.globalParameters]);
  
  return (
    <group ref={groupRef}>
      {/* Main Mars sphere */}
      <mesh ref={sphereRef} position={[0, 0, 0]}>
        <sphereGeometry args={[3, 32, 32]} />
        <meshStandardMaterial 
          color={marsColor}
          roughness={0.8}
          metalness={0.1}
        />
      </mesh>
    </group>
  );
}