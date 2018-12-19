workflow "Lint, test, & build" {
  on = "push"
  resolves = ["Push Container"]
}

action "Run Linters" {
  uses = "./.github/actions/golang"
  args = ["lint"]
}

action "Run Tests" {
  uses = "./.github/actions/golang"
  args = ["test"]
}

action "Build Binaries" {
  uses = "./.github/actions/golang"
  args = ["build"]
}

action "Build Container" {
  needs = ["Run Tests", "Run Linters", "Build Binaries"]
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
  needs = ["Login Dockerhub"]
  uses = "./.github/actions/docker"
  secrets = ["DOCKER_IMAGE"]
  args = "push"
}
