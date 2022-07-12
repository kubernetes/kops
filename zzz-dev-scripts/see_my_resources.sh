echo "--> DNS RECORDS :"
scw dns record list leila.sieben.fr
echo "\n--> LOAD - BALANCERS :"
scw lb lb list zone=fr-par-1
echo "\n--> SERVERS :"
scw instance server list zone=fr-par-1
echo "\n--> VOLUMES :"
scw instance volume list zone=fr-par-1

