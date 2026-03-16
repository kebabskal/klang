# Raylib demo — fps
Main:class

window_width := 800.0
window_height := 600.0
time := 0.0
player :Player 


update() {
	with rl

	window_width = get_render_width()
	window_height = get_render_height()
	
	delta_time := get_frame_time()
	player.update(delta_time)
}

render() {
	with math
	with rl

	begin_drawing()
	clear_background(color(32, 32, 32, 255))

	camera := player.get_camera()
	begin_mode_3d(camera)

	draw_grid(30, 2.0)

	end_mode_3d()

	end_drawing()
}

main() {
	with rl
	
	set_config_flags(Flag.WindowResizable)
	init_window(window_width, window_height, "klang fps")
	set_exit_key(Key.Zero)
	
	set_target_fps(240)
	
	player = Player()

	while not window_should_close() {
		time += get_frame_time()
		update()
		render()
	}

	close_window()
}
