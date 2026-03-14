# Raylib demo — bouncing balls
Main:class

Ball:class {
	position:vec2
	velocity:vec2
	radius:float = 10

	update(delta:float) {
		position += velocity * vec2(delta, delta)
	}
}

time := 0.0
balls:List<Ball> = []
window_width := 800.0
window_height := 600.0

update() {
	with rl
	dt := get_frame_time()
	time += dt

	# update balls
	for ball in balls {
		ball.update(dt)

		# Bounce off walls
		if ball.position.x - ball.radius < 0.0 {
			ball.velocity = vec2(math.abs(ball.velocity.x), ball.velocity.y)
		}
		if ball.position.x + ball.radius > window_width {
			ball.velocity = vec2(-math.abs(ball.velocity.x), ball.velocity.y)
		}
		if ball.position.y - ball.radius < 0.0 {
			ball.velocity = vec2(ball.velocity.x, math.abs(ball.velocity.y))
		}
		if ball.position.y + ball.radius > window_height {
			ball.velocity = vec2(ball.velocity.x, -math.abs(ball.velocity.y))
		}
	}
}

render() {
	with rl
	black := Color(0, 0, 0, 255)
	white := Color(255, 255, 255, 255)
	gray := Color(64, 64, 64, 255)
	magenta := Color(255, 0, 255, 255)
	
	# render
	begin_drawing()
	clear_background(black)
	
	# render balls
	for ball in balls {
		draw_circle_v(ball.position, ball.radius, gray)
	}

	for i in 10 {
		draw_text(
			"x", 
			100 + i * 10, 
			window_height/2 + sin(i/3.0 + time * 3.0) * 100.0, 
			20, 
			white
		)
	}

	# fps
	draw_fps(10, 10)

	end_drawing()
}



main() {
	with rl

	init_window(window_width, window_height, "klang raylib demo")
	set_target_fps(240)


	rnd := Random(1337)
	for i in 10 {
		ball := Ball()
		ball.position = vec2(rnd.rangef(100, window_width-100), rnd.rangef(100, window_height-100))
		ball.velocity = vec2(rnd.rangef(-500, 500), rnd.rangef(-500, 500))
		ball.radius = rnd.rangef(5, 20)
		balls.append(ball)
	}

	an_unused_variable_that_should_trigger_a_warning := "yeah"

	while not window_should_close() {
		update()
		render()
	}

	close_window()
}
