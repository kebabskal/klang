# Raylib demo — bouncing balls
Main:class

window_width := 800.0
window_height := 600.0
time := 0.0

get_point(t:float, r:float):vec2 {
	with math {
		return vec2(
			cos(t + time) * cos(t*0.123) * r,
			sin(t*0.91 + time) * sin(t*0.231) * r,
		) + vec2(window_width/2, window_height/2)
	}
}

render() {
	with math
	with rl

	begin_drawing()
	clear_background(color(32, 32, 32, 255))

	count := 100000
	for i in count {
		f := float(i)
		p := f/count
		c := int(p * 255)

		draw_line_v(
			get_point(0.060 * f + time * 1.5, 500.0 + 50.0 * sin(f * 0.1 + f * 0.1)),
			get_point((0.060 + sin(time * 0.5) * 0.0001) * f + time * 1.5, 450.0 + 50.0 * sin(f * 0.1 + f * 0.1)),
			color(c, c, c, 20)
		)
	}
		
	draw_text("100k lines", 20, 22, 20, color(0, 0, 0, 255))
	draw_text("100k lines", 20, 20, 20, color(200, 200, 200, 255))

	draw_fps(20, 100)
	end_drawing()
}

main() {

	rl.init_window(window_width, window_height, "klang raylib demo")
	rl.set_target_fps(240)

	while not rl.window_should_close() {
		time += rl.get_frame_time()

		render()
	}

	rl.close_window()
}
