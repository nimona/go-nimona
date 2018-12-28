workflow "Lint, test, & build" {
  on = "push"
  resolves = ["Push Container"]
}

action "Install Dependencies" {
  uses = "./.github/actions/golang"
  args = ["deps"]
  env = {
    GOPATH = "/github/home"
  }
}

action "Run Linters" {
  needs = ["Install Dependencies"]
  uses = "./.github/actions/golang"
  args = ["tools-and-lint"]
  env = {
    GOPATH = "/github/home"
  }
}

action "Run Tests" {
  needs = ["Install Dependencies"]
  uses = "./.github/actions/golang"
  args = ["test"]
  env = {
    GOPATH = "/github/home"
  }
}

action "Build Binaries" {
  needs = ["Install Dependencies"]
  uses = "./.github/actions/golang"
  args = ["build"]
  env = {
    GOPATH = "/github/home"
  }
}

action "Build Container" {
  uses = "./.github/actions/docker"
  secrets = ["DOCKER_IMAGE"]
  args = ["build", "Dockerfile"]
}

action "Login Dockerhub" {
  needs = ["Build Container"]
  uses = "actions/docker/login@master"
  secrets = ["DOCKER_USERNAME", "DOCKER_PASSWORD"]
}

action "Push Container" {
  needs = ["Login Dockerhub", "Run Tests", "Run Linters", "Build Binaries"]
  uses = "./.github/actions/docker"
  secrets = ["DOCKER_IMAGE"]
  args = "push"
}
