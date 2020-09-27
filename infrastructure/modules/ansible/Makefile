PIP := $(shell command -v pip || command -v pip3)

prepare: pip-deps galaxy-deps

.PHONY: pip-deps
pip-deps:
	$(PIP) install -r requirements.txt

.PHONY: galaxy
galaxy-deps:
	ansible-galaxy install -r requirements.yml
