#version 100
precision highp float;

uniform float uSphereRadius;

varying vec4 worldPosition;
varying vec3 vNormal;
varying vec3 vViewSphereNormal;
varying vec3 vTangent;
varying vec3 vBitangent;
varying vec2 vUv;
varying vec3 vViewNormal;

void main() {
  vUv = uv;

  vec4 worldPos = modelMatrix * vec4(position, 1.0);

  vec3 sphereDir = normalize(worldPos.xyz);
  vec3 projectedPos = sphereDir * (uSphereRadius + 0.002);

  worldPosition = vec4(projectedPos, 1.0);

  vNormal = normalize(mat3(modelMatrix) * normal);
  vViewSphereNormal = normalize((viewMatrix * vec4(sphereDir, 0.0)).xyz);

  vec3 up = abs(vNormal.y) < 0.999 ? vec3(0.0, 1.0, 0.0) : vec3(1.0, 0.0, 0.0);
  vTangent = normalize(cross(up, vNormal));
  vBitangent = cross(vNormal, vTangent);

  vViewNormal = normalize(normalMatrix * normal);

  gl_Position = projectionMatrix * viewMatrix * vec4(projectedPos, 1.0);
}
