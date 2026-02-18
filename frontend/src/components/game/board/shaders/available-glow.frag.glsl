#version 100
precision highp float;

uniform float time;
varying vec2 vUv;

void main() {
  vec2 center = vUv - 0.5;
  float distFromCenter = length(center);
  float gradient = smoothstep(0.15, 0.45, distFromCenter);
  float pulse = 0.5 + 0.3 * sin(time * 2.0);
  vec3 glowColor = vec3(0.4, 1.0, 0.4);
  vec3 finalColor = glowColor;
  float alpha = gradient * pulse * 0.6;
  gl_FragColor = vec4(finalColor, alpha);
}
