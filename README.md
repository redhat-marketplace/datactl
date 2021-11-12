# Red Hat Marketplace Control CLI (rhmctl)

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-refresh-toc -->

**Table of Contents**

- [Red Hat Marketplace Control CLI (rhmctl)](#red-hat-marketplace-control-cli-rhmctl)
- [Installation](#installation)
- [Usage](#usage)
- [Getting started](#getting-started)

<!-- markdown-toc end -->

## Installation

The tool is available via prebuilt executables on the [latest release](https://github.com/redhat-marketplace/rhmctl/releases/latest).
To install the tool to your local system, download the targz file and
extract it to a folder on your path.

```sh
# Substitute BIN for your bin directory.
# Substitute VERSION for the current released version.
# Substitute BINARY_NAME for rhmctl or oc-rhmctl.
BIN="/usr/local/bin" && \
VERSION="0.2.0" && \
BINARY_NAME="oc-rhmctl" && \
  curl -sSL \
    "https://github.com/bufbuild/buf/releases/download/v${VERSION}/${BINARY_NAME}-$(uname -s)-$(uname -m)" \
    -o "${BIN}/${BINARY_NAME}" && \
  chmod +x "${BIN}/${BINARY_NAME}"
```

## Usage

RhmCtl tool is an OpenShift CLI plugin. You can call it via `oc rhmctl`. Or you can call
the tool independently via

## Getting started

1. Setup your configuration.

   ```sh
   oc rhmctl config init
   ```

   This will create the default configuration on your home directory. `~/.rhmctl/config`

2. Get your Red Hat Marketplace Pull Secret.

3. Log in to your cluster.

4. Update your configuration with some important information.

   ```yaml
   data-service-endpoints:
     - cluster-context-name: your-current-context // Required, your kubectl context
       url: https://your-data-service-route       // RHM Data-Service route
       token: /your/token/file                    // *Full path to a token file
       token-data: token-data-as-text             // *Token data as a string
   marketplace:
     host: https://marketplace.redhat.com         // or https://sandbox.marketplace.redhat.com
     pull-secret: /your/pullsecret/file           // **Full path to the RHM pull secret file
     pull-secret-data: pullsecret-data-as-text    // **RHM pull secret as a string
   ```

   **Notes**

   - One of token or token-data is required
   - One of pull-secret or pull-secret-data is required.

   **Help**

   This script can some of the values for you.

   ```sh
   echo "data-service-url = https://$(oc get route -n openshift-redhat-marketplace rhm-data-service -o go-template='{{.spec.host}}')"
   echo "cluster-context-name = $(oc config current-context)"
   ```

5. Generate a token the RHM data-service.

   Install the role and role binding for the default service account for the `openshift-redhat-marketplace`
   namespace. Then you can get a valid token to the data-service.

   ```sh
   oc apply -f token-job.yaml // file found in release
   ```

   You can fetch the token by looking at the job logs.

   ```sh
   oc get pods -n openshift-redhat-marketplace
     redhat-marketplace-controller-manager-7f99cfbcbd-wdwhf   2/2     Running     0          3h36m
     rhm-data-service-0                                       4/4     Running     0          6h15m
     rhm-data-service-1                                       4/4     Running     0          6h13m
     rhm-data-service-2                                       4/4     Running     1          6h11m
     rhm-metric-state-5fbf6bd558-fjlmh                        4/4     Running     0          6h15m
     rhm-token-job-gx9gv                                      0/1     Completed   0          54s
   oc log -n openshift-redhat-marketplace rhm-token-job-gx9gv
   ```

   Easiest solution is to copy the log to a file and reference that in the `data-service-endpoints[].token` field.

6. Now you're configured. You can start using the export commands.

## Export

Recommended approach is to run the commands in this order:

```sh
oc rhmctl export pull
oc rhmctl export commit
oc rhmctl export push
```

Let's break down what each one is doing.

`oc rhmctl export pull`

- Pulls files from data service and stores them in a tar file under your `~/.rhmctl/data` folder.
- Writes the status of the files found in `~/.rhmctl/config`

`oc rhmctl export commit`

- Commits the files to the dataservice.
- At this point you're telling the data service that you've retrieved these files and will submit them to Red Hat Marketplace.
- After some time, the files in dataservice will be cleaned up to save space.

`oc rhmctl export push`

- Pushes the files pulled to Red Hat Marketplace.

If you want to transfer it somewhere else, you can find the tar file under your `~/.rhmctl/data/` directory.
