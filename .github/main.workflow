workflow "Lint, test, & build" {
  on = "push"
  resolves = [
    # "go-lint",
    "go-test",
    "go-build",
    # "docker-push",
  ]
}

action "deps" {
  uses = "./.github/actions/golang"
  args = [
    "deps",
  ]
  env = {
    GOPATH = "/github/home"
  }
}

# action "go-lint" {
#   needs = [
#     "deps",
#   ]
#   uses = "./.github/actions/golang"
#   args = [
#     "lint",
#   ]
#   env = {
#     GOPATH = "/github/home"
#   }
# }

action "go-test" {
  needs = [
    "deps",
  ]
  uses = "./.github/actions/golang"
  args = [
    "test",
  ]
  env = {
    GOPATH = "/github/home"
  }
}

action "go-build" {
  needs = [
    "deps",
  ]
  uses = "./.github/actions/golang"
  args = [
    "build",
  ]
  env = {
    GOPATH = "/github/home"
  }
}

# action "docker-build" {
#   needs = [
#     "deps",
#   ]
#   uses = "actions/docker/cli@master"
#   args = "build -t nimona/nimona-dev ."
#   env = {
#     GOPATH = "/github/home"
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
