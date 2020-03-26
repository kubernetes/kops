#!/usr/bin/env python3

import datetime
import json
import os
import subprocess
import sys

import nbformat
import nbconvert

if len(sys.argv) <= 1:
  sys.exit("args: <workbook> [<cleanup-workbook>]")

src = sys.argv[1]
cleanup_workbook = None
if len(sys.argv) >= 3:
  cleanup_workbook = sys.argv[2]
destdir = "/workspace/artifacts"

timeout = 3600

def run_workbook(src, destdir, parameters=None):
  now = datetime.datetime.utcnow()
  now = now.replace(microsecond=0)
  timestamp = now.strftime("%Y%m%d%H%M%s")

  if not src.endswith(".ipynb"):
    src = src + ".ipynb"

  dest = os.path.join(destdir, timestamp + "_" + src.replace('/', '_'))
  os.makedirs(os.path.abspath(os.path.dirname(dest)), exist_ok=True)

  src_dir = os.path.dirname(src)

  with open(src) as f:
    nb = nbformat.read(f, as_version=4)

  if parameters:
    code = "# Parameters provided during execution"
    for k, v in parameters.items():
      if isinstance(v, dict):
        code += '\n%s = %s' % (k, v)
      else:
        code += '\n%s = "%s"' % (k, v)

    first_cell = nb.cells[0]
    first_cell.source = code

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

  # Write output
  with open(dest, 'w', encoding='utf-8') as f:
    nbformat.write(nb, f)

  # Convert output to HTML for easier reading
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

parameters = None
if os.environ.get("PARAMETERS"):
  parameters = json.loads(os.environ["PARAMETERS"])

run_workbook(src, destdir, parameters=parameters)

if cleanup_workbook:
  prior_artifacts = os.environ.get("ARTIFACTS")
  os.environ["ARTIFACTS"] = os.path.join(destdir, "cleanup")
  run_workbook(cleanup_workbook, os.path.join(destdir, "cleanup"), parameters=parameters)
  os.environ["ARTIFACTS"] = prior_artifacts or ""

exec(["tar", "-zcvf", "/workspace/tree.tar.gz", "/workspace/artifacts"])
exec(["mv", "/workspace/tree.tar.gz", "/workspace/artifacts/tree.tar.gz"])

# TODO: We want to mark the build as failed, but if we do then we don't capture artifacts
# Maybe the build itself hasn't failed - it did run, there's no point (we think) in retrying it
#if not ok:
#  sys.exit(1)

# PYTHON HELPER to mark test failed
