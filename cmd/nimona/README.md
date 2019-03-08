## Usage

### Daemon

Start and configure a local peer.

```
nimona daemon init
nimona daemon start
```

### Client

Run commands against a daemon

```
nimona object
nimona help
```

### Provision

You can also install the daemon in a supported provider.

```
nimona provision --platform do --token <> --ssh-fingerprint <> --hostname <>
```

#### Supported Flags

* **--platform** the provider to be used for the deployment
* **--hostname** the hostname that nimona will use, if defined the dns will also be updated
* **--token** the access token required to authenticate with the provider
* **--ssh-fingerprint** the ssh fingerprint for the key that will be added to the server (needs to exist in the provider)
* **--size** size of the server, default for DO *s-1vcpu-1gb*
* **--region** region that the server will be deployed, default *lon1*

#### Suppored Providers

* do - DigitalOcean
