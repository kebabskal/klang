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
	rl.begin_drawing()
		rl.clear_background(rl.color(32, 32, 32, 255))

		rl.draw_text("Raylib with klang!", 20, 20, 20, rl.color(200, 200, 200, 255))

		count := 1000
		for i in count {
			a := get_point(i * 0.061 + time * 1.5, 500)
			b := get_point(i * 0.062 + time * 1.5 - 0.1, 250)
			rl.draw_line_v(
				a,
				b,
				rl.color(220, 220, 220, 255)
			)
		}
		
		rl.end_drawing()
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
