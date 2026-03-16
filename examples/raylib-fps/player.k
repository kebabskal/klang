Player:class

position:vec3 = vec3(0,0,0)
velocity:vec3 = vec3(0,0,0)
rotation:vec3 = vec3(0,0,0)

update(delta:float) {
	with rl

	if is_mouse_button_pressed(0) {
		disable_cursor()
	}

	if is_key_pressed(Key.Escape) or not is_window_focused() {
		enable_cursor()
	}

	if is_cursor_hidden() {
		mouse_delta := get_mouse_delta()
		rotation.x -= mouse_delta.y * 0.0005
		rotation.y -= mouse_delta.x * 0.0005
	}
}

get_camera():Camera3D {
	with vector

	yaw_q := quat_from_axis_angle(vec3(0.0, 1.0, 0.0), rotation.y)
	pitch_q := quat_from_axis_angle(vec3(1.0, 0.0, 0.0), rotation.x)

	rotation := quat_multiply(yaw_q, pitch_q)
	forward := quat_rotate_vec3(rotation, vec3(0.0, 0.0, -1.0))

	camera_position := position + vec3(0, 1.5, 0)

	return rl.camera3d(
		camera_position,
		camera_position + forward,
		vec3(0,1,0),
		50,
		CameraProjection.Perspective
	)
}