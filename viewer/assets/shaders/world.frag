#version 330 core

in vec2 texCoord;
in float brightness;

uniform sampler2D atlas;

out vec4 fragColor;

void main() {
	vec4 texel = texture(atlas, texCoord);
	if (texel.a < 0.5) {
		discard;
	}
	fragColor = vec4(texel.rgb * brightness, 1.0);
}
