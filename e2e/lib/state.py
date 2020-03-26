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
  
def init_state(s):
  global _state
  _state = s
  update_state({"schema": "e2e.kops.k8s.io/v1"})
  if not "results" in _state:
    update_state({"results": [] })
  _state["started"] = timestamp()
  save_state()
  return _state

def finished():
  _state["finished"] = timestamp()
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

def failure(message, **kwargs):
  global _state
  if not "failures" in _state:
    _state["failures"] = []
  print(f"failure: {message} {kwargs}")
  _state["failures"].append(message)
  save_state()

def _merge_dicts(l, r):
  m = {**l, **r}
  for k, v in m.items():
    if k in l and k in r:
      if isinstance(r[k],dict):
        m[k] = _merge_dicts(l[k], r[k])
      else:
        m[k] = r[k]
  return m
