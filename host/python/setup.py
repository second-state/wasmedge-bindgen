# -*- coding: utf-8 -*-

from setuptools import setup, find_packages
import os

with open('README.md') as f:
    readme = f.read()

with open(os.path.abspath(os.path.join(__file__,'../../../LICENSE'))) as f:
    license = f.read()

setup(
    name='WasmEdgeBindgen',
    version='0.1.0',
    description='WasmEdge Bindgen on top of WasmEdge Python SDK',
    long_description=readme,
    author='Shreyas Atre',
    author_email='shreyasatre16@gmail.com',
    url='https://github.com/second-state/wasmedge-bindgen',
    license=license,
    packages=find_packages(exclude=('tests', 'docs'))
)
