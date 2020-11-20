---
prometheus_scrape_configs:
  - job_name: "prometheus"
    metrics_path: "/prometheus/metrics"
    static_configs:
      - targets:
          - "localhost:9090"
  - job_name: "caddy"
    metrics_path: "/metrics/caddy"
    scheme: "https"
    basic_auth:
      username: "{{ metrics_users.items() | first | first }}"
      password: "{{ metrics_users.items() | first | last }}"
    file_sd_configs:
      - files:
          - "file_sd/*.yml"
  - job_name: "node"
    metrics_path: "/metrics/node"
    scheme: "https"
    basic_auth:
      username: "{{ metrics_users.items() | first | first }}"
      password: "{{ metrics_users.items() | first | last }}"
    file_sd_configs:
      - files:
          - "file_sd/*.yml"
  - job_name: "cadvisor"
    metrics_path: "/metrics/cadvisor"
    scheme: "https"
    basic_auth:
      username: "{{ metrics_users.items() | first | first }}"
      password: "{{ metrics_users.items() | first | last }}"
    file_sd_configs:
      - files:
          - "file_sd/*.yml"
%{for job in prometheus_jobs~}
  - job_name: "${job.name}"
    metrics_path: "${lookup(job, "metrics_path", "/metrics")}"
    scheme: "${lookup(job, "scheme", "https")}"
    basic_auth:
      username: "{{ metrics_users.items() | first | first }}"
      password: "{{ metrics_users.items() | first | last }}"
    file_sd_configs:
      - files:
          - "file_sd/${job.group}.yml"
%{endfor~}

prometheus_targets:
%{for group, servers in servers_by_group~}
  ${group}:
%{for name, server in servers~}
    - targets:
        - "${server.hostname}"
      labels:
        server: "${name}"
        group: "${group}"
%{endfor~}
%{endfor~}
