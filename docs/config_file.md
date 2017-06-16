# Kops config file

Kops will look in the following locations for a config file, in order of priority:
  - ./
  - ~/.kops.{file type}
  - ~/.config/kops/.kops.{file type}

The supported file types are: json, yaml, toml, hcl and .properties

An example of a minimal config file would be:

```yaml
state: s3://my-s3-bucket
```
