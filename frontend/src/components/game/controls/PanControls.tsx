import { useRef, useEffect, useState } from "react";
import { useThree, useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { useWorld3DSettings } from "../../../contexts/World3DSettingsContext";

export const panState = { isPanning: false };

export function PanControls() {
  const { camera, gl, size } = useThree();
  const { settings, storedCameraState, setStoredCameraState } = useWorld3DSettings();
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
  const wasFreeCameraEnabled = useRef(settings.freeCameraEnabled);

  useFrame(() => {
    // Handle free camera toggle
    if (settings.freeCameraEnabled && !wasFreeCameraEnabled.current) {
      // Switching TO free camera - store current state
      setStoredCameraState({
        position: { x: camera.position.x, y: camera.position.y, z: camera.position.z },
        spherical: { radius: spherical.radius, phi: spherical.phi, theta: spherical.theta },
      });
      wasFreeCameraEnabled.current = true;
    } else if (!settings.freeCameraEnabled && wasFreeCameraEnabled.current) {
      // Switching FROM free camera - restore state
      if (storedCameraState) {
        spherical.radius = storedCameraState.spherical.radius;
        spherical.phi = storedCameraState.spherical.phi;
        spherical.theta = storedCameraState.spherical.theta;
        targetSpherical.current.radius = storedCameraState.spherical.radius;
        targetSpherical.current.phi = storedCameraState.spherical.phi;
        targetSpherical.current.theta = storedCameraState.spherical.theta;
      }
      wasFreeCameraEnabled.current = false;
    }

    // Skip normal controls when free camera is enabled
    if (settings.freeCameraEnabled) {
      return;
    }

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
    if (!settings.freeCameraEnabled) {
      camera.position.setFromSpherical(spherical);
      camera.lookAt(marsCenter);
    }

    const handleWindowResize = () => {
      setShouldRecenter(true);
    };

    const handlePointerDown = (event: PointerEvent) => {
      if (settings.freeCameraEnabled) return;
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
      if (settings.freeCameraEnabled) return;
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

    if (!settings.freeCameraEnabled) {
      domElement.style.cursor = "grab";
    }

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
  }, [camera, gl, spherical, marsCenter, settings.freeCameraEnabled]);

  return null;
}
