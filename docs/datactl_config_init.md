## datactl config init

Initializes the config for Dataservice and API endpoints

### Synopsis

Configures the default config file ('$HOME/.datactl/config') with details about the cluster. It will also prompt for the Upload API endpoint and secret if they are not provided by flags.

```
datactl config init
```

### Examples

```
  # Initialize the config, prompting for API and Token values.
  datactl config init
  
  # Initialize the config and preset upload URL and secret. Will not prompt.
  datactl config init --api swc.saas.ibm.com --token MY_TOKEN
```

### Options

```
      --api string     upload endpoint
  -h, --help           help for init
      --token string   upload api secret
```

### Options inherited from parent commands

```
      --add-dir-header                   If true, adds the file directory to the header of the log messages
      --alsologtostderr                  log to standard error as well as files (no effect when -logtostderr=true)
      --as string                        Username to impersonate for the operation. User could be a regular user or a service account in a namespace.
      --as-group stringArray             Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --as-uid string                    UID to impersonate for the operation.
      --cache-dir string                 Default cache directory (default "/home/user/.kube/cache")
      --certificate-authority string     Path to a cert file for the certificate authority
      --client-certificate string        Path to a client certificate file for TLS
      --client-key string                Path to a client key file for TLS
      --cluster string                   The name of the kubeconfig cluster to use
      --context string                   The name of the kubeconfig context to use
      --disable-compression              If true, opt-out of response compression for all requests to the server
      --insecure-skip-tls-verify         If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string                Path to the kubeconfig file to use for CLI requests.
      --log-backtrace-at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log-dir string                   If non-empty, write log files in this directory (no effect when -logtostderr=true)
      --log-file string                  If non-empty, use this log file (no effect when -logtostderr=true)
      --log-file-max-size uint           Defines the maximum size a log file can grow to (no effect when -logtostderr=true). Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
      --logtostderr                      log to standard error instead of files (default true)
      --match-server-version             Require server version to match client version
  -n, --namespace string                 If present, the namespace scope for this CLI request
      --no-color                         no color on CLI output
      --one-output                       If true, only write logs to their native severity level (vs also writing to each lower severity level; no effect when -logtostderr=true)
      --password string                  Password for basic authentication to the API server
      --profile string                   Name of profile to capture. One of (none|cpu|heap|goroutine|threadcreate|block|mutex) (default "none")
      --profile-output string            Name of the file to write the profile to (default "profile.pprof")
      --request-timeout string           The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
      --rhm-config string                override the rhm config file
      --rhm-upload-api-host string       Override the Marketplace API host
  -s, --server string                    The address and port of the Kubernetes API server
      --skip-headers                     If true, avoid header prefixes in the log messages
      --skip-log-headers                 If true, avoid headers when opening log files (no effect when -logtostderr=true)
      --stderrthreshold severity         logs at or above this threshold go to stderr when writing to files and stderr (no effect when -logtostderr=true or -alsologtostderr=false) (default 2)
      --tls-cipher-suites strings        Comma-separated list of cipher suites for the server. Values are from tls package constants (https://golang.org/pkg/crypto/tls/#pkg-constants). If omitted, a subset will be used (default [TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384])
      --tls-min-version string           Minimum TLS version supported. Value must match version names from https://golang.org/pkg/crypto/tls/#pkg-constants. (default "VersionTLS12")
      --tls-server-name string           Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --user string                      The name of the kubeconfig user to use
      --username string                  Username for basic authentication to the API server
  -v, --v Level                          number for the log level verbosity
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
      --warnings-as-errors               Treat warnings received from the server as errors and exit with a non-zero exit code
```

### SEE ALSO

* [datactl config](datactl_config.md)	 - Modify datactl configuration

###### Auto generated by spf13/cobra on 24-Oct-2023
