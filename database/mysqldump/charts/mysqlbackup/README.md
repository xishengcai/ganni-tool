# MySQLBackup

## Introduction

This chart bootstraps a single node MySQL deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites


## Installing the Chart

To install the chart with the release name `my-release`:

```bash
$ helm install --name my-release stable/mysqlbackup
```

The command deploys MySQL on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

By default a random password will be generated for the root user. If you'd like to set your own password change the mysqlRootPassword
in the values.yaml.

You can retrieve your root password by running the following command. Make sure to replace [YOUR_RELEASE_NAME]:

    printf $(printf '\%o' `kubectl get secret [YOUR_RELEASE_NAME]-mysql -o jsonpath="{.data.mysql-root-password[*]}"`)

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```bash
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following tables lists the configurable parameters of the MySQL chart and their default values.

| Parameter         | Description                              | Default                                             |
|-------------------| ---------------------------------------- | --------------------------------------------------- |
| `schedule`        |                        | Most recent release                                 |
| `db_name_list`    |                         | `IfNotPresent`                                      |
| `source.host`     |             | `nil`                                               |
| `source.user`     |           | `nil`                                               |
| `source.password` |                | `nil`                                               |
| `target.host`     |          | `nil`                                               |
| `target.user`     |      |                        |

Some of the parameters above map to the env variables defined in the [MySQL DockerHub image](https://hub.docker.com/_/mysql/).

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```bash
$ helm install --name my-release \
  --set mysqlRootPassword=secretpassword,mysqlUser=my-user,mysqlPassword=my-password,mysqlDatabase=my-database \
    stable/mysql
```


```bash
$ helm install --name my-release -f values.yaml stable/mysql
```

> **Tip**: You can use the default [values.yaml](values.yaml)

## Persistence


## Custom MySQL configuration files


```yaml
configurationFiles:

```
