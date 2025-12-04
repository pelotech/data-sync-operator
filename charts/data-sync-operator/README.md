# data-sync-operator

### version: 0.1.1<!-- x-release-please-version -->

### appVersion: 0.1.1 <!-- x-release-please-version -->

A Helm chart to distribute the project data-sync-operator

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| certManager.enable | bool | `false` |  |
| crd.enable | bool | `true` |  |
| crd.keep | bool | `true` |  |
| manager.args[0] | string | `"--leader-elect"` |  |
| manager.env | list | `[]` |  |
| manager.image.pullPolicy | string | `"IfNotPresent"` |  |
| manager.image.repository | string | `"ghcr.io/pelotech/data-sync-operator"` |  |
| manager.image.tag | string | `"latest"` |  |
| manager.imagePullSecrets | list | `[]` |  |
| manager.podSecurityContext.runAsNonRoot | bool | `true` |  |
| manager.podSecurityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| manager.replicas | int | `1` |  |
| manager.resources.limits.cpu | string | `"500m"` |  |
| manager.resources.limits.memory | string | `"128Mi"` |  |
| manager.resources.requests.cpu | string | `"10m"` |  |
| manager.resources.requests.memory | string | `"64Mi"` |  |
| manager.securityContext.allowPrivilegeEscalation | bool | `false` |  |
| manager.securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| metrics.enable | bool | `false` |  |
| metrics.port | int | `8443` |  |
| rbacHelpers.enable | bool | `false` |  |

