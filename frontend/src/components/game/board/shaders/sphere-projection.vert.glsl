#version 100
precision highp float;

uniform float uSphereRadius;
uniform float uZOffset;
varying vec2 vUv;

void main() {
  vUv = uv;
  vec4 worldPos = modelMatrix * vec4(position, 1.0);
  vec3 sphereDir = normalize(worldPos.xyz);
  vec3 projectedPos = sphereDir * (uSphereRadius + uZOffset);
  gl_Position = projectionMatrix * viewMatrix * vec4(projectedPos, 1.0);
}
