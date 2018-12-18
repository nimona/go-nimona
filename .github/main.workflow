workflow "Lint, test, & build" {
  on = "push"
  resolves = ["Push"]
}

action "Tools" {
  uses = "./.github/actions/golang"
  args = ["make", "tools"]
}

action "Lint" {
  needs = ["Tools"]
  uses = "./.github/actions/golang"
  args = ["make", "lint"]
}

action "Test" {
  needs = ["Lint"]
  uses = "./.github/actions/golang"
  args = ["make", "test"]
}

action "Build" {
  needs = ["Test"]
  uses = "./.github/actions/golang"
  args = ["make", "build"]
}
