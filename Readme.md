# Introduction

This repository takes you through 2 demos:

1. Simple [Cloud Build](https://cloud.google.com/cloud-build/docs) pipeline to deploy to managed [Google Cloud Run](https://cloud.google.com/run/docs)
2. More advanced pipeline (testing, [vulnarability scanning](https://cloud.google.com/container-registry/docs/vulnerability-scanning), [Binary Authorization](https://cloud.google.com/binary-authorization/docs)) to deploy to [Cloud Run on Anthos / GKE](https://cloud.google.com/run/docs/gke/setup)

Please note that the content of this repository is not an officially supported Google product. It does however use officially supported Google Cloud products.

# Setup

In these instruction we will use environment variables to refer to the right project, region, cluster, etc.. After you have activated the right project for `gcloud`, update the following environmnet variables to your situation and set them using the following commands:

```bash
export PROJECT_ID=$(gcloud config list --format 'value(core.project)')
export PROJECT_NUMBER="$(gcloud projects describe "${PROJECT_ID}" --format='value(projectNumber)')"
export CLOUD_BUILD_SA_EMAIL="${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com"
export REGION=europe-west1
export ZONE=europe-west1-c
export CLUSTER=ci-cd-demo
```

These demos require the following Google Cloud API's to be enabled:

- [Cloud Build](https://cloud.google.com/cloud-build/docs) API
- [Cloud Run](https://cloud.google.com/run/docs) API

You can check if the APIs are enabled like this:
```bash
gcloud services list --enabled --filter="config.name:('cloudbuild.googleapis.com', 'run.googleapis.com')"
```

Output should be similar to:
```bash
NAME                       TITLE
cloudbuild.googleapis.com  Cloud Build API
run.googleapis.com         Cloud Run Admin API
```

Also, please make sure that the [Cloud Build](https://cloud.google.com/cloud-build/docs) service account (<proj-number>@cloudbuild.gserviceaccount.com) has the following roles:

- Service Account User (roles/iam.serviceAccountUser)
- Cloud Run Admin (roles/run.admin)
- Kubernetes Engine Admin (roles/container.admin)

You can list the roles of the service account like this:
```bash
gcloud projects get-iam-policy $PROJECT_ID \
--flatten="bindings[].members" \
--format='table(bindings.role)' \
--filter="bindings.members:$CLOUD_BUILD_SA_EMAIL"
```

Output should be similar to:
```bash
ROLE
roles/cloudbuild.builds.builder
roles/container.developer
roles/iam.serviceAccountUser
roles/run.admin
```

# DEMO 1: Cloud Build to managed Cloud Run

This demo focuses on a simple build pipeline to deploy the go app in this repo to the fully managed version of [Google Cloud Run](https://cloud.google.com/run/docs).

## Build container locally with docker

First you can test building the container locally using Docker build.

```bash
docker build .
```

```bash
docker images list
```

Check the latest image ID and run your local container image:

```bash
PORT=8080 && docker run -p 8080:${PORT} -e PORT=${PORT} <<IMAGE-ID>>>>
```

Now you can navigate to localhost:8080 and see the demo webpage.

## Build the container with Cloud Build

The following command accomplishes building the container using [Google Cloud Build](https://cloud.google.com/cloud-build/docs) and pushing the created image to [Google Cloud Container Registry](https://cloud.google.com/container-registry/docs).

```bash
gcloud builds submit --tag="gcr.io/${PROJECT_ID}/cloudrun_managed_demo:latest"
```

If you want to test the container that was pushed to the [Container Registry](https://cloud.google.com/container-registry/docs) locally, you can do the following:

_Requires the following setup first_

```bash
gcloud auth configure-docker
```

```bash
PORT=8080 && docker run -p 8080:${PORT} -e PORT=${PORT} gcr.io/${PROJECT_ID}/demo:latest
```

## Deploy to Cloud Run

With the image in [Google Cloud Container Registry](https://cloud.google.com/container-registry/docs), the next step is to deploy our small app to [Cloud Run](https://cloud.google.com/run/docs).

```bash
gcloud run deploy managed-demo \
--image gcr.io/${PROJECT_ID}/cloudrun_managed_demo \
--platform managed \
--region ${REGION} \
--allow-unauthenticated
```

## Create Cloud Build pipeline

Ok, now let's put all that into a single build pipeline that can be executed at once by [Google Cloud Build](https://cloud.google.com/cloud-build/docs). You can find the content of the configuration of the pipeline in `/cloudbuild/01-cloudrun_manual.yaml`.

Before executing the command we have to make sure that [Cloud Build](https://cloud.google.com/cloud-build/docs) is actually allowed to deploy to [Cloud Run](https://cloud.google.com/run/docs). The easiest way to accomplish this is to follow the instructions [here](https://cloud.google.com/cloud-build/docs/securing-builds/configure-access-for-cloud-build-service-account).
That out of the way, let's build the container, push it to the registry and deploy it to [Cloud Run](https://cloud.google.com/run/docs) in one command!

```bash
gcloud builds submit --config="./cloudbuild/01-cloudrun_manual.yaml"
```

## Setting up a trigger

[Google Cloud Build](https://cloud.google.com/cloud-build/docs) can automatically be triggered on a change in the repository. To demonstrate this we will set up a remote repo in [Google Cloud Repositories](https://cloud.google.com/source-repositories/docs) as [described here](https://cloud.google.com/source-repositories/docs/adding-repositories-as-remotes).

Once we have the remote repo we can [configure the trigger](https://cloud.google.com/cloud-build/docs/automating-builds/create-manage-triggers). Make sure you reference the `/cloudbuild/02-cloudrun-triggered.yaml` as the build file. This new definition automatically tags the created image with the originating commit in the repo. This is accomplished by one of [Cloud Build's default subsitution variales](https://cloud.google.com/cloud-build/docs/configuring-builds/substitute-variable-values#using_default_substitutions), namely `$SHORT_SHA`.

When we commit and push a change to our remote repository, [Cloud Build](https://cloud.google.com/cloud-build/docs) will automatically initiate the build!

# DEMO 2 :Cloud Build to Cloud Run on Anthos

The second demo focusses on

- [Cloud Run on Anthos / GKE](https://cloud.google.com/run/docs/gke/setup)
- uses a [Cloud Build](https://cloud.google.com/cloud-build/docs) pipeline with custom steps
- uses [vulnarability scanning](https://cloud.google.com/container-registry/docs/vulnerability-scanning) and [Binary Authorization](https://cloud.google.com/binary-authorization/docs)

The setup is a simplified version of [this demo](https://cloud.google.com/solutions/binary-auth-with-cloud-build-and-gke). Instead of a staging and production cluster we are using a simplified version with only one GKE cluster.

![demo 2a flow](docs/demo2a_flow.png)

Like before, we start by doing the deployment step by step manually, followed by an automated pipeline powered by [Cloud Build](https://cloud.google.com/cloud-build/docs).

_If you don't want to trigger the build to managed [Cloud Run](https://cloud.google.com/run/docs) during this demo, you can temporarily disable the trigger in the console under Cloud Build -> Triggers and clicking the three vertical dots next to your trigger._

## Create a Cluster

First we need an Anthos cluster to deploy on. Please note that we are enabling [Cloud Run on Anthos](https://cloud.google.com/run/docs/gke/setup) and [Binary Authorization](https://cloud.google.com/binary-authorization/docs) on this cluster.

```bash
gcloud container clusters create ${CLUSTER} \
--addons=HorizontalPodAutoscaling,HttpLoadBalancing,CloudRun \
--machine-type=n1-standard-4 \
--num-nodes=1 \
--enable-stackdriver-kubernetes \
--enable-ip-alias \
--cluster-version="1.15.12-gke.20" \
--enable-binauthz \
--zone=${ZONE}
```

Gain credentials to manipulate the cluster from command line:

```bash
gcloud container clusters get-credentials ${CLUSTER} --zone=$ZONE
```

## Deploy to Cloud Run on Anthos

We can simply deploy our previously created application container to [Cloud Run on Anthos](https://cloud.google.com/run/docs/gke/setup) by executing:

```bash
gcloud run deploy anthos-demo \
--image=gcr.io/${PROJECT_ID}/managed-demo \
--platform=gke \
--cluster=ci-cd-demo \
--cluster-location=${ZONE} \
--project=${PROJECT_ID}
```

To validate whether our new service is actually producing content, we first need to get the public (EXTERNAL) IP address of the istio ingress gateway. This can be found using the command:

```bash
kubectl get svc istio-ingress -n gke-system
```

For [Cloud Run on Anthos](https://cloud.google.com/run/docs/gke/setup) to receive the request, the request should be directed to the domain serving. As we do not have our new service attached to a domain, we can check the response either by:

- Using curl with an override of the request header (for example)
  ```bash
  curl -v -H "Host: anthos-demo.default.example.com" http://IP_ADDRESS/
  ```
- Adding an entry to our local hosts file, which on Linux can be found in the /etc folder:
  ```bash
  sudo nano /etc/hosts
  ```
  Add the following line at the end of the file with the IP address you found above:
  ```
  IP_ADDRESS anthos-demo.default.example.com
  ```

Then, navigate your browser to http://anthos-demo.default.example.com:

## Adding unit test

Now, let's enhance the manual [Cloud Build](https://cloud.google.com/cloud-build/docs) deployment by including an additional step for executing a simple unit test and cancel the deployment when the test fails (which happens automatically, because the test returns an exit code other than 0 on failure).

In the file `/cloudrun/03-anthos_manual.yaml` you find the build pipeline with the unit test included. Let execute this pipeline with the same command as before, but now referening the new config file.

```bash
gcloud builds submit --config="./cloudbuild/03-anthos_manual.yaml"
```

# Using Binary Authorisation

## Set up attestation definition (note)

Next in the [demo](https://cloud.google.com/solutions/binary-auth-with-cloud-build-and-gke#creating_signing_keys) is creating a note (definition) which is used by some code in a custom step of hour [Cloud Build](https://cloud.google.com/cloud-build/docs) pipeline to attest that no significant vulnaribilites were found. The note occurence (attestation) is signed with a private key.

At deployment time, [Binary Authorization](https://cloud.google.com/binary-authorization/docs) will use an Attestor for this note definition to verify created attestations using the public key.

So, let's first create the [Container Analysis](https://cloud.google.com/container-registry/docs/container-analysis) note:

```bash
curl "https://containeranalysis.googleapis.com/v1/projects/${PROJECT_ID}/notes/?noteId=vulnz-note" \
  --request "POST" \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer $(gcloud auth print-access-token)" \
  --header "X-Goog-User-Project: ${PROJECT_ID}" \
  --data-binary @- <<EOF
    {
      "name": "projects/${PROJECT_ID}/notes/vulnz-note",
      "attestation": {
        "hint": {
          "human_readable_name": "Vulnerability scan note"
        }
      }
    }
EOF
```

## Set up key pair for signing and verifying

Next part is the signing keys for [Cloud Build](https://cloud.google.com/cloud-build/docs). In our simplified demo we will only use the vulnarability check (vulnz-signer) and not the quality assurance (qa-signer) check.

Create the key ring and the signer / verifier key pair:

```bash
gcloud kms keyrings create "binauthz" \
  --project "${PROJECT_ID}" \
  --location "${REGION}"

gcloud kms keys create "vulnz-signer" \
  --project "${PROJECT_ID}" \
  --location "${REGION}" \
  --keyring "binauthz" \
  --purpose "asymmetric-signing" \
  --default-algorithm "rsa-sign-pkcs1-4096-sha512"
```

## Signing rights

When [Cloud Build](https://cloud.google.com/cloud-build/docs) wants to sign an image (create an attestation in the script 'create_attestaion.sh') the associated [gcloud command](https://cloud.google.com/sdk/gcloud/reference/alpha/container/binauthz/attestations/sign-and-create) requires the Attestor resource as one of its parameters. The command needs to verify the existence of the Attestor and retrieve its information. Therefore we give it the role *attestorsViewer* on the Attestor resource `vulnz-attestor`:

```bash
gcloud container binauthz attestors add-iam-policy-binding "vulnz-attestor" \
  --project "${PROJECT_ID}" \
  --member "serviceAccount:${CLOUD_BUILD_SA_EMAIL}" \
  --role "roles/binaryauthorization.attestorsViewer"
```

Secondly, the [Cloud Build](https://cloud.google.com/cloud-build/docs) service account needs permission to view note occurences and attach the `vulnz-note` note to container images:

```bash
curl "https://containeranalysis.googleapis.com/v1beta1/projects/${PROJECT_ID}/notes/vulnz-note:setIamPolicy" \
  --request POST \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer $(gcloud auth print-access-token)" \
  --header "X-Goog-User-Project: ${PROJECT_ID}" \
  --data-binary @- <<EOF
    {
      "resource": "projects/${PROJECT_ID}/notes/vulnz-note",
      "policy": {
        "bindings": [
          {
            "role": "roles/containeranalysis.notes.occurrences.viewer",
            "members": [
              "serviceAccount:${CLOUD_BUILD_SA_EMAIL}"
            ]
          },
          {
            "role": "roles/containeranalysis.notes.attacher",
            "members": [
              "serviceAccount:${CLOUD_BUILD_SA_EMAIL}"
            ]
          }
        ]
      }
    }
EOF
```

And lastly, [Cloud Build](https://cloud.google.com/cloud-build/docs) needs to have the permission to use the private key of the `vulnz-signer` key pair to digitally sign the notes:

```bash
gcloud kms keys add-iam-policy-binding "vulnz-signer" \
  --project "${PROJECT_ID}" \
  --location "${REGION}" \
  --keyring "binauthz" \
  --member "serviceAccount:${CLOUD_BUILD_SA_EMAIL}" \
  --role 'roles/cloudkms.signerVerifier'
```

## Attestor rights

[Binary Authorization](https://cloud.google.com/binary-authorization/docs) works with a concept called attestors. These attestors validate that certain attestations exist in [Container Analysis](https://cloud.google.com/container-registry/docs/container-analysis) storage. By adding attestors to your [Binary Authorization](https://cloud.google.com/binary-authorization/docs) policy, which is then applied to a cluster, you protect the cluster from running containers that don't have all the necessary attestations.

Create the vulnerability scan attestor. This attestor will check whether the vulz-note attestation exists for newly created images.

```bash
gcloud container binauthz attestors create "vulnz-attestor" \
  --project "${PROJECT_ID}" \
  --attestation-authority-note-project "${PROJECT_ID}" \
  --attestation-authority-note "vulnz-note" \
  --description "Vulnerability scan attestor"
```

Add the public key of the `vulnz-signer` key-pair to the attestor in order for it to verify the attestation:

```bash
gcloud beta container binauthz attestors public-keys add \
  --project "${PROJECT_ID}" \
  --attestor "vulnz-attestor" \
  --keyversion "1" \
  --keyversion-key "vulnz-signer" \
  --keyversion-keyring "binauthz" \
  --keyversion-location "${REGION}" \
  --keyversion-project "${PROJECT_ID}"
```

## Set up Binary Authorization policy

Now that we have an attestor, we can set up an appropriate [Binary Authorization](https://cloud.google.com/binary-authorization/docs) policy to block deployment unless the vulnarability attestation is present for the image. You can either generate a policy definition and uplaod it, or configure the policy in the console UI. Here is how to use the upload version:

```bash
cat > ./binauthz-policy.yaml <<EOF
admissionWhitelistPatterns:
- namePattern: docker.io/istio/*
- namePattern: gke.gcr.io/istio/*
- namePattern: gke.gcr.io/knative/*
defaultAdmissionRule:
  enforcementMode: ENFORCED_BLOCK_AND_AUDIT_LOG
  evaluationMode: ALWAYS_DENY
globalPolicyEvaluationMode: ENABLE
clusterAdmissionRules:
  ${ZONE}.${CLUSTER}:
    evaluationMode: REQUIRE_ATTESTATION
    enforcementMode: ENFORCED_BLOCK_AND_AUDIT_LOG
    requireAttestationsBy:
    - projects/${PROJECT_ID}/attestors/vulnz-attestor

EOF

gcloud container binauthz policy import ./binauthz-policy.yaml \
  --project "${PROJECT_ID}"
```

## Adding signing and verifying steps to Cloud Build

So far we have only registered the actors and their permissions in this vulnarability checking scenario. The actual code to check the number of vulnarabilities, sign off the image with the `vulnz-note` and check whether the attestation exists before deploying, was purposely written for this demo and can be found [here](https://github.com/GoogleCloudPlatform/gke-binary-auth-tools).

In a separate folder clone this repository:

```bash
git clone https://github.com/GoogleCloudPlatform/gke-binary-auth-tools
```

To use this code in [Cloud Build](https://cloud.google.com/cloud-build/docs) we build a container for it and reference it in our [Cloud Build](https://cloud.google.com/cloud-build/docs) custom steps. Then we call specifc commands and code in the container by passing arguments to the container's entrypoint.

Build and push the container to our [Container Registry](https://cloud.google.com/container-registry/docs) by executing the following command from your new `binauthz-tools` folder:

```bash
gcloud builds submit \
  --project "${PROJECT_ID}" \
  --tag "gcr.io/${PROJECT_ID}/cloudbuild-attestor" \
  .
```

Now we are ready to add additional steps to our [Cloud Build](https://cloud.google.com/cloud-build/docs) pipeline. Please look at the updated pipeline `05-anthos_triggered_2.yaml` where we have added the `Check` and `Sign` step as well as necessary substitution variables.

```yaml
# ...

- name: gcr.io/$PROJECT_ID/cloudbuild-attestor
  id: Check
  entrypoint: sh
  args:
    - -xe
    - -c
    - |
      /scripts/check_vulnerabilities.sh -p $PROJECT_ID -i gcr.io/$PROJECT_ID/$_IMAGE_NAME:$SHORT_SHA -t 5

- name: gcr.io/$PROJECT_ID/cloudbuild-attestor
  id: Sign
  entrypoint: sh
  args:
    - -xe
    - -c
    - |-
      FQ_DIGEST=$(gcloud container images describe --format 'value(image_summary.fully_qualified_digest)' gcr.io/$PROJECT_ID/$_IMAGE_NAME:$SHORT_SHA)
      /scripts/create_attestation.sh \
        -p "$PROJECT_ID" \
        -i "$${FQ_DIGEST}" \
        -a "$_VULNZ_ATTESTOR" \
        -v "$_VULNZ_KMS_KEY_VERSION" \
        -k "$_VULNZ_KMS_KEY" \
        -l "$_KMS_LOCATION" \
        -r "$_KMS_KEYRING"

# ...
```

## Try a successful deployment

In order for the deployment to work we need to update our trigger. Make sure it now uses `05-anthos_triggered_2.yaml`. Push a new version to the repository and watch the [Cloud Build](https://cloud.google.com/cloud-build/docs) progress.

## Try failing deployements

### Too many vulnarabilities

As an example to demonstrate the vulnarability check to fail, change the `Dockerfile` to make the container base image _Debian_ instead of _Alpine_. Simply uncomment and comment a few lines and make it look like:

```Dockerfile
#...

FROM debian:stable-slim
RUN apt-get update -qq &&\
    apt-get -qq install -qqy ca-certificates

# FROM alpine:3
# RUN apk add --no-cache ca-certificates

#...
```

The failing deployment is not caused by [Binary Authorization](https://cloud.google.com/binary-authorization/docs) and the attestation process. The `check_vulnerabilities.sh` script (from the `binauthz-tools` repo) retrieves a list of found vulnarabilities (which by the way are stored in [Container Analysis](https://cloud.google.com/container-registry/docs/container-analysis)) and finds that these exceed a cartain threshold. By returning an error in this case, [Cloud Build](https://cloud.google.com/cloud-build/docs) will stop the pipeline.

### Deploy without Cloud Build (and hence no attestation)

Try to deploy an unauthorized image to cluster and investigate the reponse highlighting this is forbidden:

```Shell
$ kubectl run hello-server --generator=run-pod/v1 --port 8080 --image gcr.io/google-samples/hello-app@sha256:c62ead5b8c15c231f9e786250b07909daf6c266d0fcddd93fea882eb722c3be4

Error from server (Forbidden): pods "hello-server" is forbidden: image policy webhook backend denied one or more images: Denied by cluster admission rule for europe-west1-c.ci-cd-demo. Denied by Attestor. Image gcr.io/google-samples/hello-app@sha256:c62ead5b8c15c231f9e786250b07909daf6c266d0fcddd93fea882eb722c3be4 denied by projects/<REDACTED>/attestors/vulnz-attestor: No attestations found that were valid and signed by a key trusted by the attestor
```

In this case it is the [Binary Authorization](https://cloud.google.com/binary-authorization/docs) setup preventing the deployment.

## Using the Kritis signer

A newer solution to interpret the vularability scan results and execute the signing is to leverage the open source [Kritis](https://github.com/grafeas/kritis) project that can be used as a build step in [Cloud Build](https://cloud.google.com/cloud-build/docs). [This article](https://cloud.google.com/binary-authorization/docs/vulnerability-signing-kritis) describes the set up, which is very similar to what we have done so far. So, let's make a few changes in order to use the newer [Kritis](https://github.com/grafeas/kritis) signer.

![demo 2b flow](docs/demo2b_flow.png)

First, we need to create the [Kritis](https://github.com/grafeas/kritis) container image for [Cloud Build](https://cloud.google.com/cloud-build/docs). For convenience, the build step using [Cloud Build](https://cloud.google.com/cloud-build/docs) itself is included in the [Kritis](https://github.com/grafeas/kritis) repository.

In a separate folder clone this repository:

```bash
git clone --branch signer-v1.0.0 https://github.com/grafeas/kritis.git
```

Build and push the container to our [Container Registry](https://cloud.google.com/container-registry/docs) by executing the following command from your new `kritis` folder:

```bash
gcloud --project=$PROJECT_ID builds submit . --config deploy/kritis-signer/cloudbuild.yaml
```

An image called `kritis-signer` should now be available in the project's [Container Registry](https://cloud.google.com/container-registry/docs) and can be used as a step in the build pipeline.

The [Kritis signer](https://github.com/grafeas/kritis/blob/master/docs/signer.md) needs the full digest of the container, which is being created in the pipeline. To retrieve this full digest either a `docker` or `gcloud` command can be used. These commands are not available in the Kritis build step's container, so the value needs to be passed on from a previous step. In a Cloud Build pipeline this can be accomplished by storing the value in a file on a [volume](https://cloud.google.com/cloud-build/docs/build-config#volumes), which remains available during the entire pipeline.

The first change is to the `Push` step in our pipeline. We use the same `docker` builder, but use bash as entry point in order to execute multiple `docker` commands:

- Push the container to the registry
- Store the full digest of the built container in a file called `image-digest.txt`:

```yaml
# Using an alternative "docker push" step in order to pass the full image digest to the next steps
# The Kritis container has no docker, nor gcloud to get the full digest by itself
- name: gcr.io/cloud-builders/docker
  id: Push
  entrypoint: /bin/bash
  args:
    - -c
    - |
      docker push gcr.io/$PROJECT_ID/$_IMAGE_NAME:$SHORT_SHA &&
      docker image inspect gcr.io/$PROJECT_ID/$_IMAGE_NAME:$SHORT_SHA --format '{{index .RepoDigests 0}}' > image-digest.txt &&
      cat image-digest.txt
```

The updated [Cloud Build](https://cloud.google.com/cloud-build/docs) pipeline, which uses the [Kritis signer](https://github.com/grafeas/kritis/blob/master/docs/signer.md) can be found in `06-anthos_triggered_3.yaml`.

The second thing that needs to change is to replace the `Check` and `Sign` steps with a single new `Sign` step, which uses the [Kritis signer](https://github.com/grafeas/kritis/blob/master/docs/signer.md) container we created. This step does both the checking and signing.

```yaml
- name: gcr.io/$PROJECT_ID/kritis-signer
  id: Sign
  entrypoint: /bin/bash
  args:
    - -c
    - |-
      FQ_KMS_KEY="projects/${PROJECT_ID}/\
      locations/${_KMS_LOCATION}/\
      keyRings/${_KMS_KEYRING}/\
      cryptoKeys/${_VULNZ_KMS_KEY}/\
      cryptoKeyVersions/${_VULNZ_KMS_KEY_VERSION}"

      /kritis/signer \
      -v=10 \
      -alsologtostderr \
      -image=$(/bin/cat image-digest.txt) \
      -policy=kritis-policies/policy-low.yaml \
      -kms_key_name="$${FQ_KMS_KEY}" \
      -kms_digest_alg=${_VULNZ_KMS_DIGEST_ALG} \
      -note_name="projects/${PROJECT_ID}/notes/${_VULNZ_NOTE}"
```

Notice how the following inline bash command retrieves the full image digest from the file created in the previous build step.

```bash
  -image=$(/bin/cat image-digest.txt)
```

The signer also needs a full reference to the `KMS key` and the `Attestor Note`, so we construct those as can be seen above. To keep things organized, we add two new substitution variables:

- `_VULNZ_KMS_DIGEST_ALG`: specifying the algorithm used for signing (we used SHA512)
- `_KRITIS_POLICY`: specifies the Kritis policy file to use for processing the scanning results. An example of such policy looks like:

```yaml
apiVersion: kritis.grafeas.io/v1beta1
kind: VulnzSigningPolicy
metadata:
  name: vulnz-medium
spec:
  imageVulnerabilityRequirements:
    maximumFixableSeverity: MEDIUM
    maximumUnfixableSeverity: MEDIUM
    allowlistCVEs:
      - projects/goog-vulnz/notes/CVE-2020-10543
      - projects/goog-vulnz/notes/CVE-2020-10878
      - projects/goog-vulnz/notes/CVE-2020-14155
```

In the folder `kritis-policies` we have included a _LOW_ and _Medium_ policy.

That is it. Update the Cloud Build trigger using the updated pipeline `/cloudbuild/06-anthos_triggered_3.yaml`.
