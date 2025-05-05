dev: check-modd-exists
	@modd -f ./.modd/server.modd.conf

check-modd-exists:
	@modd --version > /dev/null
