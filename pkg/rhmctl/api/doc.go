//go:generate controller-gen object paths=./...
//go:generate conversion-gen --input-dirs "./v1" -O "zz_generated.conversion" --skip-unsafe=true -v 5 -o "."

// +k8s:deepcopy-gen=package

package api
