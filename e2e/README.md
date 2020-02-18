To sanitize:

jupyter nbconvert tests/smoketest.ipynb --to notebook --ClearOutputPreprocessor.enabled=True --output smoketest.ipynb
