#version 100
precision highp float;

uniform float time;
uniform float opacity;
varying vec2 vUv;

void main() {
  vec2 center = vUv - 0.5;
  float distFromCenter = length(center);
  float gradient = smoothstep(0.15, 0.45, distFromCenter);
  vec3 glowColor = vec3(0.95, 0.95, 1.0);
  vec3 finalColor = glowColor;
  float alpha = gradient * opacity;
  gl_FragColor = vec4(finalColor, alpha);
}
