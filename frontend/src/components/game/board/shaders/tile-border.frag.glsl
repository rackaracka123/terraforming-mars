#version 100
precision highp float;

uniform vec3 uColor;
uniform float uOpacity;
uniform sampler2D uNoiseTex;
varying vec2 vUv;
varying vec2 vWorldUv;

void main() {
  float noise = texture2D(uNoiseTex, vWorldUv).r;
  float alpha = uOpacity * smoothstep(0.25, 0.55, noise);
  gl_FragColor = vec4(uColor, alpha);
}
