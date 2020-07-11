PIP := $(shell command -v pip || command -v pip3)

.PHONY: init
init: pip-deps galaxy
	terraform init

.PHONY: apply
apply: init terraform ansible

.PHONY: terraform
terraform:
	terraform apply -auto-approve

.PHONY: ansible
ansible:
	ansible-playbook site.yml

.PHONY: update
update: galaxy-update

.PHONY: pip-deps
pip-deps:
	$(PIP) install -r requirements.txt

.PHONY: galaxy
galaxy:
	ansible-galaxy install -r requirements.yml

.PHONY: galaxy-update
galaxy-update:
	ansible-galaxy install -f -r requirements.yml
