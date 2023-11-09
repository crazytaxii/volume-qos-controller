# volume-qos-controller

A Kubernetes controller used to configure PV QoS (Supporting Ceph RBD volumes now only).

**Ceph RBD CSI Driver must [set mounter as rbd-nbd](https://github.com/ceph/ceph-csi/blob/devel/docs/rbd-nbd.md#configuration)!**

## Deploying

Configure parameters in [manifests/qos-controller.yaml](./manifests/qos-controller.yaml):

- monitors
- user (admin is suggested)
- key

Lookup this configuration via:

```bash
$ ceph mon dump
$ ceph auth get client.admin
```

## Using

1. Create a PVC

    ```yaml
    ---
    apiVersion: v1
    kind: PersistentVolumeClaim
    metadata:
    name: datavol
    namespace: demo
    spec:
    accessModes:
        - ReadWriteMany
    volumeMode: Block
    resources:
        requests:
        storage: 5Gi
    ```

1. Update PVC with annotation of QoS rule, or just create the PVC with it:

    ```bash
    $ kubectl patch pvc datavol -n demo --type=merge -p '{"metadata":{"annotations": {"pv.kubernetes.io/qos-bps-limit": "10M"}}}'
    persistentvolumeclaim/datavol patched
    ```

### QoS Rules

1. IOPS class
    - pv.kubernetes.io/qos-iops-limit
    - pv.kubernetes.io/qos-read-iops-limit
    - pv.kubernetes.io/qos-write-iops-limit
    - pv.kubernetes.io/qos-iops-burst
    - pv.kubernetes.io/qos-read-iops-burst
    - pv.kubernetes.io/qos-write-iops-burst
1. BPS class
    - pv.kubernetes.io/qos-bps-limit
    - pv.kubernetes.io/qos-read-bps-limit
    - pv.kubernetes.io/qos-write-bps-limit
    - pv.kubernetes.io/qos-bps-burst
    - pv.kubernetes.io/qos-read-bps-burst
    - pv.kubernetes.io/qos-write-bps-burst

Setting multiple QoS rules simultaneously is supported, the value must follow `^[1-9][0-9]*(M|G|T)?$` regular expression:

- 1
- 10
- 100
- 1M
- 10M
- 1G

## Developing

How to build binary:

```bash
$ make build
```

How to build image:

```bash
$ make image
```

## Implementation

We need to set different QoS rule to different PVC. Considering the design in this proposal <https://docs.google.com/document/d/1x19sofRvRNmQ15E0pj8OygSevkTVhtv_H8zDJ9CqERM/edit>, we choose to put the QoS configuration in PVC's annotations:

```yaml
kind: PersistentVolumeClaim
metadata:
  annotations:
    ...
    pv.kubernetes.io/iops-limit: "1000"
    pv.kubernetes.io/bps-limit: "100Mi"
```
