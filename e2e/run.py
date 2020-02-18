#!/usr/bin/env python3

import datetime
import os
import subprocess

import nbformat
import nbconvert

import sys

src = "tests/smoketest.ipynb"
destdir = "/workspace/artifacts"

timeout = 3600

now = datetime.datetime.utcnow()
now = now.replace(microsecond=0)
timestamp = now.strftime("%Y%m%d%H%M%s")

dest = os.path.join(destdir, timestamp + "_" + src.replace('/', '_'))
os.makedirs(os.path.abspath(os.path.dirname(dest)), exist_ok=True)

src_dir = os.path.dirname(src)

with open(src) as f:
  nb = nbformat.read(f, as_version=4)

print("executing notebook %s to %s" % (src, dest))

clear_output = nbconvert.preprocessors.ClearOutputPreprocessor()

executor = nbconvert.preprocessors.ExecutePreprocessor(timeout=timeout, kernel_name='python3')

ok = False
try:
  clear_output.preprocess(nb, {'metadata': {'path': src_dir}})
  executor.preprocess(nb, {'metadata': {'path': src_dir}})
  ok = True
except nbconvert.preprocessors.CellExecutionError as e:
  print(e)
except TimeoutError as e:
  print(e)
except Exception as e:
  # catch-all
  print(e)

with open(dest, 'w', encoding='utf-8') as f:
  nbformat.write(nb, f)

args = ["jupyter", "nbconvert", "--to", "html", dest]
print(args)
p = subprocess.Popen(args)
p.wait()

# TODO: De-dup with downloads.exec
def exec(args, env = None, print_stdout=True):
    if env is None:
        env = os.environ.copy()
    print("running %s" % (args))
    r = subprocess.run(args, env=env, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    if print_stdout:
      print(r.stdout.decode())
    print(r.stderr.decode())
    if r.returncode != 0:
        r.check_returncode()
    return r.stdout.decode()


exec(["tar", "-zcvf", "/workspace/tree.tar.gz", "/workspace/artifacts"])
exec(["mv", "/workspace/tree.tar.gz", "/workspace/artifacts/tree.tar.gz"])

# TODO: We want to run a cleanup ... should this be here?


# TODO: We want to mark the build as failed, but if we do then we don't capture artifacts
# Maybe the build itself hasn't failed - it did run, there's no point (we think) in retrying it
#if not ok:
#  sys.exit(1)

# PYTHON HELPER to mark test failed
