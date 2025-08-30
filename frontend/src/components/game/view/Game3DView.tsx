import React, { useRef } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
import { Text } from '@react-three/drei';
import { PanControls } from '../controls/PanControls.tsx';
import MarsSphere from '../board/MarsSphere.tsx';
import * as THREE from 'three';


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
        gl={{ antialias: true, alpha: false }}
      >
        {/* Lighting */}
        <ambientLight intensity={0.6} />
        <directionalLight
          position={[10, 10, 5]}
          intensity={1}
          castShadow
          shadow-mapSize-width={1024}
          shadow-mapSize-height={1024}
        />

        {/* 3D Scene */}
        <BackgroundCelestials />
        <MarsSphere gameState={gameState} onHexClick={(hex) => console.log('Clicked hex:', hex)} />

        {/* Pan Controls */}
        <PanControls />
      </Canvas>
    </div>
  );
}