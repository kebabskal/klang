Player:class

position:vec3 = vec3(0,0,0)
velocity:vec3 = vec3(0,0,0)
rotation:vec3 = vec3(0,0,0)

update(delta:float) {
#	rotation.y += delta * 2.0

	rotation.y -= rl.get_mouse_delta().x * 0.001
}

get_camera():Camera3D {
	direction := vec3(
		math.cos(rotation.y),
		0,
		math.sin(rotation.y)
	)

	camera_position := position + vec3(0, 1.5, 0)

	return rl.camera3d(
		camera_position,
		camera_position + direction,
		vec3(0,1,0),
		50,
		CameraProjection.Perspective
	)
}