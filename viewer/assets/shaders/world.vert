#version 330 core

layout (location = 0) in vec3 position;
layout (location = 1) in vec2 uv;
layout (location = 2) in float shade;

uniform mat4 viewProjection;

out vec2 texCoord;
out float brightness;

void main() {
	gl_Position = viewProjection * vec4(position, 1.0);
	texCoord = uv;
	brightness = shade;
}
