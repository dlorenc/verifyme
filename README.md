# Verifiable GitHub Actions

This repo contains code to help make GitHub Actions verifiable.

[Verifiable Builds](https://cloud.google.com/security/binary-authorization-for-borg) are
a technique to build software in a manner that allows for an end-user to verify the
provenance of the final artifact.

One of the most common techniques is cryptographic verificaton, where the build environment
digitally signs information about the inputs, steps and outputs of a build.
This digitial signature is then published, along with the public key of the build environment.
End users can verify the provenance against the signature and public key.

This project uses a version of this based on [Transient-Key Cryptography](https://en.wikipedia.org/wiki/Transient-key_cryptography).

## GitHub Actions

GitHub Actions do not support verifiable builds directly today.
This repository contains a GitHub action that can be used in a Workflow to get pretty close.

### Overview

The "verifier" Action generates an ephemeral keypair (ECDSA/SHA256) which is used to sign
an artifact produced by a prior build step.

In addition to signing the artifact, this Action generates a payload of metadata about the build
environment itself, including the ID and URL of the GitHub action run.
This payload is also signed with the same private key as the artifact.

The Action outputs the final signatures, environment metadata, and public key over `STDOUT` into
the build logs and as Action output parameters.

Finally, the private key is destroyed, ensuring that no other artifacts or metadata can be verified
against this public key.

### Usage

In a workflow YAML:

```yaml
name: verified-builder
on: workflow_dispatch
jobs:
  build:
    name: Build and generate verification
    # Set the type of machine to run on
    runs-on: ubuntu-latest
    steps:
      # Checks out a copy of your repository on the ubuntu-latest machine
      - name: Checkout code
        uses: actions/checkout@v2
      - name: Do build
        < insert your build step here>
      # Runs the Super-Linter action
      - name: Run build verifier
        uses: dlorenc/verifyme/action@main
        with:
          filepath: main.go
```

### Verification

To verify an artifact built using this action, you must first have:

* The artifact you would like to verify.
* The signature of the artifact.
* The public key the artifact was signed with.
* The JSON provenance "envelope" produced with the artifact.
* The signature of this provenance "envelope".

All of this information can be found in the logs for a GitHub Action, but you might have received them
from somewhere else.
If you did not receive this data directly from the output of a GitHub Action, you **must** verify that
the data you received is the same as the in the GitHub Action after verification.

The data looks like this in a log:

```shell
Starting verifier with:  main.go
Self hash:  e7ed4bb7753d5396b5dc9eb47a3a3139fabc9bf2845b7214f8a48c524d74d078
publickey=LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFdUIvQnZyaXQ2M01aRndZaUVyU05DcDJhMlhEbwplL25MWDY5T2VZVnQ1enNpZzNuRm5aY05HZFV3WWJVTndCVElDTlRnVXJFc2puelJDSm9wNnAyY0NBPT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg==
signature=MEUCIEMdraqDqk5EaexakFa+QH/muoicTj083DbIjuASAh9ZAiEAlzhVw6FxDGoIKhvSdhppyUcF3kPcG1w86ngstw/nw74=
sha256=af478f5fc0f8dc5628cf8040a5b3a614b013e97b09f7b4d1ca56f4d6395d80f2
environment=eyJSdW5VcmwiOiJodHRwczovL2dpdGh1Yi5jb20vZGxvcmVuYy92ZXJpZnltZS9hY3Rpb25zL3J1bnMvNTIwOTY3NTk3IiwiR2l0SHViU2hhIjoiZTc1YmU4YzdhODE5ZWZjZjRjOGIwOWEwYjJkYWE5OTc0MWZlOWEzOCIsIkFydGlmYWN0U2hhIjoiYWY0NzhmNWZjMGY4ZGM1NjI4Y2Y4MDQwYTViM2E2MTRiMDEzZTk3YjA5ZjdiNGQxY2E1NmY0ZDYzOTVkODBmMiIsIlZlcmlmaWVyU2hhIjoiZTdlZDRiYjc3NTNkNTM5NmI1ZGM5ZWI0N2EzYTMxMzlmYWJjOWJmMjg0NWI3MjE0ZjhhNDhjNTI0ZDc0ZDA3OCJ9
environment_signature=MEUCIQDOujWMbhxhW8xqwyWYx0sKMpj7V1qYdlj0X7EgE3ceiwIgIwIQ84H7vkazUcY+IQ1wanDNeRndIIGEQHiGnqHmHfU=
```

To verify with this binary:

```shell
    $ verifier <signature> <public-key> <message>
```

For the above example, we first verify the artifact itself:

```shell
    $ signature="MEUCIEMdraqDqk5EaexakFa+QH/muoicTj083DbIjuASAh9ZAiEAlzhVw6FxDGoIKhvSdhppyUcF3kPcG1w86ngstw/nw74="
    $ artifact="<artifact>"
    $ publickey="LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFdUIvQnZyaXQ2M01aRndZaUVyU05DcDJhMlhEbwplL25MWDY5T2VZVnQ1enNpZzNuRm5aY05HZFV3WWJVTndCVElDTlRnVXJFc2puelJDSm9wNnAyY0NBPT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg=="
    $ verifer $signature $publickey <artifact>
    valid signature
```

Then we verify the environment:

```shell
    $ environment="eyJSdW5VcmwiOiJodHRwczovL2dpdGh1Yi5jb20vZGxvcmVuYy92ZXJpZnltZS9hY3Rpb25zL3J1bnMvNTIwOTY3NTk3IiwiR2l0SHViU2hhIjoiZTc1YmU4YzdhODE5ZWZjZjRjOGIwOWEwYjJkYWE5OTc0MWZlOWEzOCIsIkFydGlmYWN0U2hhIjoiYWY0NzhmNWZjMGY4ZGM1NjI4Y2Y4MDQwYTViM2E2MTRiMDEzZTk3YjA5ZjdiNGQxY2E1NmY0ZDYzOTVkODBmMiIsIlZlcmlmaWVyU2hhIjoiZTdlZDRiYjc3NTNkNTM5NmI1ZGM5ZWI0N2EzYTMxMzlmYWJjOWJmMjg0NWI3MjE0ZjhhNDhjNTI0ZDc0ZDA3OCJ9"
    $ environment_signature="MEUCIQDOujWMbhxhW8xqwyWYx0sKMpj7V1qYdlj0X7EgE3ceiwIgIwIQ84H7vkazUcY+IQ1wanDNeRndIIGEQHiGnqHmHfU="

    # Same public key as before
    $ publickey="LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFdUIvQnZyaXQ2M01aRndZaUVyU05DcDJhMlhEbwplL25MWDY5T2VZVnQ1enNpZzNuRm5aY05HZFV3WWJVTndCVElDTlRnVXJFc2puelJDSm9wNnAyY0NBPT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg=="

    $ verifer $environment_signature $publickey $environment
    valid signature

    # To view the environment payload:
    $ echo $environment | base64 --decode | jq .
    {
        "RunUrl": "https://github.com/dlorenc/verifyme/actions/runs/520967597",
        "GitHubSha": "e75be8c7a819efcf4c8b09a0b2daa99741fe9a38",
        "ArtifactSha": "af478f5fc0f8dc5628cf8040a5b3a614b013e97b09f7b4d1ca56f4d6395d80f2",
        "VerifierSha": "e7ed4bb7753d5396b5dc9eb47a3a3139fabc9bf2845b7214f8a48c524d74d078"
    }
```

**Note**: If you did not receive these values directly from the GitHub Actions logs, you must
check the logs in the `RunUrl`.
If you did receive the values from the logs, make sure the URL in this payload matches the URL
you viewed.


The same verifications can be done with openssl.
The signature and public-key are base64-encoded, so they must first be decoded.

```shell
    signature="MEUCIEMdraqDqk5EaexakFa+QH/muoicTj083DbIjuASAh9ZAiEAlzhVw6FxDGoIKhvSdhppyUcF3kPcG1w86ngstw/nw74="
    artifact="<artifact>"
    publickey="LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFdUIvQnZyaXQ2M01aRndZaUVyU05DcDJhMlhEbwplL25MWDY5T2VZVnQ1enNpZzNuRm5aY05HZFV3WWJVTndCVElDTlRnVXJFc2puelJDSm9wNnAyY0NBPT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg=="

    # Decode first and write to files
    echo $signature | base64 --decode > signature
    echo $publickey | base64 --decode > pub
    
    # Now verify the artifact
    openssl dgst -sha256 -verify pub -signature signature <artifact>
```

And to verify the environment: 

```shell
    $ environment="eyJSdW5VcmwiOiJodHRwczovL2dpdGh1Yi5jb20vZGxvcmVuYy92ZXJpZnltZS9hY3Rpb25zL3J1bnMvNTIwOTY3NTk3IiwiR2l0SHViU2hhIjoiZTc1YmU4YzdhODE5ZWZjZjRjOGIwOWEwYjJkYWE5OTc0MWZlOWEzOCIsIkFydGlmYWN0U2hhIjoiYWY0NzhmNWZjMGY4ZGM1NjI4Y2Y4MDQwYTViM2E2MTRiMDEzZTk3YjA5ZjdiNGQxY2E1NmY0ZDYzOTVkODBmMiIsIlZlcmlmaWVyU2hhIjoiZTdlZDRiYjc3NTNkNTM5NmI1ZGM5ZWI0N2EzYTMxMzlmYWJjOWJmMjg0NWI3MjE0ZjhhNDhjNTI0ZDc0ZDA3OCJ9"
    $ environment_signature="MEUCIQDOujWMbhxhW8xqwyWYx0sKMpj7V1qYdlj0X7EgE3ceiwIgIwIQ84H7vkazUcY+IQ1wanDNeRndIIGEQHiGnqHmHfU="

    # Same public key as before
    $ publickey="LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFdUIvQnZyaXQ2M01aRndZaUVyU05DcDJhMlhEbwplL25MWDY5T2VZVnQ1enNpZzNuRm5aY05HZFV3WWJVTndCVElDTlRnVXJFc2puelJDSm9wNnAyY0NBPT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg=="

    $ echo $environment > environment
    $ echo $environment_signature > signature
    $ echo $publickey > pub
    $ openssl dgst -sha256 -verify pub -signature signature <artifact>
```

See the same note as above around verifying the contents in the Action logs as well.
