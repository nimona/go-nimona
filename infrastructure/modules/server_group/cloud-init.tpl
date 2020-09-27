#cloud-config

# users
users:
  - name: ${user}
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    ssh_authorized_keys:
      - "${ssh_public_key}"

# packages
package_update: true
package_upgrade: true

packages:
  - curl
  - git
  - sudo

# power state
power_state:
  delay: "+1"
  timeout: 30
  mode: reboot
  message: Reboot after system upgrade
  condition: True
