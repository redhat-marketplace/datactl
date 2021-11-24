## datactl export pull

Pulls files from RHM Operator

### Synopsis

Pulls files from the Dataservice on the cluster.

 Prints a table of the files pulled with basic information. The --before or --after flags can be used to change the date range that the files are pulled from. All dates must be in RFC3339 format as defined by the Golang time package.

 If the files have already been pulled then using the --include-deleted flag may be necessary.

```
datactl export pull [(--before DATE) (--after DATE) (--include-deleted)]
```

### Examples

```
  # Pull all available files from the current dataservice cluster to Usage
  datactl export pull
  
  # Pull all files before November 14th, 2021
  datactl export pull --before 2021-11-15T00:00:00Z
  
  # Pull all files after November 14th, 2021
  datactl export pull
  
  # Pull all files between November 14th, 2021 and November 15th, 2021
  datactl export pull --after 2021-11-14T00:00:00Z --before 2021-11-15T00:00:00Z
  
  # Pull all deleted files
  datactl export pull --include-deleted
```

### Options

```
      --after string                  pull files after date
      --allow-missing-template-keys   If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
      --before string                 pull files before date
  -h, --help                          help for pull
      --include-deleted               include deleted files
      --no-headers                    When using the default or custom-column output format, don't print headers (default print headers).
  -o, --output string                 Output format. One of: json|yaml|name|go-template|go-template-file|template|templatefile|jsonpath|jsonpath-as-json|jsonpath-file|custom-columns|custom-columns-file|wide See custom columns [https://kubernetes.io/docs/reference/kubectl/overview/#custom-columns], golang template [http://golang.org/pkg/text/template/#pkg-overview] and jsonpath template [https://kubernetes.io/docs/reference/kubectl/jsonpath/].
      --template string               Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].
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

* [datactl export](datactl_export.md)	 - Export metrics from RHM Operator

###### Auto generated by spf13/cobra on 24-Nov-2021
