# Data Collection CLI (datactl)

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-refresh-toc -->

**Table of Contents**

- [Data Collection CLI (datactl)](#red-hat-marketplace-control-cli-datactl)
- [Installation](#installation)
- [Usage](#usage)
- [Getting started](#getting-started)
- [Exporting from DataService sources](#exporting-from-dataservice-sources)
- [Exporting from IBM License Metric Tool sources](#exporting-from-ibm-license-metric-tool-sources)
- [Using the FIPS enabled datactl container](#using-the-fips-enabled-datactl-container)

<!-- markdown-toc end -->

## Installation

The tool is available via prebuilt executables on the [latest release](https://github.com/redhat-marketplace/datactl/releases/latest).
To install the tool to your local system, download the targz file and
extract it to a folder on your path.

## Usage

### As an OC Plugin

If the oc-datactl file is installed to your path. The oc command will recognize it as a plugin. You may
call `oc datactl`

### As a standalone tool

Datactl tool can be used standalone. Just move oc-datactl to your path and use `oc-datactl` directly.

## Getting started

1. Get your Red Hat Marketplace Pull Secret.

2. Log in to your cluster.

3. Setup your configuration. When prompted, provide the [pull secret token](https://marketplace.redhat.com/) as the `Upload API Secret`.

   ```sh
   oc datactl config init
   ```

   This will create the default configuration on your home directory. `~/.datactl/config`

4. Add the role-binding to the default service account on operator-namespace.

   Install the ClusterRole and create the ClusterRoleBinding for the default service account for the namespace the IBM Metrics Operator's 
   DataService is installed to `by default: redhat-marketplace`. The datactl tool will use this service account by default.

   ```sh
   oc apply -f resources/service-account-role.yaml // file found in release
   oc create clusterrolebinding rhm-files --clusterrole=rhm-files --serviceaccount=redhat-marketplace:default
   ```

5. Now you're configured. You can start using the export commands.

## Exporting from DataService sources

Recommended approach is to run the commands in this order:

```sh
// Must be logged in to the cluster

// Add the dataservice as a source, to which you are logged into with your current context
datactl sources add dataservice --use-default-context --insecure-skip-tls-verify=true --allow-self-signed=true --namespace=redhat-marketplace

// Pull the data from dataservice sources
oc datactl export pull --source-type=dataservice

// If you're connected to the internet
oc datactl export push

// If no errors from push.
oc datactl export commit
```

Let's break down what each one is doing.

`oc datactl sources add dataservice --use-default-context --insecure-skip-tls-verify=true --allow-self-signed=true --namespace=redhat-marketplace`

- Adds the default-context cluster's dataservice as a source for pulling
- Writes the source data-service-endpoint to `~/.datactl/config`
- The DataService to connect to is in the `redhat-marketplace` namespace

`oc datactl export pull`

- Pulls files from data service and stores them in a tar file under your `~/.datactl/data` folder.
- Writes the status of the files found in `~/.datactl/config`

`oc datactl export push`

- Files pulled by the previous command are pushed to Red Hat Marketplace.
- If this process errors, do not commit. Retry the export push or open a support ticket.

`oc datactl export commit`

- Commits the files to the dataservice.
- At this point you're telling the data service that you've retrieved these files and have or will submit them to Red Hat Marketplace.
- After some time, the files in dataservice will be cleaned up to save space.

If you want to transfer it somewhere else, you can find the tar file under your `~/.datactl/data/` directory.

## Exporting from IBM License Metric Tool sources

_Prerequisite_: API Token is required to get data from IBM License Metric Tool (ILMT). Login to your ILMT environment, go to _Profile_ and click _Show token_ under API Token section.

First step is to configure ILMT data source. Execute following command

`datactl sources add ilmt`

and provide ILMT hostname, port number and token

To pull data from ILMT, execute command

`datactl export pull --source-type=ilmt`

First time you will be asked to provide start date. Next time last synchronization date is stored in config file and will be updated to pull data from last synchronization date.

To push data to Red Hat Marketplace execute command

`datactl export push`

## Using the FIPS enabled datactl container

A containerized FIPS enabled version of datactl is provided, built with Red Hat's [go-toolset](https://developers.redhat.com/articles/2022/05/31/your-go-application-fips-compliant)


1. Create the `.datactl` directory locally on the host
   ```
   mkdir -p $HOME/.datactl
   ```
2. Setup your configuration, binding the `.datactl` and `.kube` directories, and providing the marketplace api endpoint and [pull secret token](https://marketplace.redhat.com/en-us/account/keys)
   ```
   docker run --rm \
   --mount type=bind,source=$HOME/.datactl,target=/root/.datactl \
   --mount type=bind,source=$HOME/.kube,target=/root/.kube \
   quay.io/rh-marketplace/datactl:latest config init --api marketplace.redhat.com --token ${TOKEN}
   ```
3. Add a data source, such as `dataservice` from your current OpenShift cluster context
   ```
   docker run --rm \
   --mount type=bind,source=$HOME/.datactl,target=/root/.datactl \
   --mount type=bind,source=$HOME/.kube,target=/root/.kube \
   quay.io/rh-marketplace/datactl:latest sources add dataservice --insecure-skip-tls-verify=true --use-default-context --allow-self-signed=true --namespace=redhat-marketplace
   ```
4. Pull from the data source
   ```
   docker run --rm \
   --name datactl \
   --mount type=bind,source=$HOME/.datactl,target=/root/.datactl \
   --mount type=bind,source=$HOME/.kube,target=/root/.kube \
   quay.io/rh-marketplace/datactl:latest export pull
   ```
5. Push data to Red Hat Marketplace
   ```
   docker run --rm \
   --name datactl \
   --mount type=bind,source=$HOME/.datactl,target=/root/.datactl \
   --mount type=bind,source=$HOME/.kube,target=/root/.kube \
   quay.io/rh-marketplace/datactl:latest export push
   ```
6. Commit
   ```
   docker run --rm \
   --name datactl \
   --mount type=bind,source=$HOME/.datactl,target=/root/.datactl \
   --mount type=bind,source=$HOME/.kube,target=/root/.kube \
   quay.io/rh-marketplace/datactl:latest export commit
  ```
