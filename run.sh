#!/usr/bin/env bash

export PROJECT_ROOT="$(git rev-parse --show-toplevel)"

# export BUILD_MODE="PROD"
export BUILD_MODE="DEV"

# command to install webkit on fedora
# sudo dnf install webkit2gtk3-devel

export APP_PORT=6200
export SERVER_PORT=6201
export DEV_PORT=6202
export APP_TAGS="-tags withwebview"

# TODO
# export ADMIN_UI_PORT=6203

# export SERVER_URL="https://solid-succotash-gwjp9pr7r59265g-6201.app.github.dev"
export SERVER_URL="http://localhost:6201"
export VARS="-X main.build_mode=$BUILD_MODE -X main.port=$APP_PORT -X main.server_url=$SERVER_URL"

# for urita
# - [Can't find .so in the same directory as the executable?](https://serverfault.com/questions/279068/cant-find-so-in-the-same-directory-as-the-executable)
export CGO_LDFLAGS="-Wl,-rpath=\$ORIGIN"
export GOPROXY=direct # building on windows :/

if command -v bun >/dev/null; then
  runner="bun"
else
  runner="npm"
fi

web-build() {
  cd "$PROJECT_ROOT/frontend"

  # replaced at runtime
  SERVER_URL="%SERVER_URL%" APP_PORT="%APP_PORT%" $runner run build

  if [[ -d ../application/dist ]]; then
    rm -rf ../application/dist
  fi
  cp -r ./dist ../application/.
}

admin-web-build() {
  cd "$PROJECT_ROOT/admin"
  source ./.env

  # replaced at runtime
  SERVER_URL="%SERVER_URL%" $runner run build

  if [[ -d ../backend/dist ]]; then
    rm -rf ../backend/dist
  fi
  cp -r ./dist ../backend/.
}

# - [webview/webview](https://github.com/webview/webview?tab=readme-ov-file#windows)
#   - NOTE: install WebView2 runtime for < Windows 11
# - [MAYBE: WebView2Loader.dll](https://github.com/webview/webview?tab=readme-ov-file#ms-webview2-loader)
build-windows-app() {
  build-urita
  web-build
  
  cd "$PROJECT_ROOT/application"
  source ./.env
  export BUILD_MODE="PROD"
  # export SERVER_URL=""
  export VARS="-X main.build_mode=$BUILD_MODE -X main.port=$APP_PORT -X main.server_url=$SERVER_URL"
  export GOOS=windows
  export GOARCH=amd64
  export CGO_ENABLED=1

  echo "NOTE: building with SERVER_URL as $SERVER_URL"

  go build $APP_TAGS -ldflags "$VARS -H windowsgui" -o ../build/gravtest.exe ./src/.
}

build-windows-server() {
  admin-web-build

  cd "$PROJECT_ROOT/backend"
  source ./.env

  export BUILD_MODE="PROD"
  export VARS="-X main.build_mode=$BUILD_MODE"
  export GOOS=windows
  export GOARCH=amd64
  export CGO_ENABLED=1
  # export SERVER_URL=""

  echo "NOTE: building with SERVER_URL as $SERVER_URL"

  go build -ldflags "$VARS -H windowsgui" -o ../build/server.exe ./src/.
}

build-urita() {
  cd "$PROJECT_ROOT/urita"

  cargo build --release

  mkdir -p ../build

  if [[ -f ./target/release/liburita.so ]]; then
    cp ./target/release/liburita.so ../build/.
  fi
  if [[ -f ./target/release/urita.dll ]]; then
    cp ./target/release/urita.dll ../build/.
  fi
  if [[ -f ./target/x86_64-pc-windows-gnu/release/urita.dll ]]; then
    cp ./target/x86_64-pc-windows-gnu/release/urita.dll ../build/.
  fi
}

build-app() {
  build-urita
  web-build
  
  cd "$PROJECT_ROOT/application"
  source ./.env
  export BUILD_MODE="PROD"
  # export SERVER_URL=""
  export VARS="-X main.build_mode=$BUILD_MODE -X main.port=$APP_PORT -X main.server_url=$SERVER_URL"

  echo "NOTE: building with SERVER_URL as $SERVER_URL"

  go build $APP_TAGS -ldflags "$VARS" -o ../build/gravtest ./src/.
}

build-server() {
  admin-web-build

  cd "$PROJECT_ROOT/backend"
  source ./.env

  export BUILD_MODE="PROD"
  # export SERVER_URL=""

  echo "NOTE: building with SERVER_URL as $SERVER_URL"

  export VARS="-X main.build_mode=$BUILD_MODE"
  go build -ldflags "$VARS" -o ../build/server ./src/.
}

admin-web-dev() {
  cd "$PROJECT_ROOT/admin"
  source ./.env

  $runner run dev
}

server() {
  cd "$PROJECT_ROOT/backend"
  source ./.env

  mkdir -p ./dist
  touch ./dist/ignore

  export VARS="-X main.build_mode=$BUILD_MODE"
  go build -ldflags "$VARS" -o ../build/server ./src/.
  ../build/server $@
}

web-dev() {
  cd "$PROJECT_ROOT/frontend"

  $runner run dev
}

app() {
  cd "$PROJECT_ROOT/application"
  source ./.env

  mkdir -p ./dist
  touch ./dist/ignore

  export VARS="-X main.build_mode=$BUILD_MODE -X main.port=$APP_PORT -X main.server_url=$SERVER_URL"

  go build $APP_TAGS -ldflags "$VARS" -o ../build/gravtest ./src/.
  ../build/gravtest $@
}

setup() {
  cd "$PROJECT_ROOT/application"
  go mod tidy

  cd "$PROJECT_ROOT/backend"
  go mod tidy

  cd "$PROJECT_ROOT/admin"
  $runner i

  cd "$PROJECT_ROOT/frontend"
  $runner i
}

run() {
  set -e 
  # set -o pipefail

  command="$1"
  if [[ $# > 1 ]]; then
    shift
  fi

  case $command in
    "web-build")
      web-build
    ;;
    "admin-web-build")
      admin-web-build
    ;;
    "build-windows-app")
      build-windows-app
    ;;
    "build-windows-server")
      build-windows-server
    ;;
    "build-server")
      build-server
    ;;
    "build-urita")
      build-urita
    ;;
    "build-app")
      build-app
    ;;
    "setup")
      setup
    ;;
    "admin-web-dev")
      admin-web-dev $@
    ;;
    "server")
      server $@
    ;;
    "web-dev")
      web-dev
    ;;
    "app")
      app $@
    ;;
    *)
      echo "unknown command"
    ;;
  esac
}

run $@
