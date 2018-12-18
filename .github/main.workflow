workflow "Lint, test, & build" {
  on = "push"
  resolves = ["Build"]
}

action "Deps" {
  uses = "./.github/actions/golang"
  args = ["deps"]
}

action "Tools" {
  needs = ["Deps"]
  uses = "./.github/actions/golang"
  args = ["tools"]
}

action "Lint" {
  needs = ["Tools"]
  uses = "./.github/actions/golang"
  args = ["lint"]
}

action "Test" {
  needs = ["Lint"]
  uses = "./.github/actions/golang"
  args = ["test"]
}

action "Build" {
  needs = ["Test"]
  uses = "./.github/actions/golang"
  args = ["build"]
}
