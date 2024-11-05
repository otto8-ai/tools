# gptscript-credential-database

This is a set of GPTScript [credential helpers](https://docs.gptscript.ai/credentials) for databases.
Currently, SQLite and PostgreSQL are supported.
To use SQLite, set your GPTScript configuration to use `sqlite` as the credential store.
To use PostgreSQL, set your GPTScript configuration to use `postgres` as the credential store.

By default, all credentials are stored **unencrypted**.

Only macOS and Linux are supported.

## Default SQLite file location

- macOS: `~/Library/Application Support/gptscript/credentials.db`
  - if the `XDG_CONFIG_HOME` environment variable is set, the file will be located at `$XDG_CONFIG_HOME/gptscript/credentials.db`
- Linux: `$XDG_CONFIG_HOME/gptscript/credentials.db`

## Encryption Configuration

To enable encryption, you need to create an encryption configuration file. We use the same type of configuration
as [Kubernetes](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/) does.

This should allow you to use any Kubernetes-compatible KMS v2 providers, though we have not yet tested this.
We have only tested the built-in `aesgcm` provider, which uses a key stored locally in the configuration file.

### Default encryption configuration locations

- macOS: `~/Library/Application Support/gptscript/encryptionconfig.yaml`
  - if the `XDG_CONFIG_HOME` environment variable is set, the file will be located at `$XDG_CONFIG_HOME/gptscript/encryptionconfig.yaml`
- Linux: `$XDG_CONFIG_HOME/gptscript/encryptionconfig.yaml`

### Example: AES-GCM configuration with a key in the config file

```yaml
kind: EncryptionConfiguration
apiVersion: apiserver.config.k8s.io/v1
resources:
  - resources:
      # Note that the configuration here must be EXACTLY 'credentials'
      - credentials
    providers:
      - aesgcm:
          keys:
            - name: myKey
              secret: <key encoded in base64>
```

## Environment Variables

All helpers:
- `GPTSCRIPT_ENCRYPTION_CONFIG_FILE` - can be used to override the path to the encryption configuration file.

SQLite:
- `GPTSCRIPT_SQLITE_FILE` - can be used to override the path to the SQLite file.

PostgreSQL:
- `GPTSCRIPT_POSTGRES_DSN` - (required) the DSN (connection string) for the PostgreSQL database.

