import subprocess, os, json
from os import path
import datetime
import nbformat
import nbconvert

import downloads
import e2e

if not 'workbook_dir' in globals():
  workbook_dir = os.getcwd()

def run_notebook(src, destdir=None, timeout=3600, parameters=None):
  src = src + ".ipynb"

  if not destdir:
    ad = e2e.artifacts_dir()
    destdir = os.path.join(ad, src)

  os.makedirs(destdir, exist_ok=True)
  dest = os.path.join(destdir, "output.ipynb")
  
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

  ep = nbconvert.preprocessors.ExecutePreprocessor(timeout=timeout, kernel_name='python3')

  prior_env_artifacts = os.environ.get("ARTIFACTS")
  os.environ["ARTIFACTS"] = destdir

  ok = False
  try:
    ep.preprocess(nb, {'metadata': {'path': src_dir}})
    ok = True
  except nbconvert.preprocessors.CellExecutionError as e:
    print(e)
  except TimeoutError as e:
    print(e)
  except Exception as e:
    # catch-all
    print(e)

  os.environ["ARTIFACTS"] = prior_env_artifacts or ""
  
  with open(dest, 'w', encoding='utf-8') as f:
    nbformat.write(nb, f)

  #args = ["jupyter", "nbconvert", "--to", "notebook"]
  #args += ["--execute",os.path.abspath(dest)]
  #args += ["--output", os.path.abspath(dest)]
  #args += ["--ExecutePreprocessor.timeout=%s" % timeout]

  #downloads.exec(args, env=env)
  
  return dest


