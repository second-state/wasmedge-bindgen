.ONESHELL:
ENV_PREFIX=$(shell python -c "if __import__('pathlib').Path('.venv/bin/pip').exists(): print('.venv/bin/')")

.PHONY: help
help:             ## Show the help.
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@fgrep "##" Makefile | fgrep -v fgrep

.PHONY: install
install:          ## Install the project in dev mode.
	@echo "Don't forget to run 'make virtualenv' if you got errors."
	$(ENV_PREFIX)pip install -e .[test]

.PHONY: pypi-test
pypi-test:          ## PyPi testing
	$(ENV_PREFIX)pip install build twine
	$(ENV_PREFIX)python -m build
	$(ENV_PREFIX)twine check dist/* && $(ENV_PREFIX)twine upload --repository testpypi dist/* --verbose

.PHONY: fmt
fmt:              ## Format code using black & isort.
	$(ENV_PREFIX)isort WasmEdgeBindgen/ examples/
	$(ENV_PREFIX)black -l 79 WasmEdgeBindgen/ examples/

.PHONY: lint
lint:             ## Run pep8, black, mypy linters.
	$(ENV_PREFIX)flake8 WasmEdgeBindgen/ examples/
	$(ENV_PREFIX)black -l 79 --check WasmEdgeBindgen/ examples/
	$(ENV_PREFIX)mypy --ignore-missing-imports WasmEdgeBindgen/ examples/

.PHONY: clean
clean:            ## Clean unused files.
	@find ./ -name '*.pyc' -exec rm -f {} \;
	@find ./ -name '__pycache__' -exec rm -rf {} \;
	@find ./ -name 'Thumbs.db' -exec rm -f {} \;
	@find ./ -name '*~' -exec rm -f {} \;
	@rm -rf .cache
	@rm -rf .pytest_cache
	@rm -rf .mypy_cache
	@rm -rf build
	@rm -rf dist
	@rm -rf *.egg-info
	@rm -rf htmlcov
	@rm -rf .tox/
	@rm -rf docs/_build

.PHONY: virtualenv
virtualenv:       ## Create a virtual environment.
	@echo "creating virtualenv ..."
	@rm -rf .venv
	@python3 -m venv .venv
	@./.venv/bin/pip install -U pip
	@./.venv/bin/pip install -i https://test.pypi.org/simple/ WasmEdge>=0.2.2
	@./.venv/bin/python3 setup.py install
	@echo
	@echo "!!! Please run 'source .venv/bin/activate' to enable the environment !!!"

.PHONY: release
release:          ## Create a new tag for release.
	@echo "WARNING: This operation will create s version tag and push to github"
	@read -p "Version? (provide the next x.y.z semver) : " TAG
	@echo "creating git tag : $${TAG}"
	@git tag $${TAG}
	@echo "$${TAG}" > VERSION
	@git add VERSION
	@git commit -m "release: version $${TAG} 🚀"
	@git push -u origin HEAD --tags
	@echo "Github Actions will detect the new tag and release the new version."



