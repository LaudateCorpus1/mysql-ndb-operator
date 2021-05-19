#!/usr/bin/env bash

# Copyright (c) 2020, Oracle and/or its affiliates.
#
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/

# Script to build a custom MySQL Cluster container image
# The image is built with the binaries either present in BASEDIR (or)
# with the binaries generated by compiling the source code in SRCDIR

# Note : script assumes the docker daemon to be running in the same
# host as the client when it compiles MYSQL Cluster inside docker.
# So, to make it work with the docker daemon hosted inside minikube,
# additionally the SRCDIR and TMPDIR should be mounted on minikube
# under the same path.
#   like,  minikube mount ${SRCDIR}/${SRCDIR}
#          minikube mount ${TMPDIR}/${TMPDIR}

# Helper function to print fatal error message
function fatal() {
  echo -e "$1"
  exit 1
}

# Returns version array from version files
function get_version() {
  while read -r line; do declare "$line"; done <"$1"
  echo "${MYSQL_VERSION_MAJOR} ${MYSQL_VERSION_MINOR} ${MYSQL_VERSION_PATCH}"
}

# Returns version in major.minor.patch format from version array
function get_version_str() {
  local IFS="."
  echo "$*"
}

# Verifies the version of the MySQL Cluster
function check_version() {
  version=("$@")
  if [ "${version[0]}" -ne "8" ] ||
    [ "${version[1]}" -ne "0" ] ||
    [ "${version[2]}" -lt "22" ]; then
    version_str=$(get_version_str "${mysql_cluster_version[@]}")
    fatal "MySQL Cluster version ${version_str} is not supported. Please use version 8.0.22 or above."
  fi
}

# BASEDIR or SRCDIR must be set
if [ -z "${BASEDIR}" ] && [ -z "${SRCDIR}" ]; then
  fatal "Please pass the MySQL Cluster build or install directory location via BASEDIR\n \
 (or) pass the MySQL Cluster source directory via SRCDIR"
fi

# Don't allow both BASEDIR and SRCDIR
if [ -n "${BASEDIR}" ] && [ -n "${SRCDIR}" ]; then
  fatal "Please specify only one of BASEDIR or SRCDIR"
fi

# Move into script dir
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
cd "${script_dir}"

# Array to hold MySQL Cluster's with Major, Minor and Patch version values
declare -a mysql_cluster_version

if [ -n "${SRCDIR}" ]; then
  # Source directory has been specified
  mysql_cluster_version=($(get_version "${SRCDIR}/MYSQL_VERSION"))
  check_version "${mysql_cluster_version[@]}"

  # Compile and build the binaries
  if [ -z "${TMPDIR}" ]; then
    if ! TMPDIR=$(mktemp -d); then
      fatal "Failed to create temp build directory at ${TMPDIR}"
    fi
  fi
  build_dir=${TMPDIR}

  # Build the docker image to compile MySQL Cluster
  image_name="mysql-cluster-builder:ol8"
  if ! DOCKER_BUILDKIT=1 docker build -t "${image_name}" ./ol8-builder; then
    fatal "Failed to build mysql-cluster-builder docker image"
  fi

  # Run the image to build MySQL Cluster
  name="compile-cluster-ol8"
  docker_cmd=(docker run --name=${name} --rm)
  # Add the source code as a volume
  docker_cmd+=(-v ${SRCDIR}:/mysql-cluster)
  # Add the build directory as a volume
  docker_cmd+=(-v ${build_dir}:/build)
  docker_cmd+=(${image_name})
  echo "${docker_cmd[*]}"
  if ! "${docker_cmd[@]}"; then
    fatal "Failed to compile MySQL Cluster"
  fi

  # Successfully built cluster. Binaries are at build_dir.
  bin_dir=${build_dir}/bin

  # Cleanup
  docker_cmd=(docker rmi ${image_name})
  if ! "${docker_cmd[@]}"; then
    echo "Warning : Failed to remove container image from docker"
  fi
fi

if [ -n "${BASEDIR}" ]; then
  # Verify that BASEDIR and bin_dir directories exist
  bin_dir="${BASEDIR}/bin"
  if [[ ! -d "${BASEDIR}" || ! -d "${bin_dir}" ]]; then
    fatal "Please set a valid MySQL Cluster build or install directory in BASEDIR"
  fi

  mysql_cluster_version=($(get_version "${BASEDIR}/VERSION.dep"))
  check_version "${mysql_cluster_version[@]}"
fi

# Determine docker-entrypoint version to use
declare entrypoint_ver
if [ "${version[2]}" -le "23" ]; then
  entrypoint_ver="1.1.18"
else
  entrypoint_ver="1.2.3"
fi

# Executable files to be copied inside the Docker image
exes_to_sbin=("ndbmtd" "ndb_mgmd" "mysqld" "mysqladmin")
exes_to_bin=("ndb_mgm" "mysql" "mysql_tzinfo_to_sql")
# TODO : Add more ndb_ tools and libs in the image and enable dynamic linking of libs

# copy all the required binaries to docker context
DOCKER_CTX_FILES="cluster-docker-files"
mkdir "${DOCKER_CTX_FILES}"

DOCKER_CTX_SBIN="${DOCKER_CTX_FILES}/sbin"
mkdir "${DOCKER_CTX_SBIN}"
for exe in "${exes_to_sbin[@]}"; do
  exe_path=${bin_dir}/${exe}
  cp "${exe_path}" "${DOCKER_CTX_SBIN}"
  chmod 755 "${DOCKER_CTX_SBIN}/${exe}"
done

DOCKER_CTX_BIN="${DOCKER_CTX_FILES}/bin"
mkdir "${DOCKER_CTX_BIN}"
for exe in "${exes_to_bin[@]}"; do
  exe_path=${bin_dir}/${exe}
  cp "${exe_path}" "${DOCKER_CTX_BIN}"
  chmod 755 "${DOCKER_CTX_BIN}/${exe}"
done

# Copy the docker entrypoint
mysql_cluster_version_str=$(get_version_str "${mysql_cluster_version[@]}")
sed "s/#VERSION#/${mysql_cluster_version_str}-${entrypoint_ver}/" \
  "entrypoints/docker-entrypoint-v${entrypoint_ver}.sh" > ${DOCKER_CTX_FILES}/docker-entrypoint.sh
chmod +x ${DOCKER_CTX_FILES}/docker-entrypoint.sh

# Copy prepare-image.sh script
cp ./prepare-image.sh "${DOCKER_CTX_FILES}"
chmod +x ${DOCKER_CTX_FILES}/prepare-image.sh

# Optional IMAGE_TAG; by default tag it as custom
if [ -z "${IMAGE_TAG}" ]; then
  IMAGE_TAG="custom"
fi

# Build container image
image_name=mysql/mysql-cluster:"${mysql_cluster_version_str}-${IMAGE_TAG}"
if ! DOCKER_BUILDKIT=1 docker build -t "${image_name}" -f Dockerfile ${DOCKER_CTX_FILES} ; then
  fatal "Failed to build mysql-cluster-builder docker image"
fi

# Cleanup all copied binaries
rm -rf "${DOCKER_CTX_FILES}"
