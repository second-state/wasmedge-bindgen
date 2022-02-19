ENV_PREFIX=$(shell python -c "if __import__('pathlib').Path('.venv/bin/pip').exists(): print('.venv/bin/')")

$(ENV_PREFIX)pip install build twine
$(ENV_PREFIX)python -m build
$(ENV_PREFIX)twine check dist/* && $(ENV_PREFIX)twine upload --repository testpypi dist/* --verbose