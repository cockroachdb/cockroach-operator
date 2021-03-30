# Copyright 2021 The Cockroach Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Common helpers for teamcity-*.sh scripts.

remove_files_on_exit() {
  rm -f .cockroach-teamcity-key
  rm -rf ~/.docker
}
trap remove_files_on_exit EXIT

tc_start_block() {
  echo "##teamcity[blockOpened name='$1']"
}

tc_end_block() {
  echo "##teamcity[blockClosed name='$1']"
}

docker_login() {
  echo "${OPERATOR_DOCKER_ACCESS_TOKEN}" | docker login --username "${OPERATOR_DOCKER_ID}" --password-stdin
}

docker_login_with_redhat() {
  echo "${OPERATOR_REDHAT_REGISTRY_KEY}" | docker login --username "${OPERATOR_REDHAT_DOCKER_ID}" --password-stdin $rhel_registry
}

configure_docker_creds() {
  # Work around headless d-bus problem by forcing docker to use
  # plain-text credentials for dockerhub.
  # https://github.com/docker/docker-credential-helpers/issues/105#issuecomment-420480401
  mkdir -p ~/.docker
  cat << EOF > ~/.docker/config.json
{
  "credsStore" : "",
  "auths": {
    "https://index.docker.io/v1/" : {
    }
  }
}
EOF
}

configure_git_ssh_key() {
  # Write a private key file and populate known_hosts
  touch .cockroach-teamcity-key
  chmod 600 .cockroach-teamcity-key
  echo "${github_ssh_key}" > .cockroach-teamcity-key

  mkdir -p "$HOME/.ssh"
  ssh-keyscan github.com > "$HOME/.ssh/known_hosts"
}

push_to_git() {
  # $@ passes all arguments to this function to the command
  GIT_SSH_COMMAND="ssh -i .cockroach-teamcity-key" git push "$@"
}

docker_image_exists() {
  docker pull "$1"
  return $?
}
