import os

import downloads
import teststate

def populate_state():
  if not teststate.get_state("spec.project"):
    p = os.environ.get("PROJECT_ID")
    if not p:
      p = current_project()
    teststate.set_state("spec.project", p)
  project = teststate.get_state("spec.project")

def current_project():
  stdout = downloads.exec(["gcloud", "config", "get-value", "project"], print_stdout=False, print_running=False)
  return stdout.strip()
