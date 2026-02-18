#version 100
precision highp float;

varying vec4 worldPosition;
varying vec3 vNormal;
varying vec3 vViewSphereNormal;
varying vec3 vTangent;
varying vec3 vBitangent;
varying vec2 vUv;
varying vec3 vViewNormal;

uniform float time;
uniform float size;
uniform float alpha;
uniform float rf0;
uniform float sunIntensity;
uniform sampler2D normalSampler;
uniform vec3 sunColor;
uniform vec3 sunDirection;
uniform vec3 eye;
uniform vec3 waterColor;

uniform float uRadius;
uniform float uAspect;
uniform float uRotation;

uniform float uEdgeBand;
uniform float uEdgeStrength;
uniform float uEdgeScale;
uniform float uWarpScale;
uniform float uWarpAmount;

uniform float uSandWidth;
uniform float uGrainScale;
uniform sampler2D sandSampler;
uniform float uSandTexScale;

uniform float uShallowWidth;
uniform float uShallowStrength;

uniform float uEdgeSoftness;

uniform vec2 uSeedOffset;

// Foam uniforms
uniform float uFoamWidth;
uniform float uFoamStrength;
uniform float uFoamScale;
uniform float uFoamSpeed;
uniform float uFoamCutoff;
uniform float uFoamPulseSpeed;
uniform float uFoamPulseAmount;

#include <common>
#include <lights_pars_begin>

mat2 rot2(float a) {
  float c = cos(a), s = sin(a);
  return mat2(c, -s, s, c);
}

vec3 mod289(vec3 x) { return x - floor(x * (1.0 / 289.0)) * 289.0; }
vec2 mod289(vec2 x) { return x - floor(x * (1.0 / 289.0)) * 289.0; }
vec3 permute(vec3 x) { return mod289(((x * 34.0) + 1.0) * x); }

float snoise(vec2 v) {
  const vec4 C = vec4(0.211324865405187, 0.366025403784439,
                      -0.577350269189626, 0.024390243902439);
  vec2 i = floor(v + dot(v, C.yy));
  vec2 x0 = v - i + dot(i, C.xx);
  vec2 i1 = (x0.x > x0.y) ? vec2(1.0, 0.0) : vec2(0.0, 1.0);
  vec4 x12 = x0.xyxy + C.xxzz;
  x12.xy -= i1;
  i = mod289(i);
  vec3 p = permute(permute(i.y + vec3(0.0, i1.y, 1.0)) + i.x + vec3(0.0, i1.x, 1.0));
  vec3 m = max(0.5 - vec3(dot(x0, x0), dot(x12.xy, x12.xy), dot(x12.zw, x12.zw)), 0.0);
  m = m * m;
  m = m * m;
  vec3 x = 2.0 * fract(p * C.www) - 1.0;
  vec3 h = abs(x) - 0.5;
  vec3 ox = floor(x + 0.5);
  vec3 a0 = x - ox;
  m *= 1.79284291400159 - 0.85373472095314 * (a0 * a0 + h * h);
  vec3 g;
  g.x = a0.x * x0.x + h.x * x0.y;
  g.yz = a0.yz * x12.xz + h.yz * x12.yw;
  return 130.0 * dot(m, g);
}

vec4 getNoise(vec2 uv) {
  float t1 = sin(time * 0.3) * 2.0;
  float t2 = cos(time * 0.25) * 2.0;
  float t3 = sin(time * 0.2 + 1.0) * 2.0;
  float t4 = cos(time * 0.22 + 0.5) * 2.0;

  vec2 uv0 = (uv / 103.0) + vec2(t1 / 17.0, t2 / 29.0);
  vec2 uv1 = uv / 107.0 - vec2(t2 / -19.0, t1 / 31.0);
  vec2 uv2 = uv / vec2(8907.0, 9803.0) + vec2(t3 / 101.0, t4 / 97.0);
  vec2 uv3 = uv / vec2(1091.0, 1027.0) - vec2(t4 / 109.0, t3 / -113.0);
  vec4 noise = texture2D(normalSampler, uv0) +
    texture2D(normalSampler, uv1) +
    texture2D(normalSampler, uv2) +
    texture2D(normalSampler, uv3);
  return noise * 0.5 - 1.0;
}

void sunLight(const vec3 surfaceNormal, const vec3 eyeDirection, float shiny, float spec, float diffuse, inout vec3 diffuseColor, inout vec3 specularColor) {
  vec3 reflection = normalize(reflect(-sunDirection, surfaceNormal));
  float direction = max(0.0, dot(eyeDirection, reflection));
  specularColor += pow(direction, shiny) * sunColor * spec;
  diffuseColor += max(dot(sunDirection, surfaceNormal), 0.0) * sunColor * diffuse;
}

void main() {
  vec2 p = (vUv - 0.5) * 2.0;

  p = rot2(uRotation) * p;

  vec2 q = p / vec2(uRadius * uAspect, uRadius);

  float baseSdf = length(q) - 1.0;

  float edgeBand = 1.0 - smoothstep(0.0, uEdgeBand, abs(baseSdf));

  vec2 seedOff = uSeedOffset;

  float w1 = snoise(p * uWarpScale + seedOff);
  float w2 = snoise((p + 31.7) * uWarpScale + seedOff);
  vec2 warp = vec2(w1, w2) * uWarpAmount;

  float edgeNoise = snoise((p + warp) * uEdgeScale + seedOff);

  edgeNoise = edgeNoise / (1.0 + abs(edgeNoise));

  float displacedSdf = baseSdf + edgeNoise * uEdgeStrength * edgeBand;

  float aa = fwidth(displacedSdf) * 1.5;

  float waterMask = 1.0 - smoothstep(-uEdgeSoftness - aa, uEdgeSoftness + aa, displacedSdf);

  float shoreDist = displacedSdf;

  float sandJitter = snoise(p * (uEdgeScale * 1.2) + seedOff + 17.3) * 0.5 + 0.5;
  sandJitter = (sandJitter - 0.5) * 0.03;

  float sandWidth = max(0.02, uSandWidth + sandJitter);

  float shoreT = smoothstep(0.0 - aa, sandWidth + aa, shoreDist);

  float sandAlpha = (1.0 - shoreT) * step(0.0, shoreDist);
  sandAlpha *= 0.95;
  float discEdge = length((vUv - 0.5) * 2.0);
  sandAlpha *= 1.0 - smoothstep(0.75, 1.0, discEdge);
  sandAlpha = clamp(sandAlpha, 0.0, 1.0);

  float shallowMask = smoothstep(-uShallowWidth - aa, 0.0 + aa, shoreDist) * waterMask;

  // === Foam band (inside water near shoreline) ===
  float foamBand = smoothstep(-uFoamWidth - aa, 0.0 + aa, shoreDist) * waterMask;
  foamBand = pow(clamp(foamBand, 0.0, 1.0), 1.6);

  // Drifting foam noise
  vec2 foamUv = (p + warp) * uFoamScale + seedOff * 0.7;
  foamUv += vec2(time * uFoamSpeed, time * (uFoamSpeed * 0.73));

  float fn1 = snoise(foamUv);
  float fn2 = snoise(foamUv * 2.1 + 11.3);
  float fn3 = snoise(foamUv * 4.0 - 27.1);
  float foamNoise = fn1 * 0.55 + fn2 * 0.30 + fn3 * 0.15;
  foamNoise = foamNoise * 0.5 + 0.5;

  float foamSpots = smoothstep(uFoamCutoff, 1.0, foamNoise);

  // Pulse — per-lake desync via seedOff
  float pulseBase = sin(time * uFoamPulseSpeed + dot(seedOff, vec2(0.12, 0.37)));
  pulseBase = pulseBase * 0.5 + 0.5;
  float pulse = smoothstep(0.35, 0.95, pulseBase);
  pulse = mix(1.0, pulse, uFoamPulseAmount);

  float foamMask = clamp(foamBand * foamSpots * pulse, 0.0, 1.0);

  vec2 projectedPos = vec2(
    dot(worldPosition.xyz, vTangent),
    dot(worldPosition.xyz, vBitangent)
  );
  vec4 noise = getNoise(projectedPos * size);

  vec3 tangentNormal = normalize(noise.xzy * vec3(1.5, 1.0, 1.5));
  vec3 surfaceNormal = normalize(
    vTangent * tangentNormal.x +
    vNormal * tangentNormal.y +
    vBitangent * tangentNormal.z
  );

  vec3 diffuseLight = vec3(0.0);
  vec3 specularLight = vec3(0.0);
  vec3 worldToEye = eye - worldPosition.xyz;
  vec3 eyeDirection = normalize(worldToEye);

  sunLight(surfaceNormal, eyeDirection, 100.0, 2.0 * sunIntensity, 0.5 * sunIntensity, diffuseLight, specularLight);

  float theta = max(dot(eyeDirection, surfaceNormal), 0.0);
  float reflectance = rf0 + (1.0 - rf0) * pow((1.0 - theta), 5.0);
  vec3 scatter = max(0.0, dot(surfaceNormal, eyeDirection)) * waterColor;

  // Scene irradiance from Three.js lights (matches MeshStandardMaterial)
  vec3 sceneIrradiance = ambientLightColor;
  #if NUM_DIR_LIGHTS > 0
  for (int i = 0; i < NUM_DIR_LIGHTS; i++) {
    float dotNL = max(dot(vViewSphereNormal, directionalLights[i].direction), 0.0);
    sceneIrradiance += directionalLights[i].color * dotNL;
  }
  #endif

  // Sky reflection brightness tracks primary directional light
  #if NUM_DIR_LIGHTS > 0
  float envBrightness = mix(0.15, 1.0, max(dot(vViewSphereNormal, directionalLights[0].direction), 0.0));
  #else
  float envBrightness = 1.0;
  #endif

  vec3 skyColor = vec3(0.4, 0.5, 0.7) * envBrightness;
  vec3 horizonColor = vec3(0.7, 0.75, 0.85) * envBrightness;
  float skyGradient = max(0.0, surfaceNormal.y);
  vec3 reflectionSample = mix(horizonColor, skyColor, skyGradient);

  vec3 deepWaterColor = mix(
    (sunColor * diffuseLight * 0.3 + scatter),
    (vec3(0.1) + reflectionSample * 0.9 + reflectionSample * specularLight),
    reflectance
  );

  vec2 sandUv = (p + warp * 0.5) * uSandTexScale + seedOff * 0.3;
  vec3 sandCol = texture2D(sandSampler, sandUv).rgb;
  float grain = snoise(p * uGrainScale + seedOff * 2.0) * 0.5 + 0.5;
  sandCol *= mix(0.95, 1.05, grain);
  sandCol *= sceneIrradiance * RECIPROCAL_PI;

  vec3 shallowWaterColor = vec3(0.32, 0.58, 0.68) * sceneIrradiance * RECIPROCAL_PI;
  vec3 waterCol = deepWaterColor;
  waterCol = mix(waterCol, shallowWaterColor, shallowMask * uShallowStrength);

  float fres = pow(1.0 - clamp(vViewNormal.z, 0.0, 1.0), 3.0);
  waterCol += fres * 0.06;

  // Foam — brighten water towards white (dimmed on dark side)
  waterCol = mix(waterCol, vec3(envBrightness), foamMask * uFoamStrength);

  vec3 col = vec3(0.0);
  col += sandCol * sandAlpha;
  col += waterCol * waterMask;

  float wetMix = 0.10;
  col = mix(col, mix(sandCol, waterCol, 0.75), waterMask * sandAlpha * wetMix);

  float finalAlpha = clamp(waterMask + sandAlpha, 0.0, 1.0);

  float outA = alpha * finalAlpha;
  gl_FragColor = vec4(col * alpha, outA);
}
