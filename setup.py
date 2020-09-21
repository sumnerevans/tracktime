import codecs
import os

from setuptools import find_packages, setup

here = os.path.abspath(os.path.dirname(__file__))

with codecs.open(os.path.join(here, "README.rst"), encoding="utf-8") as f:
    long_description = f.read()

# Find the version
with codecs.open(os.path.join(here, "tracktime/__init__.py"), encoding="utf-8") as f:
    for line in f:
        if line.startswith("__version__"):
            version = eval(line.split()[-1])
            break

setup(
    name="tracktime",
    version=version,
    url="https://git.sr.ht/~sumner/tracktime",
    description="Time tracking library with command line interface.",
    long_description=long_description,
    author="Sumner Evans",
    author_email="sumner.evans98@gmail.com",
    license="GPL3",
    classifiers=[
        #   3 - Alpha
        #   4 - Beta
        #   5 - Production/Stable
        "Development Status :: 4 - Beta",
        # Indicate who your project is intended for
        "Intended Audience :: End Users/Desktop",
        "Operating System :: POSIX",
        "License :: OSI Approved :: GNU General Public License v3 (GPLv3)",
        # Specify the Python versions you support here. In particular, ensure
        # that you indicate whether you support Python 2, Python 3 or both.
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
    ],
    keywords="time tracking",
    packages=find_packages(exclude=["tests"]),
    install_requires=[
        "tabulate",
        "pdfkit",
        "docutils",
        "requests",
        "pyyaml",
    ],
    # To provide executable scripts, use entry points in preference to the
    # "scripts" keyword. Entry points provide cross-platform support and
    # allow pip to create the appropriate form of executable for the target
    # platform.
    entry_points={
        "console_scripts": [
            "tt=tracktime.__main__:main",
        ],
    },
)
