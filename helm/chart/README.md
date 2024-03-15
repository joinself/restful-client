# Self RESTful Client Helm Chart

This chart bootstraps a deployment of Self RESTful client on Kubernetes using the Heml package manager.

## Prerequisites
- Kubernetes 1.23+
- Helm 3.8.0+

## Install the Chart

Add the required Kubernetes secrets:
```yaml
# restful-client-secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: restful-client
stringData:
  auth-username: "self"
  auth-password: "secret"
  storage-key: "secret"
  jwt-signing-key: "secret"
  app-id: "c4f81d86-9dac-40fd-9830-13c66a0b2345"
  app-secret: "sk_1:56qJGhYCJmTHsYChCp3sPSjmiGlN2yG0KakYDquMAD0"
  app-environment: "sandbox"
```
> Note: Replace the values above with those specific to your environment.

```bash
kubectl apply -f restful-client-secrets.yaml
```

Install the chart:
```bash
helm install restful-client oci://ghcr.io/joinself/charts/restful-client --set existingSecret=restful-client
```

## Uninstall the Chart

To uninstall/delete the deployment:
```bash
helm uninstall restful-client
```

This removes all the Kubernetes components but the PVC's associated with the deployment still exist.

To remove the PVC's associated with the deployment run:
```bash
kubectl delete pvc -l app.kubernetes.io/name=restful-client
```
> Note: Deleting the PVC's will delete all persistant data associated with the deployment.

## Parameters

| Name | Description | Value |
|------|-------------|-------|
| image.registry | restful-client image registry | ghcr.io |
| image.repository | restful-client image repository | joinself/restful-client |
| image.tag | restful-client image tag | latest |
| image.pullPolicy | Kubernetes image pull policy | IfNotPresent |
| authUsername | Username for accessing restful-client | "" |
| authPassword | Password for accessing restful-client | "" |
| storageDir | Path to persistant storage | "/data" |
| storageKey | Storage encryption key for persistant data | "" |
| jwtSigningKey | Key used to sign JSON Web Tokens | "" |
| appID | Self application ID | "" |
| appSecret | Self application secret | "" |
| appEnvironment | Self application target environment (sandbox, production) | "" |
| existingSecret | Name of existing secret to use for restful-client secrets. If defined `authUsername`, `authPassword`, `storageKey`, `jwtSigningKey`, `appID`, `appSecret`, `appEnvironment` are all ignored. | "" |
| secretKeys.authUsername | Name of key in existing secret to use for `authUsername`. Only used if `existingSecret` if defined. | auth-username |
| secretKeys.authPassword | Name of key in existing secret to use for `authPassword`. Only used if `existingSecret` if defined. | auth-password |
| secretKeys.storageKey | Name of key in existing secret to use for `storageKey`. Only used if `existingSecret` if defined. | storage-key |
| secretKeys.jwtSigningKey | Name of key in existing secret to use for `jwtSigningKey`. Only used if `existingSecret` if defined. | jwt-signing-key |
| secretKeys.appID | Name of key in existing secret to use for `appID`. Only used if `existingSecret` if defined. | app-id |
| secretKeys.appSecret | Name of key in existing secret to use for `appSecret`. Only used if `existingSecret` if defined. | app-secret |
| secretKeys.appEnvironment | Name of key in existing secret to use for `appEnvironment`. Only used if `existingSecret` if defined. | app-environment |
| service.type | Kubernetes service type | CkusterIP |
| service.port | Restful-client service port | 8080 |
| resources | Set container requests and limits for different resources e.g. CPU, memory. | {} |
| persistance.enabled | Enable data persistence using a PVC | true |
| persistence.mountPath | Path to mount persistant volume | /data |
| persistence.accessModes | Access mode for persistant volume | ["ReadWriteOnce"] |
| persistence.size | Size of persistant volume | 8Gi |
| persistence.storageClass | Storage class for persistant volume | "" |
