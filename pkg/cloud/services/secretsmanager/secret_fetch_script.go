/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package secretsmanager

//nolint
const secretFetchScript = `#!/bin/bash

# Copyright 2020 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

REGION="{{.Region}}"
SECRET_ID="{{.SecretID}}"
FILE="/etc/secret-userdata.txt"

# Log an error and exit.
# Args:
#   $1 Message to log with the error
#   $2 The error code to return
log::error_exit() {
	local message="${1}"
	local code="${2}"

	log::error "${message}"
	log::error "aws.cluster.x-k8s.io encrypted cloud-init script $0 exiting with status ${code}"
	exit "${code}"
}

log::success_exit() {
	log::info "aws.cluster.x-k8s.io encrypted cloud-init script $0 finished"
	exit 0
}

# Log an error but keep going.
log::error() {
	local message="${1}"
	timestamp=$(date --iso-8601=seconds)
	echo "!!! [${timestamp}] ${1}" >&2
	shift
	for message; do
		echo "    ${message}" >&2
	done
}

# Print a status line.  Formatted to show up in a stream of output.
log::info() {
	timestamp=$(date --iso-8601=seconds)
	echo "+++ [${timestamp}] ${1}"
	shift
	for message; do
		echo "    ${message}"
	done
}

check_aws_command() {
	local command="${1}"
	local code="${2}"
	local out="${3}"
	local sanitised="${out//[$'\t\r\n']/}"
	case ${code} in
	"0")
		log::info "AWS CLI reported successful execution for ${command}"
		;;
	"2")
		log::error "AWS CLI reported that it could not parse ${command}"
		log::error "${sanitised}"
		;;
	"130")
		log::error "AWS CLI reported SIGINT signal during ${command}"
		log::error "${sanitised}"
		;;
	"255")
		log::error "AWS CLI reported service error for ${command}"
		log::error "${sanitised}"
		;;
	*)
		log::error "AWS CLI reported unknown error ${code} for ${command}"
		log::error "${sanitised}"
		;;
	esac
}

delete_secret_value() {
	local out
	log::info "deleting secret from AWS Secrets Manager"
	out=$(
		set +e
		aws secretsmanager --region ${REGION} delete-secret --force-delete-without-recovery --secret-id ${SECRET_ID} 2>&1
	)
	local delete_return=$?
	check_aws_command "SecretsManager::DeleteSecret" "${delete_return}" "${out}"
	if [ ${delete_return} -ne 0 ]; then
		log::error_exit "Could not delete secret value" 2
	fi
}

log::info "aws.cluster.x-k8s.io encrypted cloud-init script $0 started"
umask 006
if test -f "${FILE}"; then
	log::info "encrypted userdata already written to disk"
	log::success_exit
fi

log::info "getting userdata from AWS Secrets Manager"
log::info "getting secret value from AWS Secrets Manager"
BINARY_DATA=$(
	set +e
	aws secretsmanager --region ${REGION} get-secret-value --output text --query 'SecretBinary' --secret-id ${SECRET_ID} 2>&1
)
GET_RETURN=$?
check_aws_command "SecretsManager::GetSecretValue" "$?" "${BINARY_DATA}"
if [ ${GET_RETURN} -ne 0 ]; then
	log::error "could not get secret value, deleting secret"
	delete_secret_value
	log::error_exit "could not get secret value, but secret was deleted" 1
fi

delete_secret_value

log::info "decoding and decompressing userdata"
UNZIPPED=$(echo "${BINARY_DATA}" | base64 -d | gunzip)
GUNZIP_RETURN=$?
if [ ${GUNZIP_RETURN} -ne 0 ]; then
	log::error_exit "could not get unzip data" 4
fi

log::info "writing userdata to ${FILE}"
echo -n "${UNZIPPED}" >${FILE}
log::info "restarting cloud-init"
systemctl restart cloud-init
log::success_exit
`
