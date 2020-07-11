[node]
%{ for name in names ~}
${name} ansible_host=${ips[index(names, name)]} ansible_user=${node_user}
%{ endfor ~}
