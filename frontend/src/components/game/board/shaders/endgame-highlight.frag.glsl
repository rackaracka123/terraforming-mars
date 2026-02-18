#version 100
precision highp float;

uniform float time;
uniform vec3 highlightColor;
uniform float opacity;
varying vec2 vUv;

void main() {
  vec2 center = vUv - 0.5;
  float distFromCenter = length(center);
  float gradient = smoothstep(0.1, 0.45, distFromCenter);
  float pulse = 0.6 + 0.4 * sin(time * 3.0);
  float alpha = gradient * pulse * 0.7 * opacity;
  gl_FragColor = vec4(highlightColor, alpha);
}
