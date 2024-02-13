#!/bin/bash

set -e -x -o pipefail

export OWNER="forge4flow"
export REPO="forge4flow-manager"

# On CentOS /usr/local/bin is not included in the PATH when using sudo. 
# Running arkade with sudo on CentOS requires the full path
# to the arkade binary. 
export ARKADE=/usr/local/bin/arkade

# When running as a startup script (cloud-init), the HOME variable is not always set.
# As it is required for arkade to properly download tools, 
# set the variable to /usr/local so arkade will download binaries to /usr/local/.arkade
if [ -z "${HOME}" ]; then
  export HOME=/usr/local
fi

version=""

echo "Finding latest version from GitHub"
version=$(curl -sI https://github.com/$OWNER/$REPO/releases/latest | grep -i "location:" | awk -F"/" '{ printf "%s", $NF }' | tr -d '\r')
echo "$version"

if [ ! $version ]; then
  echo "Failed while attempting to get latest version"
  exit 1
fi

SUDO=sudo
if [ "$(id -u)" -eq 0 ]; then
  SUDO=
fi

verify_system() {
  if ! [ -d /run/systemd ]; then
    fatal 'Can not find systemd to use as a process supervisor for forge4flow-manager'
  fi
}

has_yum() {
  [ -n "$(command -v yum)" ]
}

has_apt_get() {
  [ -n "$(command -v apt-get)" ]
}

has_pacman() {
  [ -n "$(command -v pacman)" ]
}

install_required_packages() {
  if $(has_apt_get); then
    # Debian bullseye is missing iptables. Added to required packages
    # to get it working in raspberry pi. No such known issues in
    # other distros. Hence, adding only to this block.
    $SUDO apt-get update -y
    $SUDO apt-get install -y curl runc bridge-utils iptables
  elif $(has_yum); then
    $SUDO yum check-update -y
    $SUDO yum install -y curl runc iptables-services
  elif $(has_pacman); then
    $SUDO pacman -Syy
    $SUDO pacman -Sy curl runc bridge-utils
  else
    fatal "Could not find apt-get, yum, or pacman. Cannot install dependencies on this OS."
    exit 1
  fi
}

install_arkade(){
  curl -sLS https://get.arkade.dev | $SUDO sh
  arkade --help
}

install_cni_plugins() {
  cni_version=v0.9.1
  $SUDO $ARKADE system install cni --version ${cni_version} --path /opt/cni/bin --progress=false
}

install_containerd() {
  CONTAINERD_VER=1.7.0
  $SUDO systemctl unmask containerd || :

  arch=$(uname -m)
  if [ $arch == "armv7l" ]; then
    $SUDO curl -fSLs "https://github.com/alexellis/containerd-arm/releases/download/v${CONTAINERD_VER}/containerd-${CONTAINERD_VER}-linux-armhf.tar.gz" --output "/tmp/containerd.tar.gz"
    $SUDO tar -xvf /tmp/containerd.tar.gz -C /usr/local/bin/
    $SUDO curl -fSLs https://raw.githubusercontent.com/containerd/containerd/v${CONTAINERD_VER}/containerd.service --output "/etc/systemd/system/containerd.service"
    $SUDO systemctl enable containerd
    $SUDO systemctl start containerd
  else
    $SUDO $ARKADE system install containerd --systemd --version v${CONTAINERD_VER}  --progress=false
  fi
  
  sleep 5
}

install_forge4flow-manager() {
  arch=$(uname -m)
  case $arch in
  x86_64 | amd64)
    suffix=""
    ;;
  aarch64)
    suffix=-arm64
    ;;
  armv7l)
    suffix=-armhf
    ;;
  *)
    echo "Unsupported architecture $arch"
    exit 1
    ;;
  esac

  $SUDO curl -fSLs "https://github.com/forge4flow/forge4flow-manager/releases/download/${version}/f4f-manager${suffix}" --output "/usr/local/bin/f4f-manager"
  $SUDO chmod a+x "/usr/local/bin/f4f-manager"

  mkdir -p /tmp/forge4flow-manager-${version}-installation/hack
  cd /tmp/forge4flow-manager-${version}-installation
  $SUDO curl -fSLs "https://raw.githubusercontent.com/forge4flow/forge4flow-manager/${version}/docker-compose.yaml" --output "docker-compose.yaml"
  $SUDO curl -fSLs "https://raw.githubusercontent.com/forge4flow/forge4flow-manager/${version}/prometheus.yml" --output "prometheus.yml"
  $SUDO curl -fSLs "https://raw.githubusercontent.com/forge4flow/forge4flow-manager/${version}/resolv.conf" --output "resolv.conf"
  $SUDO curl -fSLs "https://raw.githubusercontent.com/forge4flow/forge4flow-manager/${version}/hack/forged-provider.service" --output "hack/forged-provider.service"
  $SUDO curl -fSLs "https://raw.githubusercontent.com/forge4flow/forge4flow-manager/${version}/hack/f4f-manager.service" --output "hack/f4f-manager.service"
  $SUDO /usr/local/bin/f4f-manager install
}

install_forge_cli() {
  echo "Installing forge-cli..."
  # Determine system architecture
  arch=$(uname -m)
  case $arch in
  x86_64 | amd64)
    suffix=""
    ;;
  aarch64)
    suffix=-arm64
    ;;
  armv7l)
    suffix=-armhf
    ;;
  *)
    echo "Unsupported architecture $arch"
    exit 1
    ;;
  esac

  # Download the appropriate binary
  $SUDO curl -fsSL https://github.com/forge4flow/forge-cli/releases/latest/download/forge-cli${suffix} --output /usr/local/bin/forge-cli
  $SUDO chmod +x /usr/local/bin/forge-cli
}


verify_system
install_required_packages

$SUDO /sbin/sysctl -w net.ipv4.conf.all.forwarding=1
echo "net.ipv4.conf.all.forwarding=1" | $SUDO tee -a /etc/sysctl.conf

install_arkade
install_cni_plugins
install_containerd
install_forge_cli
install_forge4flow-manager