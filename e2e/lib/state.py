import os
import os.path
import json

import e2e

def save_state(state):
  p = os.path.join(e2e.artifacts_dir(), "state.json")
  with open(p, 'w') as f:
    json.dump(state, f, sort_keys=True, indent=4, separators=(',', ': '))

def load_state():
  p = os.path.join(e2e.artifacts_dir(), "state.json")
  state = {}
  if os.path.exists(p):
    with open(p) as f:
      state = json.load(f)
  return state

def update_state(state, add):
  m = merge_dicts(state, add)
  for k, v in m.items():
    state[k] =v

def merge_dicts(l, r):
  m = {**l, **r}
  for k, v in m.items():
    if k in l and k in r:
      if isinstance(r[k],dict):
        m[k] = merge_dicts(l[k], r[k])
      else:
        m[k] = r[k]
  return m
