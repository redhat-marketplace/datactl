## datactl config init

Initializes the config for Dataservice and API endpoints

### Synopsis

Configures the default config file ('$HOME/.datactl/config') with details about the cluster. The command will attempt to resolve the Dataservice URL. It will also prompt for the Upload API endpoint and secret if they are not provided by flags.

 If you are attempting to configure a host machine to upload payloads only; the --config-api-only flag is provided to prevent kubernetes resources from being queried. This is to prevent unnecessary errors with lack of access.

```
datactl config init
```

### Examples

```
  # Initialize the config, prompting for API and Token values.
  datactl config init
  
  # Initialize the config and preset upload URL and secret. Will not prompt.
  datactl config init --api marketplace.redhat.com --token MY_TOKEN
  
  # Initialize only the API config, prompting for API and Token values.
  datactl config init --api-only
  
  # Initialize the config, force resetting of values if they are already set.
  datactl config init --force
```

### Options

```
      --allow-non-system-ca   allows non system CA certificates to be added to the dataService config
      --allow-self-signed     allows self-signed certificates to be added to the dataService configs
      --api string            upload endpoint
      --config-api-only       only configure Upload API components
      --force                 force configuration updates and prompts
  -h, --help                  help for init
```

### Options inherited from parent commands

```
      --as string                      Username to impersonate for the operation. User could be a regular user or a service account in a namespace.
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --cache-dir string               Default cache directory (default "$HOME/.kube/cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
      --match-server-version           Require server version to match client version
  -n, --namespace string               If present, the namespace scope for this CLI request
      --password string                Password for basic authentication to the API server
      --profile string                 Name of profile to capture. One of (none|cpu|heap|goroutine|threadcreate|block|mutex) (default "none")
      --profile-output string          Name of the file to write the profile to (default "profile.pprof")
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
      --rhm-config string              override the rhm config file
      --rhm-upload-api-host string     Override the Marketplace API host
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
      --username string                Username for basic authentication to the API server
      --warnings-as-errors             Treat warnings received from the server as errors and exit with a non-zero exit code
```

### SEE ALSO

* [datactl config](datactl_config.md)	 - Modify datactl files

###### Auto generated by spf13/cobra on 24-Nov-2021