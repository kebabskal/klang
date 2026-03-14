# klang is a opinionated programming language made for games

# - opinionated
# - fast (enough)
# - compiles to c, which compiles to platform code
# - cross platform through the c layer
# - automatic memory management (gc?)
# - object oriented
# - inspired by gdscript, c#, typescript and odin
# - no bullshit, everything is public and overridable by default for example
# - strongly typed 
# - strong generics
# - built in support for something like raylib or sdl
# - support for hot reloading code while the app is running
# - methods are just function calls under the hood, no vtables
# - kl build release|debug (builds release or debug), kl run (builds and runs with hot reload)

namespace Game # optional namespacing. folder structure isn't important

# file scoped concepts clean up brackets and indentation
# Name:type structure
Entity:class

# variables
health   :int = 10              # name:type = value
position :vec2 = vec2(10, 10)   # explicit initialization
velocity :vec2 = {10, 10}    # inferred initialization
cooldown := 5.0                 # inferring float type

# events and inline types
died:event<class { delta:int }>

# property
is_alive:bool {
	get => health > 0
	set(value) => {

		if not value {
			hurt(health)
		}	
	}
}

# constructor
new(health:int) {
	this.health = health
}

# can also have regular concepts like regular nested classes
# these use brackets and the type is accessible outside scopes as Entity.Action 
Action:class {
	# enums
	# this one can be accessed from outside the class scope as Action.Type
	Type:enum = {
		Regular = 0,
		Charged,
	}

	Result:class {
		is_success := false
	}

	name:string
	entity:Entity
	type:Type = Type.Regular
	
	# method with type options
	act(target:Entity|vec2):Result {
		result = {} # inferring type from return below
		result.is_success = true
		return result
	}	
}

# inherit from Action
DamageAction:Action {
	name = "Damage Action" # override default values
	# override method and infer params and return type
	# to have less boilerplate to update when you make architectural changes
	act(...) {
		damage_amount := rng.randi_range(5, 10)
		
		if target is Entity {
			return target.hurt(damage_amount, entity)
		}

		else if target is vec2 { 
			# fluent with optional chaining and default values
			result = Scene.get_entity_at(target)
				?.hurt(damage_amount, entity) 
				?? { is_success = false }

			return result

		} 
	}
}

TeleportAction:Action {
	name = "Teleport Action"
	act(...) {
		if target is vec2 {
			entity.position = target
			return {is_success = true}
		} else {
			return Action.act(...) # call default parent implementation
		}
	}
}

# a generic list of actions
actions:List<Action> = [
	DamageAction,
	TeleportAction,
]

# method on the file scoped class. void return is inferred
update(delta_time:float) {
	# vector maths built in
	position += velocity * delta_time
	position.y = max(position.y) 

	target := Scene.get_random_target_entity_or_position()

	for action in actions {
		if action.act(target).is_success {
			print("did it!")
		} else {
			# string templating
			print($"{action.name} failed") 
		}
	}
}

hurt(damage_amount:int, damage_source:Entity) {
	# return early if we aren't alive
	if not is_alive return

	health -= damage_amount

	# fire event if we died from this
	if not is_alive {
		died({delta = delta})
	}	
}