import codecs
import os

from setuptools import setup

here = os.path.abspath(os.path.dirname(__file__))

with codecs.open(os.path.join(here, 'README.rst'), encoding='utf-8') as f:
    long_description = f.read()

version = '0.9.0'

setup(
    name='tracktime',
    version=version,
    url='https://gitlab.com/sumner/tracktime',
    description='Time tracking library with command line interface.',
    long_description=long_description,
    author='Sumner Evans',
    author_email='sumner.evans98@gmail.com',
    license='GPL3',
    classifiers=[
        #   3 - Alpha
        #   4 - Beta
        #   5 - Production/Stable
        'Development Status :: 3 - Alpha',

        # Indicate who your project is intended for
        'Intended Audience :: End Users/Desktop',
        'Operating System :: POSIX',
        'License :: OSI Approved :: GNU General Public License v3 (GPLv3)',

        # Specify the Python versions you support here. In particular, ensure
        # that you indicate whether you support Python 2, Python 3 or both.
        'Programming Language :: Python :: 3.6',
        'Programming Language :: Python :: 3.7',
    ],
    keywords='time tracking',
    packages=['tracktime'],
    install_requires=['tabulate', 'pdfkit', 'docutils', 'requests'],

    # To provide executable scripts, use entry points in preference to the
    # "scripts" keyword. Entry points provide cross-platform support and
    # allow pip to create the appropriate form of executable for the target
    # platform.
    entry_points={
        'console_scripts': [
            'tt=tracktime.__main__:main',
        ],
    },
)
