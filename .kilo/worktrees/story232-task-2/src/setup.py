"""Setup configuration for the ChronoCone CLI.

This module defines the package setup including console script entry points.
"""

from setuptools import setup, find_packages

setup(
    name="chronocone-cli",
    version="0.1.0",
    description="ChronoCone CLI - Manage metadata in the metrics service",
    author="ChronoCone Team",
    packages=find_packages(),
    python_requires=">=3.11",
    install_requires=[
        "click>=8.1.0",
        "requests>=2.25.0",
        "rich>=13.0.0",
    ],
    entry_points={
        "console_scripts": [
            "chronocone=cli.main:cli",
        ],
    },
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
    ],
)
