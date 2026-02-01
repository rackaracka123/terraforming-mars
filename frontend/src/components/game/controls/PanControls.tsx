import { useRef, useEffect, useState } from "react";
import { useThree, useFrame } from "@react-three/fiber";
import * as THREE from "three";

export const panState = { isPanning: false };

export function PanControls() {
  const { camera, gl, size } = useThree();
  const isPointerDown = useRef(false);
  const previousPointer = useRef({ x: 0, y: 0 });
  const [shouldRecenter, setShouldRecenter] = useState(false);
  const previousSize = useRef({ width: size.width, height: size.height });

  const [spherical] = useState(() => {
    const s = new THREE.Spherical();
    s.setFromVector3(camera.position);
    if (s.radius < 3) s.radius = 8;
    return s;
  });

  const targetSpherical = useRef(
    new THREE.Spherical(spherical.radius, spherical.phi, spherical.theta),
  );

  const marsCenter = new THREE.Vector3(0, 0, 0);

  useFrame(() => {
    if (size.width !== previousSize.current.width || size.height !== previousSize.current.height) {
      setShouldRecenter(true);
      previousSize.current = { width: size.width, height: size.height };
    }

    if (shouldRecenter) {
      targetSpherical.current.theta = 0;
      targetSpherical.current.phi = Math.PI / 2;
      setShouldRecenter(false);
    }

    const lerpFactor = 0.1;

    spherical.theta += (targetSpherical.current.theta - spherical.theta) * lerpFactor;
    spherical.phi += (targetSpherical.current.phi - spherical.phi) * lerpFactor;
    spherical.radius += (targetSpherical.current.radius - spherical.radius) * lerpFactor;

    camera.position.setFromSpherical(spherical);
    camera.lookAt(marsCenter);
  });

  useEffect(() => {
    camera.position.setFromSpherical(spherical);
    camera.lookAt(marsCenter);

    const handleWindowResize = () => {
      setShouldRecenter(true);
    };

    const handlePointerDown = (event: PointerEvent) => {
      isPointerDown.current = true;
      panState.isPanning = true;
      previousPointer.current = { x: event.clientX, y: event.clientY };
      gl.domElement.style.cursor = "grabbing";

      document.addEventListener("pointermove", handlePointerMove);
      document.addEventListener("pointerup", handlePointerUp);
    };

    const handlePointerMove = (event: PointerEvent) => {
      if (!isPointerDown.current) return;

      const deltaX = event.clientX - previousPointer.current.x;
      const deltaY = event.clientY - previousPointer.current.y;

      const orbitSpeed = 0.001;
      const maxAngle = Math.PI / 4;
      const equator = Math.PI / 2;

      targetSpherical.current.theta -= deltaX * orbitSpeed;
      targetSpherical.current.phi -= deltaY * orbitSpeed;

      targetSpherical.current.theta = Math.max(
        -maxAngle,
        Math.min(maxAngle, targetSpherical.current.theta),
      );
      targetSpherical.current.phi = Math.max(
        equator - maxAngle,
        Math.min(equator + maxAngle, targetSpherical.current.phi),
      );

      previousPointer.current = { x: event.clientX, y: event.clientY };
    };

    const handlePointerUp = () => {
      isPointerDown.current = false;
      panState.isPanning = false;
      gl.domElement.style.cursor = "grab";

      document.removeEventListener("pointermove", handlePointerMove);
      document.removeEventListener("pointerup", handlePointerUp);
    };

    const handleWheel = (event: WheelEvent) => {
      event.preventDefault();

      const zoomSpeed = 0.5;
      const zoomDelta = event.deltaY * zoomSpeed * 0.01;
      const minDistance = 2.4;
      const maxDistance = 20;

      targetSpherical.current.radius += zoomDelta;
      targetSpherical.current.radius = Math.max(
        minDistance,
        Math.min(maxDistance, targetSpherical.current.radius),
      );
    };

    const domElement = gl.domElement;

    domElement.style.cursor = "grab";

    domElement.addEventListener("pointerdown", handlePointerDown);
    domElement.addEventListener("wheel", handleWheel, { passive: false });
    window.addEventListener("resize", handleWindowResize);

    return () => {
      domElement.removeEventListener("pointerdown", handlePointerDown);
      domElement.removeEventListener("wheel", handleWheel);

      document.removeEventListener("pointermove", handlePointerMove);
      document.removeEventListener("pointerup", handlePointerUp);
      window.removeEventListener("resize", handleWindowResize);
    };
  }, [camera, gl, spherical, marsCenter]);

  return null;
}
