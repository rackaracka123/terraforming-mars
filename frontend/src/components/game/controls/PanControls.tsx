import { useRef, useEffect } from 'react';
import { useThree } from '@react-three/fiber';
import * as THREE from 'three';

export function PanControls() {
  const { camera, gl } = useThree();
  const isPointerDown = useRef(false);
  const previousPointer = useRef({ x: 0, y: 0 });
  const panOffset = useRef(new THREE.Vector3());

  useEffect(() => {
    const handlePointerDown = (event: PointerEvent) => {
      isPointerDown.current = true;
      previousPointer.current = { x: event.clientX, y: event.clientY };
      gl.domElement.style.cursor = 'grabbing';
    };

    const handlePointerMove = (event: PointerEvent) => {
      if (!isPointerDown.current) return;

      const deltaX = event.clientX - previousPointer.current.x;
      const deltaY = event.clientY - previousPointer.current.y;

      // Convert screen space movement to world space panning
      const panSpeed = 0.01;
      const panDeltaX = -deltaX * panSpeed;
      const panDeltaY = deltaY * panSpeed;

      // Apply panning relative to camera's current orientation
      const panVector = new THREE.Vector3(panDeltaX, panDeltaY, 0);
      panVector.applyMatrix3(new THREE.Matrix3().getNormalMatrix(camera.matrixWorld));
      
      camera.position.add(panVector);
      panOffset.current.add(panVector);

      previousPointer.current = { x: event.clientX, y: event.clientY };
    };

    const handlePointerUp = () => {
      isPointerDown.current = false;
      gl.domElement.style.cursor = 'grab';
    };

    const handleWheel = (event: WheelEvent) => {
      event.preventDefault();
      
      // Zoom functionality
      const zoomSpeed = 0.001;
      const zoomDelta = -event.deltaY * zoomSpeed;
      
      // Move camera forward/backward along its looking direction
      const direction = new THREE.Vector3();
      camera.getWorldDirection(direction);
      direction.multiplyScalar(zoomDelta);
      
      camera.position.add(direction);
      
      // Clamp zoom distance
      const distanceFromOrigin = camera.position.length();
      const minDistance = 3;
      const maxDistance = 20;
      
      if (distanceFromOrigin < minDistance) {
        camera.position.normalize().multiplyScalar(minDistance);
      } else if (distanceFromOrigin > maxDistance) {
        camera.position.normalize().multiplyScalar(maxDistance);
      }
    };

    const domElement = gl.domElement;
    
    // Set initial cursor
    domElement.style.cursor = 'grab';
    
    // Add event listeners
    domElement.addEventListener('pointerdown', handlePointerDown);
    domElement.addEventListener('pointermove', handlePointerMove);
    domElement.addEventListener('pointerup', handlePointerUp);
    domElement.addEventListener('pointerleave', handlePointerUp);
    domElement.addEventListener('wheel', handleWheel, { passive: false });

    return () => {
      domElement.removeEventListener('pointerdown', handlePointerDown);
      domElement.removeEventListener('pointermove', handlePointerMove);
      domElement.removeEventListener('pointerup', handlePointerUp);
      domElement.removeEventListener('pointerleave', handlePointerUp);
      domElement.removeEventListener('wheel', handleWheel);
    };
  }, [camera, gl]);

  return null;
}