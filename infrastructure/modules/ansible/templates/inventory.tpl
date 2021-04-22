%{for group, instances in server_groups~}
[${group}]
%{for name, instance in instances~}
${instance.hostname} ansible_host=${instance.ip_address} ansible_user=${instance.user}
%{endfor~}
%{if contains(keys(instances), group)~}

[${group}_root]
${instances[group].hostname} ansible_host=${instances[group].ip_address} ansible_user=${instances[group].user}
%{endif~}

%{endfor~}
