.PHONY: dev

dev:
	@trap 'kill 0' INT TERM; \
	air & \
	npm run dev:css & \
	npm run dev:js & \
	wait
