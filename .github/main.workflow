workflow "make" {
  on = "push"
  resolves = [
    "all",
    # "go-lint",
    # "go-test",
    # "go-build",
    # "docker-push",
  ]
}

action "all" {
  uses = "docker://golang:latest"
  runs = "make deps"
  env = {
    CI = "true"
    GOPATH = "/github/workspace/.go"
  }
}

action "deps" {
  uses = "docker://golang:latest"
  runs = "make deps"
  env = {
    GOPATH = "/github/workspace/.go"
  }
}

action "go-lint" {
  needs = [
    "deps",
  ]
  uses = "docker://golang:latest"
  runs = "make lint"
  env = {
    GOPATH = "/github/workspace/.go"
  }
}

action "go-test" {
  needs = [
    "deps",
  ]
  uses = "docker://golang:latest"
  runs = "make test"
  env = {
    GOPATH = "/github/workspace/.go"
  }
}

action "go-build" {
  needs = [
    "deps",
  ]
  uses = "docker://golang:latest"
  runs = "make build"
  env = {
    GOPATH = "/github/workspace/.go"
  }
}

# action "docker-build" {
#   needs = [
#     "deps",
#   ]
#   uses = "actions/docker/cli@master"
#   args = "build -t nimona/nimona-dev ."
#   env = {
#     GOPATH = "/github/workspace/.go"
#   }
# }

# action "docker-login" {
#   needs = [
#     "docker-build",
#   ]
#   uses = "actions/docker/login@master"
#   secrets = [
#     "DOCKER_USERNAME",
#     "DOCKER_PASSWORD",
#   ]
# }

# action "docker-push" {
#   needs = [
#     "docker-login",
#     "go-test",
#     "go-lint",
#     "go-build",
#   ]
#   uses = "actions/docker/cli@master"
#   args = "push nimona/nimona-dev"
# }
