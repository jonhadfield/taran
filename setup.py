#!/usr/bin/env python

import os
import re
import sys

from setuptools import find_packages
from setuptools import setup
from setuptools.command.test import test as TestCommand

sys.path.insert(0, os.path.abspath('lib'))

with open('requirements.txt') as install_reqs_file:
    install_requires = install_reqs_file.read().splitlines()

with open('test_requirements.txt') as test_reqs_file:
    tests_require = test_reqs_file.read().splitlines()
    tests_require += install_requires

if sys.argv[-1] == 'publish':
    os.system('python setup.py sdist upload')
    os.system('python setup.py bdist_wheel upload')
    sys.exit()

with open('lib/taran/__init__.py', 'r') as fd:
    version = re.search(r'^__version__\s*=\s*[\'"]([^\'"]*)[\'"]',
                        fd.read(), re.MULTILINE).group(1)

if sys.argv[-1] == 'tag':
    os.system("git tag -a %s -m 'version %s'" % (version, version))
    os.system("git push --tags")
    sys.exit()

readme = open('README.md').read()

long_description = readme + '\n\n'

if sys.argv[-1] == 'readme':
    print(long_description)
    sys.exit()


class PyTest(TestCommand):
    user_options = [('pytest-args=', 'a', "Arguments to pass to py.test")]

    def initialize_options(self):
        TestCommand.initialize_options(self)
        self.pytest_args = []

    def finalize_options(self):
        TestCommand.finalize_options(self)
        self.test_args = []
        self.test_suite = True

    def run_tests(self):
        import pytest
        errno = pytest.main(self.pytest_args)
        sys.exit(errno)


setup(
    name='taran',
    version=version,
    author='Jon Hadfield',
    author_email='jon@lessknown.co.uk',
    url='https://github.com/jonhadfield/taran',
    download_url='https://github.com/jonhadfield/taran/tarball/{0}'.format(version),
    install_requires=install_requires,
    description='An automated deployment framework',
    long_description=long_description,
    packages=find_packages('lib'),
    platforms='any',
    package_dir={'': 'lib'},
    license='MIT',
    classifiers=[
        'Programming Language :: Python',
        'Development Status :: 4 - Beta',
        'Natural Language :: English',
        'Environment :: Console',
        'Intended Audience :: System Administrators',
        'License :: OSI Approved :: MIT License',
        'Operating System :: OS Independent',
        'Topic :: System :: Operating System',
        'Topic :: System :: Networking',
    ],
    keywords=(
        'deployment, python'
    ),
    cmdclass={'test': PyTest},
    tests_require=tests_require
)
