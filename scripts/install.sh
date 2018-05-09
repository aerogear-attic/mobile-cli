#!/usr/bin/env bash

function usage {
  echo "usage: ${0} [version]"
  exit 0
}

function error_msg {
  echo "${1}. Exit code: ${2}"
  exit ${2}
}

function get_os {
  case "$OSTYPE" in
    darwin*)   
      os="darwin" ;;
    linux*)  
      os="linux" ;;
    *)
      echo "Unsupported OS: $OSTYPE" 
      exit 1 ;;
  esac
}

function get_arch {
  case "$(uname -m)" in 
    x86_64) 
      arch="amd64" ;;
    *)
      echo "Unsupported ARCH: $(uname -m)" 
      exit 1 ;;
  esac
}

function install {
  echo "Installing the Mobile CLI..."
  if [ -n "${1}" ]; then 
    readonly version=${1}
  else
    readonly version=$(curl -s "https://api.github.com/repos/aerogear/mobile-cli/releases/latest" | grep '"tag_name":' | sed 's/[^0-9.]*//g')
  fi

  echo "Downloading version ${version}"
  get_os
  get_arch

  curl -s -f -L -O -J https://github.com/aerogear/mobile-cli/releases/download/v${version}/mobile-cli_${version}_${os}_${arch}.tar.gz
  exit_code=${?}
  
  if [ $exit_code -ne 0 ]; then 
    error_msg "Failed to download binary" ${exit_code}
  fi

  tar -xf mobile-cli_${version}_${os}_${arch}.tar.gz && sudo cp mobile /usr/local/bin
  exit_code=${?}

  if [ $exit_code -ne 0 ]; then 
    error_msg "Failed to unpack binary" ${exit_code}
  fi

  rm -rf mobile-cli_${version}_${os}_${arch}.tar.gz && rm -rf mobile
  echo "Mobile CLI installed. Use 'mobile --help' for more information."
}

if [[ $* == *-h* ]] || [[ $* == *--help* ]]; then
  usage
fi

install ${1}