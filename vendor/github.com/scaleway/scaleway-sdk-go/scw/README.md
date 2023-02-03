# Scaleway config

## TL;DR

Recommended config file:

```yaml
# Get your credentials on https://console.scaleway.com/project/credentials
access_key: SCWXXXXXXXXXXXXXXXXX
secret_key: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
default_organization_id: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
default_project_id: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
default_region: fr-par
default_zone: fr-par-1
```

## Config file path

The function [`GetConfigPath`](https://godoc.org/github.com/scaleway/scaleway-sdk-go/scw#GetConfigPath) will try to locate the config file in the following ways:

1. Custom directory: `$SCW_CONFIG_PATH`
2. [XDG base directory](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html): `$XDG_CONFIG_HOME/scw/config.yaml`
3. Unix home directory: `$HOME/.config/scw/config.yaml`
4. Windows home directory: `%USERPROFILE%/.config/scw/config.yaml`

## Reading config order

[ClientOption](https://godoc.org/github.com/scaleway/scaleway-sdk-go/scw#ClientOption) ordering will decide the order in which the config should apply:

```go
p, _ := scw.MustLoadConfig().GetActiveProfile()

scw.NewClient(
    scw.WithProfile(p),                     // active profile applies first
    scw.WithEnv(),                          // existing env variables may overwrite active profile
    scw.WithDefaultRegion(scw.RegionFrPar)  // any prior region set will be discarded to usr the new one
)
```

## Environment variables

| Variable                       | Description                                                                                      | Legacy variables                                                                                              |
| :----------------------------- | :----------------------------------------------------------------------------------------------- | :------------------------------------------------------------------------------------------------------------ |
| `$SCW_ACCESS_KEY`              | Access key of a token ([get yours](https://console.scaleway.com/project/credentials))            | `$SCALEWAY_ACCESS_KEY` (used by terraform)                                                                    |
| `$SCW_SECRET_KEY`              | Secret key of a token ([get yours](https://console.scaleway.com/project/credentials))            | `$SCW_TOKEN` (used by cli), `$SCALEWAY_TOKEN` (used by terraform), `$SCALEWAY_ACCESS_KEY` (used by terraform) |
| `$SCW_DEFAULT_ORGANIZATION_ID` | Your default organization ID ([get yours](https://console.scaleway.com/project/credentials))     | `$SCW_ORGANIZATION` (used by cli),`$SCALEWAY_ORGANIZATION` (used by terraform)                                |
| `$SCW_DEFAULT_PROJECT_ID`      | Your default project ID ([get yours](https://console.scaleway.com/project/credentials))          |                                                                                                               |
| `$SCW_DEFAULT_REGION`          | Your default [region](https://developers.scaleway.com/en/quickstart/#region-and-zone)            | `$SCW_REGION` (used by cli),`$SCALEWAY_REGION` (used by terraform)                                            |
| `$SCW_DEFAULT_ZONE`            | Your default [availability zone](https://developers.scaleway.com/en/quickstart/#region-and-zone) | `$SCW_ZONE` (used by cli),`$SCALEWAY_ZONE` (used by terraform)                                                |
| `$SCW_API_URL`                 | Url of the API                                                                                   | -                                                                                                             |
| `$SCW_INSECURE`                | Set this to `true` to enable the insecure mode                                                   | `$SCW_TLSVERIFY` (inverse flag used by the cli)                                                               |
| `$SCW_PROFILE`                 | Set the config profile to use                                                                    | -                                                                                                             |
