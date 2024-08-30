build docker:
	@$(MAKE) -C backend $@
	@$(MAKE) -C frontend $@
