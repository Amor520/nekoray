SRC_ROOT="$PWD"
DEPLOYMENT="$SRC_ROOT/deployment"
BUILD="$SRC_ROOT/build"
version_standalone="nekoray-"$(cat nekoray_version.txt) # 下次改
version_sing_box="$(git -C ../sing-box describe --tags --always 2>/dev/null || echo unknown)"
