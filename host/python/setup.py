# -*- coding: utf-8 -*-

from setuptools import setup, find_packages
import os
import io

def read(*paths, **kwargs):
    """Read the contents of a text file safely.
    >>> read("WasmEdge", "VERSION")
    '0.1.0'
    >>> read("README.md")
    ...
    """

    content = ""
    with io.open(
        os.path.join(os.path.dirname(__file__), *paths),
        encoding=kwargs.get("encoding", "utf8"),
    ) as open_file:
        content = open_file.read().strip()
    return content

setup(
    name="WasmEdgeBindgen",
    version="0.1.0",
    description="WasmEdge Bindgen on top of WasmEdge Python SDK",
    long_description=read("README.md"),    
    long_description_content_type="text/markdown",
    author="Shreyas Atre",
    author_email="shreyasatre16@gmail.com",
    url="https://github.com/second-state/wasmedge-bindgen",
    license=read("LICENSE"),
    packages=find_packages(exclude=("tests", "docs")),
    zip_safe=False,
)
