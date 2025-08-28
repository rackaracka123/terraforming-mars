import React, { useRef } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
import { Text } from '@react-three/drei';
import { PanControls } from './PanControls.tsx';
import HexGrid from './HexGrid.tsx';
import * as THREE from 'three';

// Mars disk component
function MarsDisk({ globalParams }: { globalParams: any }) {
  const meshRef = useRef<THREE.Mesh>(null);
  
  useFrame((state) => {
    if (meshRef.current) {
      meshRef.current.rotation.y = state.clock.elapsedTime * 0.1;
    }
  });

  // Color Mars based on terraforming progress
  const getMarColor = () => {
    const temp = globalParams?.temperature || -30;
    const oxygen = globalParams?.oxygen || 0;
    
    // Red to orange to green as terraforming progresses
    const tempProgress = Math.max(0, (temp + 30) / 38); // -30 to +8 = 38 degrees
    const oxygenProgress = oxygen / 14;
    
    const red = 1 - tempProgress * 0.5;
    const green = tempProgress * 0.5 + oxygenProgress * 0.5;
    const blue = oxygenProgress * 0.3;
    
    return new THREE.Color(red, green, blue);
  };

  return (
    <mesh ref={meshRef} position={[0, 0, 0]}>
      <cylinderGeometry args={[3, 3, 0.1, 32]} />
      <meshStandardMaterial 
        color={getMarColor()}
        roughness={0.8}
        metalness={0.1}
      />
      
      {/* Hexagonal grid overlay */}
      <mesh position={[0, 0.06, 0]} rotation={[-Math.PI / 2, 0, 0]}>
        <ringGeometry args={[2.9, 3, 6]} />
        <meshBasicMaterial 
          color="#ffffff" 
          transparent 
          opacity={0.3} 
          side={THREE.DoubleSide} 
        />
      </mesh>
    </mesh>
  );
}

// Background celestial bodies with parallax
function BackgroundCelestials() {
  const earthRef = useRef<THREE.Mesh>(null);
  const moon1Ref = useRef<THREE.Mesh>(null);
  const moon2Ref = useRef<THREE.Mesh>(null);

  useFrame((state) => {
    // Parallax effect - move at different speeds when camera moves
    const time = state.clock.elapsedTime;
    
    if (earthRef.current) {
      earthRef.current.position.x = Math.sin(time * 0.1) * 15;
      earthRef.current.position.y = Math.cos(time * 0.1) * 10 + 5;
      earthRef.current.rotation.y = time * 0.05;
    }
    
    if (moon1Ref.current) {
      moon1Ref.current.position.x = Math.sin(time * 0.15) * -20;
      moon1Ref.current.position.y = Math.cos(time * 0.15) * 8 - 3;
    }
    
    if (moon2Ref.current) {
      moon2Ref.current.position.x = Math.sin(time * 0.08) * 25;
      moon2Ref.current.position.y = Math.cos(time * 0.08) * 12 + 2;
    }
  });

  return (
    <>
      {/* Earth */}
      <mesh ref={earthRef} position={[15, 5, -20]}>
        <sphereGeometry args={[2, 32, 32]} />
        <meshStandardMaterial 
          color="#4a90e2"
          roughness={0.6}
          metalness={0.1}
        />
      </mesh>
      
      {/* Moons */}
      <mesh ref={moon1Ref} position={[-20, -3, -15]}>
        <sphereGeometry args={[0.8, 16, 16]} />
        <meshStandardMaterial color="#cccccc" roughness={1} />
      </mesh>
      
      <mesh ref={moon2Ref} position={[25, 2, -25]}>
        <sphereGeometry args={[0.6, 16, 16]} />
        <meshStandardMaterial color="#e0e0e0" roughness={1} />
      </mesh>

      {/* Stars */}
      {Array.from({ length: 100 }).map((_, i) => (
        <mesh
          key={i}
          position={[
            (Math.random() - 0.5) * 100,
            (Math.random() - 0.5) * 50,
            -30 - Math.random() * 20
          ]}
        >
          <sphereGeometry args={[0.05, 8, 8]} />
          <meshBasicMaterial color="#ffffff" />
        </mesh>
      ))}
    </>
  );
}

// Parameter display in 3D space
function ParameterDisplay({ globalParams }: { globalParams: any }) {
  if (!globalParams) return null;

  return (
    <>
      <Text
        position={[4, 2, 0]}
        fontSize={0.5}
        color="white"
        anchorX="left"
        anchorY="middle"
      >
        Temperature: {globalParams.temperature}°C
      </Text>
      <Text
        position={[4, 1, 0]}
        fontSize={0.5}
        color="white"
        anchorX="left"
        anchorY="middle"
      >
        Oxygen: {globalParams.oxygen}%
      </Text>
      <Text
        position={[4, 0, 0]}
        fontSize={0.5}
        color="white"
        anchorX="left"
        anchorY="middle"
      >
        Oceans: {globalParams.oceans}/9
      </Text>
    </>
  );
}

interface Game3DViewProps {
  gameState?: any;
}

export default function Game3DView({ gameState }: Game3DViewProps) {
  return (
    <div style={{ width: '100%', height: '100vh', background: '#000011' }}>
      <Canvas
        camera={{ position: [0, 5, 8], fov: 60 }}
        gl={{ antialias: true, alpha: true }}
      >
        {/* Lighting */}
        <ambientLight intensity={0.3} />
        <directionalLight
          position={[10, 10, 5]}
          intensity={1}
          castShadow
          shadow-mapSize-width={1024}
          shadow-mapSize-height={1024}
        />

        {/* 3D Scene */}
        <BackgroundCelestials />
        <MarsDisk globalParams={gameState?.globalParameters} />
        <HexGrid gameState={gameState} onHexClick={(hex) => console.log('Clicked hex:', hex)} />
        <ParameterDisplay globalParams={gameState?.globalParameters} />

        {/* Pan Controls */}
        <PanControls />
      </Canvas>
    </div>
  );
}