# data-sync-operator
A Kubernetes Operator designed to handle the lifecycle around "Workspaces". Name change pending


## Description
This project is currently in active development and should be considered pre-alpha in its current state.

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## Getting Started

### Prerequisites
- go version v1.25.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.33.3+ cluster.
- kind version v0.29.0+
- kubebuilder 4.10.1+

### Install Kubebuilder

# Download the latest release
`curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)"`

# Make it executable
```bash
chmod +x kubebuilder

sudo mv kubebuilder /usr/local/bin/
```

### Local Development

To develop locally the following steps must be taken to run the application.


#### Init the cluster

To init the cluster please run

```bash
make setup-test-e2e
```

This command will stand up a kind cluster locally. The cluster will have all all dependencies and our crd installed.

> **NOTE**: Please ensure you are using the correct context before running the below commands. We are installing stuff into the cluster.

Next run the below command to install our CRD onto your cluster.

```bash
make install
```

##### Running the operator locally.
To run locally ensure you have [air](https://github.com/air-verse/air) installed.

```bash
go install github.com/air-verse/air@latest
```

Once installed use the below make command to run the operator with hot reload enabled

> [!NOTE]
> While the `make dev` command will regenerate the manifests should changes occur you will need to manually install them.

```bash
make dev
```


##### Making Changes to the shape of our CRDs.

If you need to make changes to the shape of a CRD you will need to regenerate manifests and code created via kubebuilder.

The below command will refresh the crd and reinstall it.

```bash
make regenerate-crd
```


## Project Distribution

Following the options to release and provide this solution to the users.

### Helm Chart

1. Build the chart using the optional helm plugin

```sh
make generate-chart
```

2. See that a chart was generated under './chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Please be aware
any custom changes made prior will need to be reapplied



