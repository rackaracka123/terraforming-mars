#version 100
precision highp float;

uniform float time;
varying vec2 vUv;

void main() {
  vec2 center = vUv - 0.5;
  float distFromCenter = length(center);
  float gradient = smoothstep(0.2, 0.45, distFromCenter);
  vec3 darkBlue = vec3(0.05, 0.28, 0.63);
  vec3 finalColor = darkBlue;
  float alpha = gradient * 0.56;
  gl_FragColor = vec4(finalColor, alpha);
}
