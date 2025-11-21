SHELL := /bin/bash

PROJECT_REF ?= fzqcgjwsewkytrijenkf
LOCAL_SUPABASE ?= yarn supabase
SUPABASE ?= supabase
BREW ?= brew
NODE_OPTIONS ?=

.PHONY: help \
        local-install-cli local-install-node local-bootstrap \
        local-init local-start local-stop local-status local-reset \
        quick-start-macos \
        prod-link prod-new-migration prod-db-push apply-migrations-prod

help:
	@echo "Supabase make targets"
	@echo "Local:"
	@echo "  quick-start-macos      Install CLI, init, and start Supabase locally"
	@echo "  local-install-cli      Install the Supabase CLI with Homebrew"
	@echo "  local-install-node     Add the Supabase CLI npm package (dev dependency)"
	@echo "  local-init             Initialize Supabase config in this repo"
	@echo "  local-start            Start local Supabase stack (Docker)"
	@echo "  local-stop             Stop the local Supabase stack"
	@echo "  local-status           Show status of local Supabase services"
	@echo "  local-reset            Reset local database (DANGEROUS)"
	@echo "Production:"
	@echo "  prod-link              Link CLI to the remote Supabase project"
	@echo "  prod-new-migration     Create a new migration (requires name=<name>)"
	@echo "  prod-db-push           Push migrations to the linked project"
	@echo "  apply-migrations-prod  Alias for prod-db-push (for backwards compatibility)"

local-install-cli:
	$(BREW) install supabase/tap/supabase

local-install-node:
	NODE_OPTIONS=$(NODE_OPTIONS) yarn add supabase --dev

local-bootstrap: local-install-cli local-install-node

local-init:
	NODE_OPTIONS=$(NODE_OPTIONS) $(LOCAL_SUPABASE) init

local-start:
	$(LOCAL_SUPABASE) start

local-stop:
	$(LOCAL_SUPABASE) stop

local-status:
	$(LOCAL_SUPABASE) status

local-reset:
	$(LOCAL_SUPABASE) db reset --force

quick-start-macos: local-bootstrap local-init local-start

prod-link:
	@test -n "$(SUPABASE_ACCESS_TOKEN)" || (echo "Usage: make prod-link SUPABASE_ACCESS_TOKEN=<access-token>" && exit 1)
	$(SUPABASE) link --project-ref $(PROJECT_REF)

prod-new-migration:
	@test -n "$(name)" || (echo "Usage: make prod-new-migration name=<migration-name>" && exit 1)
	$(SUPABASE) migration new $(name)

prod-db-push:
	$(SUPABASE) db push

apply-migrations-prod: prod-db-push