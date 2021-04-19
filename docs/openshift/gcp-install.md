# Create an openshift cluster on GCP

## Prereqisites

- GCP Project
- Linux or MacOS
- GCP permissions to create a service account
- various binaries downloaded below
- a domain that is hosted in GCP
- gcloud

## Install Binaries

### Login to RedHat Site

1. Login to openshift https://cloud.redhat.com/openshift/
2. Press "Create Cluster" button
3. Scroll down to "Run it yourself"
4. Click on Google Cloud
5. Select Installer Provisioned
6. Click on download pull secret	 

You have to download the pull secret file as it is required for the install.

### Install Required RH Binaries

#### MacOS

```bash
curl -LO https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-install-mac.tar.gz
curl -LO https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-client-mac.tar.gz

tar -xzf openshift-install-mac.tar.gz 
tar -xzf openshift-client-mac.tar.gz 
chmod +x ./openshift-install
chmod +x ./oc
sudo mv ./openshift-install /usr/local/bin/openshift-install
sudo mv ./oc /usr/local/bin/oc
rm openshift-client-mac.tar.gz openshift-install-mac.tar.gz
```
#### Linux

```bash
curl -LO https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-install-linux.tar.gz
curl -LO https://mirror.openshift.com/pub/openshift-v4/clients/ocp/latest/openshift-client-linux.tar.gz

tar -xzf openshift-install-linux.tar.gz 
tar -xzf openshift-client-linux.tar.gz 
chmod +x ./openshift-install
chmod +x ./oc

sudo mv ./openshift-install /usr/local/bin/openshift-install
sudo mv ./oc /usr/local/bin/oc
rm openshift-client-linux.tar.gz openshift-install-linux.tar.gz
```

## Configure GCP setting and Service Account

The installer will use the default configuration, so set the PROJECT on the default configuration.
If you have not logged in to use gcloud run:
```bash
gcloud auth application-default login
```

## Get the GCP Project Id

Run the following command and select the project id for GCP.

```bash
gcloud projects list --format="value(PROJECT_ID)"
```

This will return the full project id and not the project name. 

If your organization has too many projects then you can add a filter as well.  Replace `<start of project name>`
to the name of the project.

```bash
gcloud projects list --format="value(PROJECT_ID)" --filter="PROJECT_ID ~ ^<start of project name>"
```

See https://cloud.google.com/sdk/gcloud/reference/projects/list for more information. 
You may also access the full project id via the GCP console.


## Check you have a DNS domain

Make sure you have a DNS domain. Here is a command to view the domains.

```bash
gcloud dns managed-zones list --project <gcp project id>
```

## Execute the creation script

Create a script using the following example. See the comments on the script for running the commands.

```bash
./openshift-gcp-create.sh -p <gcp project id> -s <pull secret file> -z <root domain> -n <cluster name> -r <gcp region> 
```

An example:

```bash
./openshift-gcp-create.sh -p $(gcloud config get-value project) -s ~/Downloads/pull-secret.txt -z foo.com -n foo -r us-central1
```

The script is located in the same directory as the documentation.

## Accessing Cluster

See the following documentation for using `oc`.

https://docs.openshift.com/container-platform/4.7/installing/installing_gcp/installing-gcp-customizations.html#cli-logging-in-kubeadmin_installing-gcp-customizations

## Post-Install

Do not delete the folder where the installation was made, in our case `$HOME/openshift-<cluster name>`. If you delete this you will have to decommission manually the infrastructure from GCP.

## Clean-up

If you want to delete the cluster run this command to decommission infrastructure on GCP:
```bash
GOOGLE_CREDENTIALS=~/.gcp/osServiceAccount.json \
openshift-install destroy cluster --dir ~/oshift --log-level=debug
```

If you want to delete the service account run:

```bash
./openshift-iam-delete.sh -p <gcp project id> -n <cluster name>
```

An example:

```bash
./openshift-iam-delete.sh -p $(gcloud config get-value project) -n foo
```

To remove the cluster from Redhat list you have to go to Clusters menu, select the cluster, press Actions and choose Archive cluster option.
