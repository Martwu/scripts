#!/usr/bin/env bash

CONFIG_FILE=${HOME}/.denvrc

###### 默认配置
# 默认的设置项目的目录，让docker可以直接读
#DENV_VOLUME_DIR=${HOME}/.denv.dir
DENV_VOLUME_DIR=/Users/blackfat/Downloads/Store/AllStar/Removable/Projects
# 设置默认的镜像，就不自定义了
DENV_DOCKER_IMAGE=centos:centos7
######

function _get_env_name () {
    img=$1
    echo "${img/:/.}_$2"
}

function _build_env () {
    # 新创一个容器
    img=$1
    env_name=$(_get_env_name $1 $2)
    docker run -itd -v ${DENV_VOLUME_DIR}:/Projects --name ${env_name} ${img} /bin/cat
}

function _run_env () {
    # 启动已有的容器
    env_name=$(_get_env_name $1 $2)
    [[ $(docker start ${env_name} 2>&1) =~ "EOF" ]] && echo "no container: ${env_name}" && return 1
}

function login_env() {
    # 进入已有的容器
    env_name=$(_get_env_name $1 $2)
    echo "login ${env_name}"
    docker exec -w /Projects -it ${env_name} /bin/bash
}

function _container_status () {
    name=$1
    container_status=$(docker exec ${name} echo "good" 2>&1)
    # 如果存在，并且运行着，返回 0
    [[ "${container_status}" == "good" ]] && return 0
    # 如果存在，但是没运行，返回 1
    [[ "${container_status}" =~ "is not running" ]] && return 1
    # 不存在，返回 2
    [[ "${container_status}" =~ "No such container" ]] && return 2
    # 异常情况，返回 3
    return 3
}

function start_env () {
    env_name=$(_get_env_name $1 $2)
    echo "start ${env_name}"
    case "$(_container_status ${env_name}; echo $?)" in
        "0")
            return 0
        ;;
        "1")
            _run_env $1 $2
            return 0
        ;;
        "2")
            _build_env $1 $2
            return 0
        ;;
        "3")
            return 1
        ;;
    esac
}

function set_env () {
  echo 1
}


# 第一个参数
case "$1" in
    "start")
    start_env ${DENV_DOCKER_IMAGE} $2
    ;;
    "in")
    start_env ${DENV_DOCKER_IMAGE} $2
    login_env ${DENV_DOCKER_IMAGE} $2
    ;;
    *)
    echo "$(basename $0) start name"
    echo "$(basename $0) in name"
    exit 1
    ;;
esac

# usage="$(basename "$0") [-h] [-s/--start envname] [-i/--login envname] [-c/--config KEY=VALUE]
# 
# details:
#     tools for using docker container as runtime env easily.
# 
# where:
#     -h          show this help text
#     -c --config KEY=VALUE set some values with $(basename "$0")
#                     keys: volume - volume of container default is [ ~/.denv.dir ]
#     -s --start  ENVNAME startup a container
#     -i --login  ENVNAME login the container
# "

# seed=42
# while getopts ':hs:' option; do
#   case "$option" in
#     h) echo "$usage"
#        exit
#        ;;
#     s) seed=$OPTARG
#        ;;
#     :) printf "missing argument for -%s\n" "$OPTARG" >&2
#        echo "$usage" >&2
#        exit 1
#        ;;
#    \?) printf "illegal option: -%s\n" "$OPTARG" >&2
#        echo "$usage" >&2
#        exit 1
#        ;;
#   esac
# done
# shift $((OPTIND - 1))
