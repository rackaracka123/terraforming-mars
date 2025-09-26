import { useRef, useEffect } from "react";
import { useThree } from "@react-three/fiber";
import * as THREE from "three";
import { useMarsRotation } from "../../../contexts/MarsRotationContext.tsx";

export function PanControls() {
  const { camera, gl } = useThree();
  const { marsGroupRef } = useMarsRotation();
  const isPointerDown = useRef(false);
  const previousPointer = useRef({ x: 0, y: 0 });

  useEffect(() => {
    const handlePointerDown = (event: PointerEvent) => {
      isPointerDown.current = true;
      previousPointer.current = { x: event.clientX, y: event.clientY };
      gl.domElement.style.cursor = "grabbing";

      // Add document-level listeners for global pointer tracking
      document.addEventListener("pointermove", handlePointerMove);
      document.addEventListener("pointerup", handlePointerUp);
    };

    const handlePointerMove = (event: PointerEvent) => {
      if (!isPointerDown.current || !marsGroupRef.current) return;

      const deltaX = event.clientX - previousPointer.current.x;
      const deltaY = event.clientY - previousPointer.current.y;

      // Convert screen space movement to rotation (inverted)
      const rotationSpeed = 0.002;
      const rotationX = deltaY * rotationSpeed; // Vertical drag rotates around X axis (inverted)
      const rotationY = deltaX * rotationSpeed; // Horizontal drag rotates around Y axis (inverted)

      // Apply rotation to Mars
      marsGroupRef.current.rotation.x += rotationX;
      marsGroupRef.current.rotation.y += rotationY;

      // Clamp X rotation to prevent flipping upside down
      marsGroupRef.current.rotation.x = Math.max(
        -Math.PI / 2,
        Math.min(Math.PI / 2, marsGroupRef.current.rotation.x),
      );

      previousPointer.current = { x: event.clientX, y: event.clientY };
    };

    const handlePointerUp = () => {
      isPointerDown.current = false;
      gl.domElement.style.cursor = "grab";

      // Remove document-level listeners when drag ends
      document.removeEventListener("pointermove", handlePointerMove);
      document.removeEventListener("pointerup", handlePointerUp);
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
    domElement.style.cursor = "grab";

    // Add event listeners - only pointerdown on canvas, others handled dynamically
    domElement.addEventListener("pointerdown", handlePointerDown);
    domElement.addEventListener("wheel", handleWheel, { passive: false });

    return () => {
      domElement.removeEventListener("pointerdown", handlePointerDown);
      domElement.removeEventListener("wheel", handleWheel);

      // Clean up any remaining document listeners in case component unmounts during drag
      document.removeEventListener("pointermove", handlePointerMove);
      document.removeEventListener("pointerup", handlePointerUp);
    };
  }, [camera, gl]);

  return null;
}
