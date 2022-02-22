# Data Collection CLI (datactl)

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-refresh-toc -->

**Table of Contents**

- [Data Collection CLI (datactl)](#red-hat-marketplace-control-cli-datactl)
- [Installation](#installation)
- [Usage](#usage)
- [Getting started](#getting-started)
- [Export](#export)

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

3. Setup your configuration.

   ```sh
   oc datactl config init
   ```
   *Note*: 
   If you are using an IBM Entitlement Key to authenticate to Redhat Marketplace use the token in the password field of your entitlement key as API Secret when prompted.
   ```sh
   oc project openshift-redhat-marketplace
   kubectl get secret ibm-entitlement-key -o go-template='{{ $dc := index .data ".dockerconfigjson"|base64decode }}{{$dc}}' | grep -e "password"
   ```

   Or using `jq`
   For a prod entitlement key:
   ```sh
   oc project openshift-redhat-marketplace
   kubectl get secret ibm-entitlement-key -ojson | jq '.data | map_values(@base64d) .".dockerconfigjson" | fromjson.auths."cp.icr.io".password' 
   ```

   For a stage entitlement key:
   ```sh
   oc project openshift-redhat-marketplace
   kubectl get secret ibm-entitlement-key -ojson | jq '.data | map_values(@base64d) .".dockerconfigjson" | fromjson.auths."stg.icr.io".password' 
   ```

   This will create the default configuration on your home directory. `~/.datactl/config`

4. Add the role-binding to the default service account on operator-namespace.

   Install the role and role binding for the default service account for the `openshift-redhat-marketplace`
   namespace. The datactl tool will use these by default.

   ```sh
   oc apply -f resources/service-account-role.yaml // file found in release
   ```

5. Now you're configured. You can start using the export commands.

## Export

Recommended approach is to run the commands in this order:

```sh
// Must be logged in to the cluster
oc datactl export pull

// If you're in a connect
oc datactl export push

// If no errors from push.
oc datactl export commit
```

Let's break down what each one is doing.

`oc datactl export pull`

- Pulls files from data service and stores them in a tar file under your `~/.datactl/data` folder.
- Writes the status of the files found in `~/.datactl/config`

`oc datactl export push`

- Pushes the files pulled to Red Hat Marketplace.
- If this process errors, do not commit. Retry the export push or open a support ticket.

`oc datactl export commit`

- Commits the files to the dataservice.
- At this point you're telling the data service that you've retrieved these files and have or will submit them to Red Hat Marketplace.
- After some time, the files in dataservice will be cleaned up to save space.

If you want to transfer it somewhere else, you can find the tar file under your `~/.datactl/data/` directory.
