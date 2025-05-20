Usage:

numParser-linux-amd64  [-p PORT] [--proxy http://user:password@ip:port]

default PORT 38888
default host http://rutor.info

Possible to change some setting with config file config.yml in the directory of numParser. Settings from the command line have priority under settings in the config.yml

##### Exemple of config.yml (possible to put settings in any combinations)
>
>host: http://6tor.org
>
>port: 29999
>
>useproxy: false
>
>proxy: http://user:password@ip:port
>
>tmdbtoken: 'Bearer [API Read Access Token]'


#####
Сборка 
```shell
go build -o NUMParser_deb ./cmd
```
