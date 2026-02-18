uniform float uSphereRadius;
uniform float uZOffset;
//#pragma body
vec4 worldPos = modelMatrix * vec4(position, 1.0);
vec3 sphereDir = normalize(worldPos.xyz);
vec3 projectedPos = sphereDir * (uSphereRadius + uZOffset);
vec3 transformed = (inverse(modelMatrix) * vec4(projectedPos, 1.0)).xyz;
