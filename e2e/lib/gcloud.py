import downloads

def current_project():
  stdout = downloads.exec(["gcloud", "config", "get-value", "project"])
  return stdout.strip()
