import { useRef, useEffect, useState } from "react";
import { useThree, useFrame } from "@react-three/fiber";
import * as THREE from "three";

export function PanControls() {
  const { camera, gl, size } = useThree();
  const isPointerDown = useRef(false);
  const previousPointer = useRef({ x: 0, y: 0 });
  const [shouldRecenter, setShouldRecenter] = useState(false);
  const previousSize = useRef({ width: size.width, height: size.height });

  // Spherical coordinate state for camera orbit
  const [spherical] = useState(() => {
    // Initialize spherical coordinates from current camera position
    const spherical = new THREE.Spherical();
    spherical.setFromVector3(camera.position);
    // Ensure initial radius is reasonable
    if (spherical.radius < 3) spherical.radius = 8;
    return spherical;
  });

  // Mars center position (origin)
  const marsCenter = new THREE.Vector3(0, 0, 0);

  // Use frame loop to handle recentering smoothly and detect size changes
  useFrame(() => {
    // Check if size has changed
    if (size.width !== previousSize.current.width || size.height !== previousSize.current.height) {
      // Canvas size changed, trigger recentering
      setShouldRecenter(true);
      previousSize.current = { width: size.width, height: size.height };
    }

    if (shouldRecenter) {
      // Reset spherical coordinates to center position
      spherical.theta = 0; // Front view
      spherical.phi = Math.PI / 2; // Equatorial plane
      // Keep current radius

      // Update camera position and orientation
      camera.position.setFromSpherical(spherical);
      camera.lookAt(marsCenter);

      setShouldRecenter(false);
    }
  });

  useEffect(() => {
    // Set up initial camera position and orientation
    camera.position.setFromSpherical(spherical);
    camera.lookAt(marsCenter);

    // Handle window resize to keep Mars centered
    const handleWindowResize = () => {
      // Trigger recentering in the next frame
      setShouldRecenter(true);
    };

    const handlePointerDown = (event: PointerEvent) => {
      isPointerDown.current = true;
      previousPointer.current = { x: event.clientX, y: event.clientY };
      gl.domElement.style.cursor = "grabbing";

      // Add document-level listeners for global pointer tracking
      document.addEventListener("pointermove", handlePointerMove);
      document.addEventListener("pointerup", handlePointerUp);
    };

    const handlePointerMove = (event: PointerEvent) => {
      if (!isPointerDown.current) return;

      const deltaX = event.clientX - previousPointer.current.x;
      const deltaY = event.clientY - previousPointer.current.y;

      // Convert screen space movement to spherical coordinate changes
      const orbitSpeed = 0.001; // Much slower movement for both directions

      // Horizontal movement affects azimuthal angle (theta)
      spherical.theta -= deltaX * orbitSpeed;

      // Vertical movement affects polar angle (phi) - inverted for natural feel
      spherical.phi -= deltaY * orbitSpeed;

      // Restrict movement to ±15 degrees from center position
      const maxAngle = Math.PI / 12; // 15 degrees in radians

      // For horizontal movement (theta), restrict around the initial front view
      const centerTheta = 0; // Front view is at 0
      spherical.theta = Math.max(
        centerTheta - maxAngle,
        Math.min(centerTheta + maxAngle, spherical.theta),
      );

      // For vertical movement (phi), restrict around equatorial plane (π/2)
      const equator = Math.PI / 2;
      spherical.phi = Math.max(equator - maxAngle, Math.min(equator + maxAngle, spherical.phi));

      // Update camera position based on new spherical coordinates
      camera.position.setFromSpherical(spherical);
      camera.lookAt(marsCenter);

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

      // Zoom by changing the radius in spherical coordinates
      const zoomSpeed = 0.5;
      const zoomDelta = event.deltaY * zoomSpeed * 0.01;

      spherical.radius += zoomDelta;

      // Clamp zoom distance
      const minDistance = 3;
      const maxDistance = 20;
      spherical.radius = Math.max(minDistance, Math.min(maxDistance, spherical.radius));

      // Update camera position
      camera.position.setFromSpherical(spherical);
      camera.lookAt(marsCenter);
    };

    const domElement = gl.domElement;

    // Set initial cursor
    domElement.style.cursor = "grab";

    // Add event listeners - only pointerdown on canvas, others handled dynamically
    domElement.addEventListener("pointerdown", handlePointerDown);
    domElement.addEventListener("wheel", handleWheel, { passive: false });
    window.addEventListener("resize", handleWindowResize);

    return () => {
      domElement.removeEventListener("pointerdown", handlePointerDown);
      domElement.removeEventListener("wheel", handleWheel);

      // Clean up any remaining document listeners in case component unmounts during drag
      document.removeEventListener("pointermove", handlePointerMove);
      document.removeEventListener("pointerup", handlePointerUp);
      window.removeEventListener("resize", handleWindowResize);
    };
  }, [camera, gl, spherical, marsCenter]);

  return null;
}
