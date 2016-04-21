#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""Common utilities"""
from __future__ import (absolute_import, print_function, unicode_literals)

import json
import os
import select
import shlex
import subprocess
import sys

from contracts import contract
from six import string_types


# Daemonize functionality taken from Ansible project
@contract(cmd='unicode', cwd='unicode|None', env='unicode|None', daemonize='bool|None')
def execute_command(cmd=None, cwd=None, env=None, daemonize=False):
    """Wrapper around function to execute a command to allow for daemonizing."""
    if not daemonize:
        return run_command(cmd=cmd, cwd=cwd, env=env)
    pipe = os.pipe()
    pid = os.fork()
    if pid == 0:
        os.close(pipe[0])
        fd = os.open(os.devnull, os.O_RDWR)
        if fd != 0:
            os.dup2(fd, 0)
        if fd != 1:
            os.dup2(fd, 1)
        if fd != 2:
            os.dup2(fd, 2)
        if fd not in (0, 1, 2):
            os.close(fd)
        pid = os.fork()
        if pid > 0:
            exit(0)
        os.setsid()
        os.chdir("/")
        pid = os.fork()
        if pid > 0:
            exit(0)

        if isinstance(cmd, string_types):
            cmd = shlex.split(cmd.encode('ascii'))
        p = subprocess.Popen(args=cmd, cwd=cwd, env=env, shell=False, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                             preexec_fn=lambda: os.close(pipe[1]))
        stdout = ""
        stderr = ""
        fds = [p.stdout, p.stderr]
        # Wait for all output, or until the main process is dead and its output is done.
        while fds:
            rfd, wfd, efd = select.select(fds, [], fds, 1)
            if not (rfd + wfd + efd) and p.poll() is not None:
                break
            if p.stdout in rfd:
                dat = os.read(p.stdout.fileno(), 4096)
                if not dat:
                    fds.remove(p.stdout)
                stdout += dat
            if p.stderr in rfd:
                dat = os.read(p.stderr.fileno(), 4096)
                if not dat:
                    fds.remove(p.stderr)
                stderr += dat
        p.wait()
        # Return a JSON blob to parent
        os.write(pipe[1], json.dumps([p.returncode, stdout, stderr]))
        os.close(pipe[1])
        exit(0)
    elif pid == -1:
        exit("unable to fork")
    else:
        os.close(pipe[1])
        os.waitpid(pid, 0)
        # Wait for data from daemon process and process it.
        data = ""
        while True:
            rfd, wfd, efd = select.select([pipe[0]], [], [pipe[0]])
            if pipe[0] in rfd:
                dat = os.read(pipe[0], 4096)
                if not dat:
                    break
                data += dat
        return json.loads(data)


@contract(cmd='unicode')
def run_command(cmd=None, cwd=None, env=None):
    """Interactively run the command and return any output."""
    if isinstance(cmd, string_types):
        cmd = shlex.split(cmd.encode('ascii'))
    output = subprocess.Popen(args=cmd, env=env, cwd=cwd, stdout=subprocess.PIPE)
    out, _ = output.communicate()
    return out.rstrip('\n')


@contract(bytestring='str')
def commandline_arg(bytestring=None):
    """Return unicode for parsed arguments."""
    unicode_string = bytestring.decode(sys.getfilesystemencoding())
    return unicode_string
