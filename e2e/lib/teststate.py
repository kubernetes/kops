import copy
import datetime
import json
import os
import os.path

import e2e

if not '_state' in globals():
  _state = {}

def timestamp():
  now = datetime.datetime.utcnow()
  now = now.replace(microsecond=0)
  return now.isoformat()

# init_state sets our global to the same value as the workbook state value
# it also populates some initial values
def init_state(s):
  global _state
  seedState = copy.deepcopy(s)
  _state = s
  _state["apiVersion"] = "e2e.kops.k8s.io/v1alpha1"
  _state["kind"] = "TestRun"
  _state["seedState"] = seedState
  if not "spec" in _state:
    _state["spec"] = {}
  if not "status" in _state:
    _state["status"] = {}
  if not "results" in _state["status"]:
    _state["status"]["results"] = []
  _state["status"]["started"] = timestamp()
  save_state()
  return _state

def finished():
  _state["status"]["finished"] = timestamp()
  save_state()

def save_state():
  global _state
  p = os.path.join(e2e.artifacts_dir(), "state.json")
  with open(p, 'w') as f:
    json.dump(_state, f, sort_keys=True, indent=4, separators=(',', ': '))

#def _load_state():
#  p = os.path.join(e2e.artifacts_dir(), "state.json")
#  if os.path.exists(p):
#    with open(p) as f:
#      state = json.load(f)
#  return state

def update_state(add):
  global _state
  m = _merge_dicts(_state, add)
  for k, v in m.items():
    _state[k] = v
  save_state()

def append_state(path, o):
  global _state
  a = get_or_set_state(path, [])
  a.append(o)
  set_state(path, a)
  save_state()

def failure(message, **kwargs):
  global _state
  print(f"failure: {message} {kwargs}")
  append_state("status.failures", message)

def _merge_dicts(l, r):
  m = {**l, **r}
  for k, v in m.items():
    if k in l and k in r:
      if isinstance(r[k],dict):
        m[k] = _merge_dicts(l[k], r[k])
      else:
        m[k] = r[k]
  return m

def get_state(path):
  global _state
  current = _state
  for token in path.split("."):
    current = current.get(token)
    if not current:
      return None
  return current

def set_state(path, v):
  global _state
  current = _state
  tokens = path.split(".")
  for token in tokens[:-1]:
    next = current.get(token)
    if not next:
      next = {}
      current[token] = next
    current = next
  current[tokens[-1]] = v
  return v

def get_or_set_state(path, v):
  global _state
  existing = get_state(path)
  if not existing:
    return set_state(path, v)
  else:
    return existing
