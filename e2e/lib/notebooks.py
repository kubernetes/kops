import subprocess, os, json
from os import path
import datetime
import nbformat
import nbconvert

import downloads
import e2e
import teststate

if not 'workbook_dir' in globals():
  workbook_dir = os.getcwd()

def run(src, destdir=None, timeout=3600, parameters=None):
  result = {
    "src": src,
    "parameters": parameters,
  }

  if not destdir:
    ad = e2e.artifacts_dir()
    destdir = os.path.join(ad, src)

  src = src + ".ipynb"

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

  clear_output = nbconvert.preprocessors.ClearOutputPreprocessor()

  ep = nbconvert.preprocessors.ExecutePreprocessor(timeout=timeout, kernel_name='python3')

  prior_env_artifacts = os.environ.get("ARTIFACTS")
  os.environ["ARTIFACTS"] = destdir

  ok = False
  try:
    clear_output.preprocess(nb, {'metadata': {'path': src_dir}})
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

  state_file = os.path.join(destdir, "state.json")
  if os.path.exists(state_file):
    with open(state_file) as f:
      child_state = json.load(f)
    result["state"] = child_state

  result["success"] = ok

  teststate.append_state("status.subtasks", result)

  # Convert output to HTML for easier reading
  args = ["jupyter", "nbconvert", "--to", "html", dest]
  downloads.exec(args)
  
  return dest


